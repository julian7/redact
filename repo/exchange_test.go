package repo_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5/helper/chroot"
	"github.com/go-git/go-billy/v5/memfs"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/repo"
	"github.com/julian7/tester"
)

func genGitRepo() (*repo.Repo, error) {
	fs := memfs.New()

	secretKey, err := files.NewSecretKey(chroot.New(fs, ".git"))
	if err != nil {
		return nil, err
	}

	r := &repo.Repo{
		SecretKey: secretKey,
		Workdir:   fs,
	}

	return r, nil
}

func writeKey(r *repo.Repo) error {
	return writeFile(
		r,
		r.SecretKey.Keyfile(),
		0600,
		"key contents",
	)
}

func writeKX(r *repo.Repo) error {
	kxdir := "/.redact"
	if err := r.Workdir.MkdirAll(kxdir, 0755); err != nil {
		return fmt.Errorf("creating exchange dir: %w", err)
	}

	return writeFile(
		r,
		filepath.Join(kxdir, ".gitattributes"),
		0644,
		"* !filter !diff\n*.gpg binary\n",
	)
}

func writeFile(r *repo.Repo, fname string, perms os.FileMode, contents string) error {
	of, err := r.Workdir.OpenFile(fname, os.O_CREATE|os.O_WRONLY, perms)
	if err != nil {
		return fmt.Errorf("creating %s file: %w", fname, err)
	}

	if _, err := io.WriteString(of, contents); err != nil {
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
			".redact/6465616462656566646561646265656664656164",
			nil,
		},
		{
			"repo",
			true,
			".redact/6465616462656566646561646265656664656164",
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

			ret, err := mk.GetExchangeFilenameStubFor(fingerprint, nil)
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
	if err := checkString("stub.asc", repo.ExchangePubKeyFile("stub")); err != nil {
		t.Error(err)
	}
}

func TestExchangeSecretKeyFile(t *testing.T) {
	if err := checkString("stub.key", repo.ExchangeSecretKeyFile("stub")); err != nil {
		t.Error(err)
	}
}
