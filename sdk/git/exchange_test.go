package git_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/sdk/git"
	"github.com/julian7/tester"
	"github.com/spf13/afero"
)

func genGitRepo() (*git.Repo, error) {
	r := &git.Repo{
		Fs:       afero.NewMemMapFs(),
		Logger:   logger.New(),
		Common:   ".git",
		Toplevel: "/git/repo",
	}

	err := r.Mkdir(r.Keydir(), 0700)
	if err != nil {
		return nil, fmt.Errorf("creating key dir %s: %w", r.Keydir(), err)
	}

	return r, nil
}

func writeKey(r *git.Repo) error {
	return writeFile(
		r,
		filepath.Join(r.Keydir(), "key"),
		0600,
		"key contents",
	)
}

func writeKX(r *git.Repo) error {
	kxdir := filepath.Join(r.Toplevel, ".redact")
	if err := r.MkdirAll(kxdir, 0755); err != nil {
		return fmt.Errorf("creating exchange dir: %w", err)
	}

	return writeFile(
		r,
		filepath.Join(kxdir, ".gitattributes"),
		0644,
		"* !filter !diff\n*.gpg binary\n",
	)
}

func writeFile(r *git.Repo, fname string, perms os.FileMode, contents string) error {
	of, err := r.OpenFile(fname, os.O_CREATE|os.O_WRONLY, perms)
	if err != nil {
		return fmt.Errorf("creating %s file: %w", fname, err)
	}

	if _, err := of.WriteString(contents); err != nil {
		return fmt.Errorf("writing %s file: %w", fname, err)
	}

	if err := of.Close(); err != nil {
		return fmt.Errorf("closing %s file: %w", fname, err)
	}

	return nil
}

func checkString(expected, received string) error {
	if received != expected {
		return fmt.Errorf(
			`Unexpected result.
Expected: %q
Received: %q`,
			expected,
			received,
		)
	}

	return nil
}

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
	if err := checkString("stub.asc", git.ExchangePubKeyFile("stub")); err != nil {
		t.Error(err)
	}
}

func TestExchangeSecretKeyFile(t *testing.T) {
	if err := checkString("stub.key", git.ExchangeSecretKeyFile("stub")); err != nil {
		t.Error(err)
	}
}
