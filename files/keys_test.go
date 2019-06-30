package files

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
)

const (
	sampleAES  = "0123456789abcdefghijklmnopqrstuv"
	sampleHMAC = "0123456789abcdefghijklmnopqrstuv0123456789abcdefghijklmnopqrstuv"
)

func genGitRepo(t *testing.T) *MasterKey {
	k := &MasterKey{Fs: afero.NewMemMapFs(), KeyDir: "/.git/test"}
	err := k.Fs.Mkdir(k.KeyDir, 0700)
	if err != nil {
		t.Errorf("cannot create key dir %s: %v", k.KeyDir, err)
		return nil
	}
	return k
}

func prebuild(t *testing.T, k *MasterKey) bool {
	keyfile := buildKeyFileName(k.KeyDir)
	fd, err := k.Fs.OpenFile(keyfile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Errorf("cannot open key file %s for writing: %v", keyfile, err)
		return false
	}
	defer fd.Close()
	_, err = fd.WriteString(
		"\000REDACT\000" + // preamble
			"\000\000\000\000" + // key type == 0
			"\000\000\000\001" + // first key epoch
			sampleAES + sampleHMAC + // sample key
			"\000\000\000\002" + // second key epoch
			sampleAES + sampleHMAC, // sample key
	)
	if err != nil {
		t.Errorf("cannot write key file %s: %v", keyfile, err)
		return false
	}
	if err != nil {
		t.Errorf("cannot write key file %s: %v", keyfile, err)
		return false
	}
	return true
}

func TestGenerate(t *testing.T) {
	k := genGitRepo(t)
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
		keyfunc func(KeyHandler) []byte
	}{
		{"AES", (KeyHandler).AES},
		{"HMAC", (KeyHandler).HMAC},
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

func TestLoad(t *testing.T) {
	tt := []struct {
		name          string
		hasKey        bool
		expectedError string
	}{
		{"uninitialized", false, "open /.git/test/key: file does not exist"},
		{"initialized", true, ""},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			k := genGitRepo(t)
			if tc.hasKey {
				if !prebuild(t, k) {
					return
				}
			}
			err := k.Load()
			if !checkError(t, tc.expectedError, err) || err != nil {
				return
			}
			key, err := k.Key(0)
			if err != nil {
				t.Errorf("Cannot retrieve latest key: %v", err)
				return
			}
			received := key.AES()
			if bytes.Compare(received, []byte(sampleAES)) != 0 {
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
			k := genGitRepo(t)
			if tc.hasKey {
				if !prebuild(t, k) {
					return
				}
			}
			err := k.Generate()
			if err != nil {
				t.Errorf("Error generating keys: %v", err.Error())
				return
			}
			err = k.Save()
			if !checkError(t, tc.expectedError, err) {
				return
			}
			finfo, err := k.Fs.Stat("/.git/test/key")
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
