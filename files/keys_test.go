package files_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/go-git/go-billy/v5"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/repo"
	"github.com/julian7/tester"
	"github.com/julian7/tester/ioprobe"
)

type key struct {
	epoch uint32
	key   []byte
}

func genKey(keyType uint32, keys []key) []byte {
	out := bytes.NewBuffer(nil)
	_, _ = out.WriteString("\x00REDACT\x00")
	_ = binary.Write(out, binary.BigEndian, keyType)

	for _, aKey := range keys {
		_ = binary.Write(out, binary.BigEndian, aKey.epoch)
		_, _ = out.Write(aKey.key[:96])
	}

	return out.Bytes()
}

func TestRead(t *testing.T) { //nolint:funlen
	tt := []struct {
		name   string
		reader io.Reader
		err    error
	}{
		{
			"read error",
			ioprobe.NewFailingReader(),
			errors.New("reading preamble from key file: unexpected EOF"),
		},
		{
			"invalid preamble",
			bytes.NewReader([]byte(samplePlaintext)),
			errors.New("invalid key file preamble"),
		},
		{
			"read error 2",
			ioprobe.NewTimeoutReader(
				bytes.NewReader(genKey(0, []key{{1, []byte(sampleCode + sampleCode + sampleCode)}})),
				2,
			),
			errors.New("reading key type: unexpected EOF"),
		},
		{
			"invalid key type",
			bytes.NewReader(genKey(5, []key{{1, []byte(sampleCode + sampleCode + sampleCode)}})),
			errors.New("invalid key type"),
		},
		{
			"read error 3",
			ioprobe.NewTimeoutReader(
				bytes.NewReader(genKey(0, []key{{1, []byte(sampleCode + sampleCode + sampleCode)}})),
				3,
			),
			errors.New("reading key data: unexpected EOF"),
		},
		{
			"duplicate epoch",
			bytes.NewReader(genKey(
				0,
				[]key{
					{1, []byte(sampleCode + sampleCode + sampleCode)},
					{1, []byte(sampleCode + sampleCode + sampleCode)},
				},
			)),
			errors.New("invalid key: duplicate epoch number (1)"),
		},
		{
			"success",
			bytes.NewReader(genKey(
				0,
				[]key{
					{1, []byte(sampleCode + sampleCode + sampleCode)},
				},
			)),
			nil,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			k := &files.SecretKey{}
			if err := tester.AssertError(tc.err, k.Read(tc.reader)); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	k, err := genGitRepo()
	if err != nil {
		t.Error(err)

		return
	}

	if k == nil {
		t.Error("cannot generate git repo")

		return
	}

	for _, name := range []string{"first", "second"} {
		err := k.Generate()
		if err != nil {
			t.Errorf("Error generating %s keypair: %v", name, err.Error())

			return
		}
	}

	for idx, name := range []string{"latest", "first", "second"} {
		idx, name := idx, name
		t.Run(fmt.Sprintf("%s key", name), func(t *testing.T) {
			key, err := k.Key(uint32(idx)) //nolint:gosec
			if err != nil {
				t.Errorf("cannot retrieve %s key: %v", name, err)

				return
			}

			val := key.Secret()
			if len(val) == 0 {
				t.Errorf("empty %s secret key", name)
			}

			nonzeros := 0

			for _, c := range val {
				if c > 0 {
					nonzeros++
				}
			}

			if nonzeros == 0 {
				t.Errorf("%s secret key is just zero bytes", name)
			}
		})
	}
}

type noOpenFile struct {
	billy.Filesystem
}

func (fs *noOpenFile) OpenFile(string, int, os.FileMode) (billy.File, error) {
	return nil, errors.New("open file returns error")
}

// type noMkdir struct {
// 	billy.Filesystem
// }

// func (fs *noMkdir) Mkdir(string, os.FileMode) error {
// 	return errors.New("Mkdir returns error")
// }

func TestLoad(t *testing.T) { //nolint:funlen,gocognit
	tt := []struct {
		name          string
		hasKey        bool
		fsMods        func(*repo.Repo)
		expectedError error
	}{
		{
			"uninitialized",
			false,
			func(*repo.Repo) {},
			errors.New("keydir \"redact\" not available: file does not exist"),
		},
		{
			"initialized",
			true,
			func(*repo.Repo) {},
			nil,
		},
		{
			"no key dir",
			true,
			func(k *repo.Repo) {
				err := k.Remove()
				if err != nil {
					panic(err)
				}
			},
			errors.New("key file \"redact/key\": file does not exist"),
		},
		{
			"excessive rights",
			true,
			func(k *repo.Repo) {
				if fs, ok := k.Workdir.(billy.Change); ok {
					_ = fs.Chmod(".git/test/key", 0777)
				}
			},
			nil,
		},
		{
			"restrictive rights",
			true,
			func(k *repo.Repo) {
				if fs, ok := k.Workdir.(billy.Change); ok {
					_ = fs.Chmod(".git/test/key", 0)
				}
			},
			nil,
		},
		// ### TMP FIX: cannot rewire dotgit FS after go-git repo creation
		// {
		// 	"read error",
		// 	true,
		// 	func(k *repo.Repo) {
		// 		k.Workdir = &noOpenFile{Filesystem: k.Workdir}
		// 	},
		// 	errors.New("opening key file for reading: open file returns error"),
		// },
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			k, err := genGitRepo()
			if err != nil {
				t.Error(err)

				return
			}

			if tc.hasKey {
				if err := writeKey(k); err != nil {
					t.Error(err)

					return
				}
			}

			tc.fsMods(k)
			err = k.Load(false)

			if err2 := tester.AssertError(tc.expectedError, err); err2 != nil {
				t.Error(err2)

				return
			}

			if err != nil {
				return
			}

			key, err := k.Key(0)
			if err != nil {
				t.Errorf("Cannot retrieve latest key: %v", err)

				return
			}

			received := key.Secret()

			if !bytes.Equal(received, []byte(sampleCode+sampleCode+sampleCode)) {
				t.Errorf(`Wrong AES key\nExpected: "%s"\nReceived: "%s"`, sampleCode+sampleCode+sampleCode, received)
			}
		})
	}
}

