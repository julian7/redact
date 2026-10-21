package ext

import (
	"fmt"
	"os"
	"os/exec"
)

// get connects the input command directly to the extension without a shell or
// temporary files. Output is accepted only after both processes succeed.
func (ext *Ext) get(target *exec.Cmd) error {
	if len(ext.Input) == 0 {
		return ext.result("get", target.Run())
	}

	if ext.Input[0] == "" {
		return ext.inputError(fmt.Errorf("empty executable"))
	}

	// The configured argument vector is deliberately executable, like Command.
	input := exec.Command(ext.Input[0], ext.Input[1:]...) //nolint:gosec
	input.Dir = target.Dir
	input.Env = target.Environ()
	input.Stdin = os.Stdin
	input.Stderr = os.Stderr

	reader, writer, err := os.Pipe()
	if err != nil {
		return ext.inputError(err)
	}
	defer reader.Close()
	defer writer.Close()

	input.Stdout = writer
	target.Stdin = reader

	if err := target.Start(); err != nil {
		return ext.result("get", err)
	}

	if err := input.Start(); err != nil {
		_ = target.Process.Kill()
		_ = target.Wait()

		return ext.inputError(err)
	}

	// Only the children retain pipe handles, so EOF and broken-pipe detection
	// work even when either child exits without consuming its entire stream.
	_ = reader.Close()
	_ = writer.Close()

	return ext.waitInput(input, target)
}

func (ext *Ext) inputError(err error) error {
	return fmt.Errorf("extension %s get input command: %w", ext.name, &Error{Code: Failure, Err: err})
}

func (ext *Ext) waitInput(input, target *exec.Cmd) error {
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
				return ext.inputError(result.err)
			}

			return ext.result("get", result.err)
		}
	}

	return nil
}
