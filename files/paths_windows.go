//+build windows

package files

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/hectane/go-acl"
)

func checkFileMode(name string, fileinfo os.FileInfo, expected os.FileMode) error {
	var syserr syscall.Errno

	if err := acl.Chmod(name, expected); err != nil && (!errors.As(err, &syserr) || syserr != 0) {
		return fmt.Errorf("setting permissions for %q: %T", name, syserr)
	}

	return nil
}
