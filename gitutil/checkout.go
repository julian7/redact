package gitutil

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// Checkout checks out files provided
func Checkout(files []string, force bool) ([]*NamedError, error) {
	attrs := []string{
		"checkout",
		"--",
	}
	attrs = append(attrs, files...)
	cmd := exec.Command("git", attrs...)

	if force {
		cmd.Env = append(cmd.Env, "REDACT_GIT_CLEAN_EPOCH=0")
	}

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

			issues = append(issues, NewError(string(line), ErrGitCheckout))
		}
	}(errStream)

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("checking out files: %w", err)
	}

	wg.Wait()

	return issues, nil
}