func TestSaveTo(t *testing.T) { //nolint:funlen
	tt := []struct {
		name   string
		writer io.Writer
		keys   int
		len    int
		err    error
	}{
		{
			name:   "success",
			writer: bytes.NewBuffer(nil),
			keys:   1,
			len:    112,
			err:    nil,
		},
		{
			name:   "no keys",
			writer: bytes.NewBuffer(nil),
			keys:   0,
			len:    12,
			err:    nil,
		},
		{
			name:   "5 keys",
			writer: bytes.NewBuffer(nil),
			keys:   5,
			len:    512,
			err:    nil,
		},
		{
			name:   "write error 1",
			writer: ioprobe.NewTimeoutWriter(bytes.NewBuffer(nil), 1),
			keys:   1,
			err:    errors.New("writing key preamble: unexpected EOF"),
		},
		{
			name:   "write error 2",
			writer: ioprobe.NewTimeoutWriter(bytes.NewBuffer(nil), 2),
			keys:   1,
			len:    -1,
			err:    errors.New("writing key type header: unexpected EOF"),
		},
		{
			name:   "write error 3",
			writer: ioprobe.NewTimeoutWriter(bytes.NewBuffer(nil), 3),
			keys:   1,
			len:    -1,
			err:    errors.New("writing key contents: key #1: unexpected EOF"),
		},
		{
			name:   "write error 5",
			writer: ioprobe.NewTimeoutWriter(bytes.NewBuffer(nil), 5),
			keys:   3,
			len:    -1,
			err:    errors.New("writing key contents: key #3: unexpected EOF"),
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			k := files.SecretKey{}
			for i := 0; i < tc.keys; i++ {
				err := k.Generate()
				if err != nil {
					t.Errorf("cannot generate key: %v", err)

					return
				}
			}

			received := k.SaveTo(tc.writer)

			if err := tester.AssertError(tc.err, received); err != nil {
				t.Error(err)
			}

			if received != nil || tc.len < 0 {
				return
			}

			buf, ok := tc.writer.(*bytes.Buffer)
			if !ok {
				t.Errorf("invalid type (%T) for writer success check", tc.writer)
			}

			recvlen := len(buf.Bytes())

			if recvlen != tc.len {
				t.Errorf("invalid key length. Received: %d, expected: %d", recvlen, tc.len)
			}
		})
	}
}

func TestSave(t *testing.T) { //nolint:funlen,gocognit
	tt := []struct {
		name    string
		saveKey bool
		fsMods  func(*repo.Repo)
		keyErr  error
		saveErr error
	}{
		{
			"uninitialized",
			false,
			func(*repo.Repo) {},
			nil,
			nil,
		},
		{
			"initialized",
			true,
			func(*repo.Repo) {},
			nil,
			nil,
		},
		// billy memfs creates parent dir on demand, and doesn't fail if parent dir doesn't exist
		// {
		// 	"error getting keydir",
		// 	true,
		// 	func(k *repo.Repo) {
		// 		_ = util.RemoveAll(k.Filesystem, k.Repo.Keydir())
		// 		k.Filesystem = &noMkdir{Filesystem: k.Filesystem}
		// 	},
		// 	errors.New("creating keydir: Mkdir returns error"),
		//  nil,
		// },
		{
			"error writing key",
			true,
			func(k *repo.Repo) {
				k.Workdir = &noOpenFile{Filesystem: k.Workdir}
			},
			errors.New("saving key file: open file returns error"),
			nil,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			k, err := genGitRepo()
			if err != nil {
				t.Error(err)

				return
			}

			if err := k.Generate(); err != nil {
				t.Errorf("Error generating keys: %v", err.Error())

				return
			}

			if tc.saveKey {
				if err := writeKey(k); err != nil {
					if err2 := tester.AssertError(tc.keyErr, err); err2 != nil {
						t.Error(err2)
					}

					return
				}
			}

			tc.fsMods(k)

			err = k.Save()
			if err2 := tester.AssertError(tc.saveErr, err); err2 != nil {
				t.Error(err2)

				return
			}

			if err != nil {
				return
			}

			finfo, err := k.Workdir.Stat(".git/redact/key")

			if err != nil {
				t.Errorf("key not created: %v", err)

				return
			}

			if !finfo.Mode().IsRegular() {
				t.Errorf("key is not regular file")
			}
		})
	}
}
