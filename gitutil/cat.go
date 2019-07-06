package gitutil

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

// Cat "cat"s a file by SHA1 hash
func Cat(objectID []byte) (io.ReadCloser, error) {
	cmd := exec.Command(
		"git",
		"cat-file",
		"blob",
		fmt.Sprintf("%x", objectID),
	)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "getting git command output pipe")
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return out, nil
}
