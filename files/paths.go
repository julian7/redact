package files

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func buildKeyDir(gitdir string) string {
	return filepath.Join(gitdir, "redact")
}

func buildKeyFileName(path string) string {
	return filepath.Join(path, "key")
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
