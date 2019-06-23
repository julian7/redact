package keys

import (
	"bytes"
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
	_, err = fd.WriteString("\000REDACT\000\000\000\000\000\000\000\000\001" +
		sampleAES + sampleHMAC,
	)
	if err != nil {
		t.Errorf("cannot write key file %s: %v", keyfile, err)
		return false
	}
	return true
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
			if bytes.Compare(k.Key.AES(), []byte(sampleAES)) != 0 {
				t.Errorf(`Wrong AES key\nExpected: "%s"\nReceived: "%s"`, sampleAES, k.Key.AES())
			}
		})
	}
}
