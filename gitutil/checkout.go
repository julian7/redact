package gitutil

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/julian7/redact/log"
)

// Checkout checks out files provided
func Checkout(files []string) error {
	attrs := []string{
		"checkout",
		"--",
	}
	attrs = append(attrs, files...)
	cmd := exec.Command("git", attrs...)

	errStream, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func(reader io.ReadCloser) {
		bufreader := bufio.NewReader(reader)
		l := log.Log()
		for {
			line, _, err := bufreader.ReadLine()
			if err != nil {
				return
			}
			l.Warnf("checkout: %s", line)
		}
	}(errStream)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("checking out files: %w", err)
	}

	return nil
}
