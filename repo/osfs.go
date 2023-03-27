package repo

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5"
)

type OS struct {
	billy.Filesystem
}

func NewOSFS(fs billy.Filesystem) *OS {
	return &OS{Filesystem: fs}
}

func (osfs *OS) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(filepath.Join(osfs.Root(), name), mode)
}

func (osfs *OS) Lchown(name string, uid, gid int) error {
	return os.Lchown(filepath.Join(osfs.Root(), name), uid, gid)
}

func (osfs *OS) Chown(name string, uid, gid int) error {
	return os.Chown(filepath.Join(osfs.Root(), name), uid, gid)
}

func (osfs *OS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(filepath.Join(osfs.Root(), name), atime, mtime)
}
