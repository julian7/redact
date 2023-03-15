//go:build windows

package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/go-git/go-billy/v5"
)

func checkFileMode(fs billy.Filesystem, name, filename string, expected os.FileMode, strict bool) error {
	var syserr syscall.Errno

	if chfs, ok := fs.(billy.Change); ok {
		if err := chfs.Chmod(filename, expected); err != nil && strict && (!errors.As(err, &syserr) || syserr != 0) {
			return fmt.Errorf("setting permissions for %s: %w", name, syserr)
		}
	}

	return nil
}
