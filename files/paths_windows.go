//go:build windows
// +build windows

package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/afero"
)

func checkFileMode(fs afero.Fs, name, filename string, expected os.FileMode, strict bool) error {
	var syserr syscall.Errno

	if err := fs.Chmod(filename, expected); err != nil && strict && (!errors.As(err, &syserr) || syserr != 0) {
		return fmt.Errorf("setting permissions for %s: %w", name, syserr)
	}

	return nil
}
