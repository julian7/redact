package files

import (
	"os"

	"github.com/hectane/go-acl"
	"github.com/spf13/afero"
)

type OsFs struct {
	afero.Fs
}

func NewOsFs() afero.Fs {
	return &OsFs{Fs: afero.NewOsFs()}
}

func (OsFs) Chmod(name string, mode os.FileMode) error {
	return acl.Chmod(name, mode)
}
