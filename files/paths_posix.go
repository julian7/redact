//go:build !windows

package files

import (
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5"
)

func checkFileMode(fs billy.Filesystem, name, filename string, expected os.FileMode, strict bool) error {
	st, err := fs.Stat(filename)
	if err != nil {
		return fmt.Errorf("%s %q: %w", name, filename, err)
	}

	if checkFileModeOnce(name, st, expected) != nil {
		if !strict {
			return nil
		}

		if chfs, ok := fs.(billy.Change); ok {
			if err := chfs.Chmod(filename, expected); err != nil {
				return fmt.Errorf("enforcing file mode on %s: %w", name, err)
			}
		}

		st, err := fs.Stat(filename)
		if err != nil {
			return fmt.Errorf("open %s: %w", name, err)
		}

		return checkFileModeOnce(name, st, expected)
	}

	return nil
}

func checkFileModeOnce(name string, fileinfo os.FileInfo, expected os.FileMode) error {
	mode := fileinfo.Mode().Perm()

	if mode&^expected != 0 {
		return fmt.Errorf("excessive rights on %s", name)
	}

	if mode&expected != expected {
		return fmt.Errorf("insufficient rights on %s", name)
	}

	return nil
}
