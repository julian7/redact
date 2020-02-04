//+build !windows

package files

import (
	"fmt"
	"os"
)

func checkFileMode(name string, fileinfo os.FileInfo, expected os.FileMode) error {
	if checkFileModeOnce(name, fileinfo, expected) != nil {
		if err := os.Chmod(name, expected); err != nil {
			return fmt.Errorf("enforcing file mode on %q: %w", name, err)
		}

		return checkFileModeOnce(name, fileinfo, expected)
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
