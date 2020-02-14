package gitutil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Checkout checks out files provided
func Checkout(files []string) ([]*NamedError, error) {
	attrs := []string{
		"checkout",
		"--",
	}
	attrs = append(attrs, files...)
	cmd := exec.Command("git", attrs...)

	errStream, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	issues := []*NamedError{}
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(reader io.ReadCloser) {
		defer wg.Done()

		bufreader := bufio.NewReader(reader)

		for {
			line, _, err := bufreader.ReadLine()
			if err != nil {
				return
			}

			issues = append(issues, NewError(string(line), errors.New("git checkout")))
		}
	}(errStream)

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("checking out files: %w", err)
	}

	wg.Wait()

	return issues, nil
}
