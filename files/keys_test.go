package files_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/pkg/errors"
)

func TestNewMasterKey(t *testing.T) {
	oldgitdir := files.GitDirFunc
	files.GitDirFunc = func(info *gitutil.GitRepoInfo) error { return errors.New("no git dir") }
	_, err := files.NewMasterKey()
	files.GitDirFunc = oldgitdir
	if err == nil || err.Error() != "not a git repository" {
		t.Errorf("Unexpected error: %v", err)
	}
	files.GitDirFunc = func(info *gitutil.GitRepoInfo) error {
		info.Common = ".git"
		info.Toplevel = "/git/repo"
		return nil
	}
	k, err := files.NewMasterKey()
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
	aes   []byte
	hmac  []byte
}

func genKey(keyType uint32, keys []key) []byte {
	out := bytes.NewBuffer(nil)
	_, _ = out.WriteString("\x00REDACT\x00")
	_ = binary.Write(out, binary.BigEndian, uint32(keyType))
	for _, aKey := range keys {
		_ = binary.Write(out, binary.BigEndian, aKey.epoch)
		_, _ = out.Write(aKey.aes[:32])
		_, _ = out.Write(aKey.hmac[:64])
	}
	return out.Bytes()
}

func TestRead(t *testing.T) {
	tt := []struct {
		name   string
		reader io.Reader
		err    string
	}{
		{"read error", &failingReader{}, "reading preamble from key file: unexpected EOF"},
		{"invalid preamble", bytes.NewReader([]byte(samplePlaintext)), "invalid key file preamble"},
		{
			"read error 2",
			TimeoutReader(
				bytes.NewReader(genKey(0, []key{{1, []byte(sampleAES), []byte(sampleHMAC)}})),
				2,
			),
			"reading key type: unexpected EOF",
		},
		{
			"invalid key type",
			bytes.NewReader(genKey(5, []key{{1, []byte(sampleAES), []byte(sampleHMAC)}})),
			"invalid key type",
		},
		{
			"read error 3",
			TimeoutReader(
				bytes.NewReader(genKey(0, []key{{1, []byte(sampleAES), []byte(sampleHMAC)}})),
				3,
			),
			"reading key data: unexpected EOF",
		},
		{
			"duplicate epoch",
			bytes.NewReader(genKey(
				0,
				[]key{
					{1, []byte(sampleAES), []byte(sampleHMAC)},
					{1, []byte(sampleAES), []byte(sampleHMAC)},
				},
			)),
			"invalid key: duplicate epoch number (1)",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			k := &files.MasterKey{}
			if err := checkError(tc.err, k.Read(tc.reader)); err != nil {
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
	tt := []struct {
		name    string
		keyfunc func(files.KeyHandler) []byte
	}{
		{"AES", (files.KeyHandler).AES},
		{"HMAC", (files.KeyHandler).HMAC},
	}
	for idx, name := range []string{"latest", "first", "second"} {
		t.Run(fmt.Sprintf("%s key", name), func(t *testing.T) {
			key, err := k.Key(uint32(idx))
			if err != nil {
				t.Errorf("cannot retrieve %s key: %v", name, err)
				return
			}
			for _, tc := range tt {
				t.Run(tc.name, func(t *testing.T) {
					val := tc.keyfunc(key)
					if len(val) == 0 {
						t.Errorf("empty %s %s key", name, tc.name)
					}
					nonzeros := 0
					for _, c := range val {
						if c > 0 {
							nonzeros++
						}
					}
					if nonzeros == 0 {
						t.Errorf("%s %s key is just zero bytes", name, tc.name)
					}
				})
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

type noWrite struct {
	afero.Fs
}

func (fs *noWrite) Write([]byte) (int, error) {
	return 0, errors.New("write returns error")
}

func TestLoad(t *testing.T) {
	tt := []struct {
		name          string
		hasKey        bool
		fsMods        func(*files.MasterKey)
		expectedError string
	}{
		{"uninitialized", false, func(*files.MasterKey) {}, "open /git/repo/.git/test/key: file does not exist"},
		{"initialized", true, func(*files.MasterKey) {}, ""},
		{
			"no key dir",
			true,
			func(k *files.MasterKey) { k.KeyDir = "/a/b/c" },
			"keydir not available: open /a/b/c: file does not exist",
		},
		{
			"excessive rights",
			true,
			func(k *files.MasterKey) {
				_ = k.Chmod("/git/repo/.git/test/key", 0777)
			},
			"excessive rights on key file",
		},
		{
			"restrictive rights",
			true,
			func(k *files.MasterKey) {
				_ = k.Chmod("/git/repo/.git/test/key", 0)
			},
			"insufficient rights on key file",
		},
		{
			"read error",
			true,
			func(k *files.MasterKey) {
				k.Fs = &noOpenFile{Fs: k.Fs}
			},
			"opening key file for reading: open file returns error",
		},
	}
	for _, tc := range tt {
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
			err = k.Load()
			if err2 := checkError(tc.expectedError, err); err2 != nil {
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
			received := key.AES()
			if !bytes.Equal(received, []byte(sampleAES)) {
				t.Errorf(`Wrong AES key\nExpected: "%s"\nReceived: "%s"`, sampleAES, received)
			}
		})
	}
}

func TestSave(t *testing.T) {
	tt := []struct {
		name          string
		hasKey        bool
		expectedError string
	}{
		{"uninitialized", false, ""},
		{"initialized", false, ""},
	}
	for _, tc := range tt {
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
			err = k.Generate()
			if err != nil {
				t.Errorf("Error generating keys: %v", err.Error())
				return
			}
			err = k.Save()
			if err2 := checkError(tc.expectedError, err); err2 != nil {
				t.Error(err2)
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
