package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/spf13/afero"
)

const (
	sampleAES  = "0123456789abcdefghijklmnopqrstuv"
	sampleHMAC = "0123456789abcdefghijklmnopqrstuv0123456789abcdefghijklmnopqrstuv"
)

func genGitRepo(t *testing.T) *files.MasterKey {
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
		t.Errorf("cannot create key dir %s: %v", k.KeyDir, err)
		return nil
	}
	return k
}

func writeKey(t *testing.T, k *files.MasterKey) bool {
	return writeFile(
		t,
		k,
		filepath.Join(k.KeyDir, "key"),
		0600,
		"\000REDACT\000"+ // preamble
			"\000\000\000\000"+ // key type == 0
			"\000\000\000\001"+ // first key epoch
			sampleAES+sampleHMAC+ // sample key
			"\000\000\000\002"+ // second key epoch
			sampleAES+sampleHMAC, // sample key
	)
}

func writeKX(t *testing.T, k *files.MasterKey) bool {
	kxdir := filepath.Join(k.RepoInfo.Toplevel, ".redact")
	if err := k.MkdirAll(kxdir, 0755); err != nil {
		t.Errorf("cannot create exchange dir: %v", err)
		return false
	}
	return writeFile(
		t,
		k,
		filepath.Join(kxdir, ".gitattributes"),
		0644,
		"* !filter !diff\n*.gpg binary\n",
	)
}

func writeFile(t *testing.T, k *files.MasterKey, fname string, perms os.FileMode, contents string) bool {
	of, err := k.OpenFile(fname, os.O_CREATE|os.O_WRONLY, perms)
	if err != nil {
		t.Errorf("cannot create %s file: %v", fname, err)
		return false
	}
	if _, err := of.WriteString(contents); err != nil {
		t.Errorf("cannot write %s file: %v", fname, err)
		return false
	}
	if err := of.Close(); err != nil {
		t.Errorf("cannot close %s file: %v", fname, err)
		return false
	}
	return true
}

func checkError(t *testing.T, expected string, receivedError error) bool {
	if receivedError != nil {
		received := receivedError.Error()
		if expected == "" {
			t.Errorf("Unexpected error: %s", received)
			return false
		}
		if received != expected {
			t.Errorf(
				`Unexpected error.
Expected: "%s"
Received: "%s"`,
				expected,
				received,
			)
			return false
		}
	} else if expected != "" {
		t.Errorf("Unexpected success. Expected error: %s", expected)
		return false
	}
	return true
}
