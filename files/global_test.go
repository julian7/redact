package files_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/logger"
	"github.com/spf13/afero"
)

const (
	sampleCode = "0123456789abcdefghijklmnopqrstuv"
)

func genGitRepo() (*files.SecretKey, error) {
	k := &files.SecretKey{
		Fs:     afero.NewMemMapFs(),
		Logger: logger.New(),
		RepoInfo: gitutil.GitRepoInfo{
			Common:   ".git",
			Toplevel: "/git/repo",
		},
		KeyDir: "/git/repo/.git/test",
		Cache:  map[string]string{},
	}

	err := k.Mkdir(k.KeyDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("creating key dir %s: %w", k.KeyDir, err)
	}

	return k, nil
}

func writeKey(k *files.SecretKey) error {
	return writeFile(
		k,
		filepath.Join(k.KeyDir, "key"),
		0600,
		"\000REDACT\000"+ // preamble
			"\000\000\000\000"+ // key type == 0
			"\000\000\000\001"+ // first key epoch
			sampleCode+sampleCode+sampleCode+ // sample key
			"\000\000\000\002"+ // second key epoch
			sampleCode+sampleCode+sampleCode, // sample key
	)
}

func writeKX(k *files.SecretKey) error {
	kxdir := filepath.Join(k.RepoInfo.Toplevel, ".redact")
	if err := k.MkdirAll(kxdir, 0755); err != nil {
		return fmt.Errorf("creating exchange dir: %w", err)
	}

	return writeFile(
		k,
		filepath.Join(kxdir, ".gitattributes"),
		0644,
		"* !filter !diff\n*.gpg binary\n",
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
