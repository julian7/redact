package files

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	// DefaultKeyDir contains standard key directory name inside .git/ directory
	DefaultKeyDir = "redact"
	// DefaultKeyFile contains standard key file name inside key directory
	DefaultKeyFile = "key"
)

func buildKeyDir(gitdir string) string {
	return filepath.Join(gitdir, DefaultKeyDir)
}

func buildKeyFileName(path string) string {
	return filepath.Join(path, DefaultKeyFile)
}

func checkFileMode(name string, fileinfo os.FileInfo, expected os.FileMode) error {
	mode := fileinfo.Mode().Perm()
	if mode&^expected != 0 {
		return errors.Errorf("excessive rights on %s", name)
	}
	if mode&expected != expected {
		return errors.Errorf("insufficient rights on %s", name)
	}
	return nil
}
