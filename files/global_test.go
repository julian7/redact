package files_test

import (
	"fmt"
	"os"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/sdk/git"
	"github.com/spf13/afero"
)

const (
	sampleCode = "0123456789abcdefghijklmnopqrstuv"
)

func genGitRepo() (*files.SecretKey, error) {
	k := &files.SecretKey{
		Repo: &git.Repo{
			Fs:       afero.NewMemMapFs(),
			Logger:   logger.New(),
			Common:   ".git",
			Toplevel: "/git/repo",
		},
		Cache: map[string]string{},
	}

	err := k.Mkdir(k.Repo.Keydir(), 0700)
	if err != nil {
		return nil, fmt.Errorf("creating key dir %s: %w", k.Repo.Keydir(), err)
	}

	return k, nil
}

func writeKey(k *files.SecretKey) error {
	return writeFile(
		k,
		k.Repo.Keyfile(),
		0600,
		"\000REDACT\000"+ // preamble
			"\000\000\000\000"+ // key type == 0
			"\000\000\000\001"+ // first key epoch
			sampleCode+sampleCode+sampleCode+ // sample key
			"\000\000\000\002"+ // second key epoch
			sampleCode+sampleCode+sampleCode, // sample key
	)
}

func writeFile(k *files.SecretKey, fname string, perms os.FileMode, contents string) error {
	of, err := k.OpenFile(fname, os.O_CREATE|os.O_WRONLY, perms)
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
