package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/julian7/redact/ageutil"
)

// Replaceable so tests never read the controlling terminal.
var promptPassphrase ageutil.PassphraseFunc = identityPassphrase

func identityPassphrase(prompt string) (string, error) {
	if value, ok := os.LookupEnv("REDACT_AGE_PASSPHRASE"); ok {
		return value, nil
	}

	return terminalPassphrase(prompt)
}

// Use the controlling terminal because stdin and stdout carry extension data.
func terminalPassphrase(prompt string) (string, error) {
	input, output, err := withTerminal()
	if err != nil {
		return "", err
	}
	defer closeTerminal(input, output)

	if _, err := fmt.Fprint(output, prompt+" "); err != nil {
		return "", fmt.Errorf("writing the passphrase prompt: %w", err)
	}

	value, err := term.ReadPassword(int(input.Fd()))
	if err != nil {
		return "", fmt.Errorf("reading the passphrase: %w", err)
	}

	// ReadPassword switches echo off, so the Enter key that ends the passphrase
	// leaves no line behind: print one, or the next message starts on the same
	// line as the prompt.
	if _, err := fmt.Fprintln(output); err != nil {
		return "", fmt.Errorf("ending the passphrase prompt: %w", err)
	}

	return string(value), nil
}

func withTerminal() (*os.File, *os.File, error) {
	input, output, err := openTerminal()
	if err == nil {
		return input, output, nil
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		return os.Stdin, os.Stderr, nil
	}

	return nil, nil, ageutil.ErrPassphraseUnavailable
}

func closeTerminal(input, output *os.File) {
	if output != nil && output != os.Stderr {
		output.Close() // nolint:errcheck
	}

	if input != nil && input != output && input != os.Stdin {
		input.Close() // nolint:errcheck
	}
}
