package files_test

import (
	"testing"

	"github.com/julian7/redact/files"
	"github.com/julian7/tester"
)

func TestGetExchangeFilenameStubFor(t *testing.T) {
	tt := []struct {
		name     string
		preload  bool
		expected string
		expErr   error
	}{
		{
			"empty",
			false,
			"/git/repo/.redact/6465616462656566646561646265656664656164",
			nil,
		},
		{
			"repo",
			true,
			"/git/repo/.redact/6465616462656566646561646265656664656164",
			nil,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fingerprint := []byte("deadbeefdeadbeefdead")
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
			if err2 := tester.AssertError(tc.expErr, err); err2 != nil {
				t.Error(err2)
			}
			if err == nil {
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

func TestExchangeSecretKeyFile(t *testing.T) {
	if err := checkString("stub.key", files.ExchangeSecretKeyFile("stub")); err != nil {
		t.Error(err)
	}
}
