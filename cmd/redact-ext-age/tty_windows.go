//go:build windows

package main

import (
	"os"
)

func openTerminal() (*os.File, *os.File, error) {
	input, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}

	output, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		input.Close() // nolint:errcheck

		return nil, nil, err
	}

	return input, output, nil
}
