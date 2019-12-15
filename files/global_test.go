package files_test

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/spf13/afero"
)

const (
	sampleCode = "0123456789abcdefghijklmnopqrstuv"
)

func genGitRepo() (*files.MasterKey, error) {
	k := &files.MasterKey{
		Fs: afero.NewMemMapFs(),
		RepoInfo: gitutil.GitRepoInfo{
			Common:   ".git",
			Toplevel: "/git/repo",
		},
		KeyDir: "/git/repo/.git/test",
		Cache:  map[string]string{},
	}

	err := k.Mkdir(k.KeyDir, 0700)
	if err != nil {
		return nil, errors.Wrapf(err, "creating key dir %s", k.KeyDir)
	}

	return k, nil
}

func writeKey(k *files.MasterKey) error {
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

func writeKX(k *files.MasterKey) error {
	kxdir := filepath.Join(k.RepoInfo.Toplevel, ".redact")
	if err := k.MkdirAll(kxdir, 0755); err != nil {
		return errors.Wrap(err, "creating exchange dir")
	}

	return writeFile(
		k,
		filepath.Join(kxdir, ".gitattributes"),
		0644,
		"* !filter !diff\n*.gpg binary\n",
	)
}

func writeFile(k *files.MasterKey, fname string, perms os.FileMode, contents string) error {
	of, err := k.OpenFile(fname, os.O_CREATE|os.O_WRONLY, perms)
	if err != nil {
		return errors.Wrapf(err, "creating %s file", fname)
	}

	if _, err := of.WriteString(contents); err != nil {
		return errors.Wrapf(err, "writing %s file", fname)
	}

	if err := of.Close(); err != nil {
		return errors.Wrapf(err, "closing %s file", fname)
	}

	return nil
}

func checkString(expected, received string) error {
	if received != expected {
		return errors.Errorf(
			`Unexpected result.
Expected: %q
Received: %q`,
			expected,
			received,
		)
	}

	return nil
}
