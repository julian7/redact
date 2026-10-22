package ext

import (
	"fmt"
	"os"
	"os/exec"
)

// run keeps the pipeline in memory and waits for both processes before
// accepting their output.
func (ext *Ext) run(operation string, target *exec.Cmd) error {
	settings := ext.operation(operation)
	if settings == nil || len(settings.Filter) == 0 {
		return ext.result(operation, target.Run())
	}

	filter := settings.Filter

	if filter[0] == "" {
		return ext.filterError(operation, fmt.Errorf("empty executable"))
	}

	// The configured argument vector is deliberately executable, like Command.
	input := exec.Command(filter[0], filter[1:]...) //nolint:gosec
	input.Dir = target.Dir
	input.Env = target.Environ()
	input.Stdin = target.Stdin
	input.Stderr = os.Stderr

	reader, writer, err := os.Pipe()
	if err != nil {
		return ext.filterError(operation, err)
	}
	defer reader.Close()
	defer writer.Close()

	input.Stdout = writer
	target.Stdin = reader

	if err := target.Start(); err != nil {
		return ext.result(operation, err)
	}

	if err := input.Start(); err != nil {
		_ = target.Process.Kill()
		_ = target.Wait()

		return ext.filterError(operation, err)
	}

	// Only the children retain pipe handles, so EOF and broken-pipe detection
	// work even when either child exits without consuming its entire stream.
	_ = reader.Close()
	_ = writer.Close()

	return ext.waitFilter(operation, input, target)
}

func (ext *Ext) filterError(operation string, err error) error {
	return fmt.Errorf("extension %s %s filter: %w", ext.name, operation, &Error{Code: Failure, Err: err})
}

func (ext *Ext) waitFilter(operation string, input, target *exec.Cmd) error {
	type result struct {
		input bool
		err   error
	}

	results := make(chan result, 2)

	go func() { results <- result{input: true, err: input.Wait()} }()
	go func() { results <- result{err: target.Wait()} }()

	first := <-results
	if first.err != nil {
		other := input
		if first.input {
			other = target
		}

		_ = other.Process.Kill()
	}

	second := <-results
	for _, result := range []result{first, second} {
		if result.err != nil {
			if result.input {
				return ext.filterError(operation, result.err)
			}

			return ext.result(operation, result.err)
		}
	}

	return nil
}
