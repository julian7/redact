package files_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/julian7/redact/files"
)

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

func TestLoad(t *testing.T) {
	tt := []struct {
		name          string
		hasKey        bool
		expectedError string
	}{
		{"uninitialized", false, "open /git/repo/.git/test/key: file does not exist"},
		{"initialized", true, ""},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			k := genGitRepo(t)
			if tc.hasKey {
				if !writeKey(t, k) {
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
			k := genGitRepo(t)
			if tc.hasKey {
				if !writeKey(t, k) {
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
