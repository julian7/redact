package kx

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/julian7/redact/repo"
)

type walkErrorFS struct {
	billy.Filesystem
	op   string
	path string
	err  error
}

func (fs *walkErrorFS) Lstat(name string) (os.FileInfo, error) {
	if fs.op == "lstat" && name == fs.path {
		return nil, &os.PathError{Op: fs.op, Path: name, Err: fs.err}
	}

	return fs.Filesystem.Lstat(name)
}

func (fs *walkErrorFS) ReadDir(name string) ([]os.FileInfo, error) {
	if fs.op == "readdir" && name == fs.path {
		return nil, &os.PathError{Op: fs.op, Path: name, Err: fs.err}
	}

	return fs.Filesystem.ReadDir(name)
}

func TestRefreshWalkErrors(t *testing.T) {
	root := repo.DefaultKeyExchangeDir
	nested := filepath.Join(root, "nested")
	tests := []struct {
		name string
		op   string
		path string
		err  error
	}{
		{name: "root stat permission", op: "lstat", path: root, err: os.ErrPermission},
		{name: "root read permission", op: "readdir", path: root, err: os.ErrPermission},
		{name: "root disappears during read", op: "readdir", path: root, err: os.ErrNotExist},
		{name: "nested stat permission", op: "lstat", path: nested, err: os.ErrPermission},
		{name: "nested disappears", op: "lstat", path: nested, err: os.ErrNotExist},
		{name: "nested read permission", op: "readdir", path: nested, err: os.ErrPermission},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := &walkErrorFS{Filesystem: memfs.New(), op: tc.op, path: tc.path, err: tc.err}
			if err := fs.MkdirAll(nested, 0755); err != nil {
				t.Fatal(err)
			}

			updated, err := refreshWithoutRecipients(t, fs)
			if !errors.Is(err, tc.err) || updated != 0 {
				t.Fatalf("expected %v and zero updates, got %v and %d updates", tc.err, err, updated)
			}
		})
	}
}

func TestRefreshWithoutRecipients(t *testing.T) {
	for _, present := range []bool{false, true} {
		name := "missing directory"
		if present {
			name = "empty directory"
		}

		t.Run(name, func(t *testing.T) {
			fs := memfs.New()
			if present {
				if err := fs.MkdirAll(repo.DefaultKeyExchangeDir, 0755); err != nil {
					t.Fatal(err)
				}
			}

			updated, err := refreshWithoutRecipients(t, fs)
			if err != nil || updated != 0 {
				t.Fatalf("expected no updates, got %d updates and %v", updated, err)
			}

			_, err = fs.Stat(repo.DefaultKeyExchangeDir)
			if !present && !errors.Is(err, os.ErrNotExist) {
				t.Fatal("refresh created the missing exchange directory")
			}
		})
	}
}

func TestRefreshExchangeNotDirectory(t *testing.T) {
	fs := memfs.New()
	if err := util.WriteFile(fs, repo.DefaultKeyExchangeDir, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := refreshWithoutRecipients(t, fs)
	if !errors.Is(err, repo.ErrExchangeIsNotDir) {
		t.Fatalf("expected exchange directory error, got %v", err)
	}
}

func refreshWithoutRecipients(t *testing.T, fs billy.Filesystem) (int, error) {
	t.Helper()

	return UpdateGPGKeysInKX(&repo.Repo{Workdir: fs}, func(io.Writer) error {
		t.Fatal("unexpected recipient serialization")

		return nil
	})
}
