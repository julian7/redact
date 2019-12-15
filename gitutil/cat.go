package gitutil

import (
	"fmt"
	"io"
	"os/exec"
)

// Cat "cat"s a file by SHA1 hash
func Cat(objectID []byte) (io.ReadCloser, error) {
	cmd := exec.Command( //nolint:gosec
		"git",
		"cat-file",
		"blob",
		fmt.Sprintf("%x", objectID),
	)

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("getting git command output pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return out, nil
}
