//go:build !windows

package main

import "os"

func openTerminal() (*os.File, *os.File, error) {
	terminal, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}

	return terminal, terminal, nil
}
