package files_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/logger"
	"github.com/julian7/tester"
	"github.com/julian7/tester/ioprobe"
)

func TestNewMasterKey(t *testing.T) {
	oldgitdir := files.GitDirFunc
	files.GitDirFunc = func(info *gitutil.GitRepoInfo) error { return errors.New("no git dir") }
	_, err := files.NewMasterKey(logger.New())
	files.GitDirFunc = oldgitdir

	if err == nil || err.Error() != "not a git repository" {
		t.Errorf("Unexpected error: %v", err)
	}

	files.GitDirFunc = func(info *gitutil.GitRepoInfo) error {
		info.Common = ".git"
		info.Toplevel = "/git/repo"

		return nil
	}
	k, err := files.NewMasterKey(logger.New())
	files.GitDirFunc = oldgitdir

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if _, ok := k.Fs.(afero.Fs); !ok {
		t.Errorf("Unexpected filesystem type: %T", k.Fs)
	}

	if k.KeyDir != ".git/redact" {
		t.Errorf("invalid keydir: %s", k.KeyDir)
	}
}

func TestKeyFile(t *testing.T) {
	k := files.MasterKey{KeyDir: ".git/redact"}
	keyfile := k.KeyFile()

	if keyfile != ".git/redact/key" {
		t.Errorf("invalid keyfile: %s", keyfile)
	}
}

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
			k := &files.MasterKey{}
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
			key, err := k.Key(uint32(idx))
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
	afero.Fs
}

func (fs *noOpenFile) OpenFile(string, int, os.FileMode) (afero.File, error) {
	return nil, errors.New("open file returns error")
}

type noMkdir struct {
	afero.Fs
}

func (fs *noMkdir) Mkdir(string, os.FileMode) error {
	return errors.New("Mkdir returns error")
}

func TestLoad(t *testing.T) { //nolint:funlen
	tt := []struct {
		name          string
		hasKey        bool
		fsMods        func(*files.MasterKey)
		expectedError error
	}{
		{
			"uninitialized",
			false,
			func(*files.MasterKey) {},
			errors.New("open /git/repo/.git/test/key: file does not exist"),
		},
		{
			"initialized",
			true,
			func(*files.MasterKey) {},
			nil,
		},
		{
			"no key dir",
			true,
			func(k *files.MasterKey) { k.KeyDir = "/a/b/c" },
			errors.New("keydir not available: open /a/b/c: file does not exist"),
		},
		{
			"excessive rights",
			true,
			func(k *files.MasterKey) {
				_ = k.Chmod("/git/repo/.git/test/key", 0777)
			},
			nil,
		},
		{
			"restrictive rights",
			true,
			func(k *files.MasterKey) {
				_ = k.Chmod("/git/repo/.git/test/key", 0)
			},
			nil,
		},
		{
			"read error",
			true,
			func(k *files.MasterKey) {
				k.Fs = &noOpenFile{Fs: k.Fs}
			},
			errors.New("opening key file for reading: open file returns error"),
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
			if tc.hasKey {
				if err := writeKey(k); err != nil {
					t.Error(err)
					return
				}
			}
			tc.fsMods(k)
			err = k.Load(true)
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
			k := files.MasterKey{}
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

func TestSave(t *testing.T) { //nolint:funlen
	tt := []struct {
		name   string
		hasKey bool
		fsMods func(*files.MasterKey)
		err    error
	}{
		{
			"uninitialized",
			false,
			func(*files.MasterKey) {},
			nil,
		},
		{
			"initialized",
			true,
			func(*files.MasterKey) {},
			nil,
		},
		{
			"error getting keydir",
			false,
			func(k *files.MasterKey) {
				_ = k.RemoveAll(k.KeyDir)
				k.Fs = &noMkdir{Fs: k.Fs}
			},
			errors.New("creating keydir: Mkdir returns error"),
		},
		{
			"error writing key",
			false,
			func(k *files.MasterKey) {
				k.Fs = &noOpenFile{Fs: k.Fs}
			},
			errors.New("saving key file: open file returns error"),
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
			err = k.Generate()
			if err != nil {
				t.Errorf("Error generating keys: %v", err.Error())
				return
			}
			if tc.hasKey {
				if err := writeKey(k); err != nil {
					t.Error(err)
					return
				}
			}

			tc.fsMods(k)
			err = k.Save()
			if err2 := tester.AssertError(tc.err, err); err2 != nil {
				t.Error(err2)
				return
			}

			if err != nil {
				return
			}
			finfo, err := k.Stat("/git/repo/.git/test/key")
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
