package files_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/repo"
)

const (
	sampleCode = "0123456789abcdefghijklmnopqrstuv"
)

func genGitRepo() (*repo.Repo, error) {
	fs := memfs.New()

	dot, err := fs.Chroot(".git")
	if err != nil {
		return nil, err
	}

	secretKey, err := files.NewSecretKey(dot)
	if err != nil {
		return nil, err
	}

	r := &repo.Repo{
		Workdir:                fs,
		SecretKey:              secretKey,
		StrictPermissionChecks: false,
	}

	return r, nil
}

func writeKey(r *repo.Repo) error {
	return writeFileTo(
		r.Workdir,
		filepath.Join(".git", r.Keyfile()),
		0600,
		"\000REDACT\000"+ // preamble
			"\000\000\000\000"+ // key type == 0
			"\000\000\000\001"+ // first key epoch
			sampleCode+sampleCode+sampleCode+ // sample key
			"\000\000\000\002"+ // second key epoch
			sampleCode+sampleCode+sampleCode, // sample key
	)
}

func writeFile(r *repo.Repo, fname string, perms os.FileMode, contents string) error {
	return writeFileTo(r.Workdir, fname, perms, contents)
}

func writeFileTo(to billy.Filesystem, fname string, perms os.FileMode, contents string) error {
	of, err := to.OpenFile(fname, os.O_CREATE|os.O_WRONLY, perms)
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
