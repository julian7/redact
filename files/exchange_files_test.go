package files_test

import (
	"testing"

	"github.com/julian7/redact/files"
)

func TestGetExchangeFilenameStubFor(t *testing.T) {
	tt := []struct {
		name     string
		preload  bool
		expected string
		expErr   string
	}{
		{"empty", false, "", "writing .gitattributes file in key exchange dir: open /git/repo/.redact/.gitattributes: no such file or directory"},
		{"repo", true, "/git/repo/.redact/6465616462656566646561646265656664656164", ""},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var fingerprint [20]byte
			copy(fingerprint[:], []byte("deadbeefdeadbeefdead"))
			mk, err := genGitRepo()
			if err != nil {
				t.Error(err)
				return
			}
			if err := writeKey(mk); err != nil {
				t.Error(err)
				return
			}
			if tc.preload {
				if err := writeKX(mk); err != nil {
					t.Error(err)
					return
				}
			}
			ret, err := mk.GetExchangeFilenameStubFor(fingerprint)
			if err != nil {
				if err2 := checkError(tc.expErr, err); err2 != nil {
					t.Error(err2)
				}
			} else {
				if err := checkString(tc.expected, ret); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestExchangePubKeyFile(t *testing.T) {
	if err := checkString("stub.asc", files.ExchangePubKeyFile("stub")); err != nil {
		t.Error(err)
	}
}

func TestExchangeMasterKeyFile(t *testing.T) {
	if err := checkString("stub.key", files.ExchangeMasterKeyFile("stub")); err != nil {
		t.Error(err)
	}
}
