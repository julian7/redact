package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"golang.org/x/term"
)

const passphraseEnv = "REDACT_GPG_PASSPHRASE" //nolint:gosec // Environment variable name.

var (
	ErrPassphraseUnavailable = errors.New("no terminal available; set " + passphraseEnv)
	promptPassphrase         = terminalPassphrase
)

func privateKeyPrompt() openpgp.PromptFunction {
	attempted := false

	return func(keys []openpgp.Key, _ bool) ([]byte, error) {
		if attempted || len(keys) == 0 {
			return nil, failed(fmt.Errorf("unable to decrypt with the supplied private key"))
		}

		attempted = true

		passphrase, err := privateKeyPassphrase()
		if err != nil {
			return nil, err
		}
		defer clear(passphrase)

		unlocked := false

		for _, key := range keys {
			if err := key.PrivateKey.Decrypt(passphrase); err == nil {
				unlocked = true
			}
		}

		if !unlocked {
			return nil, failed(fmt.Errorf("incorrect private-key passphrase"))
		}

		return nil, nil
	}
}

func privateKeyPassphrase() ([]byte, error) {
	if value, ok := os.LookupEnv(passphraseEnv); ok {
		return []byte(value), nil
	}

	value, err := promptPassphrase("OpenPGP private-key passphrase:")
	if errors.Is(err, ErrPassphraseUnavailable) {
		return nil, missing(err)
	}

	if err != nil {
		return nil, failed(fmt.Errorf("prompting for private-key passphrase: %w", err))
	}

	return value, nil
}

func terminalPassphrase(prompt string) ([]byte, error) {
	input, output, err := openTerminal()
	if err != nil {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return nil, ErrPassphraseUnavailable
		}

		input, output = os.Stdin, os.Stderr
	} else {
		defer input.Close()

		if output != input {
			defer output.Close()
		}
	}

	if _, err := fmt.Fprint(output, prompt+" "); err != nil {
		return nil, fmt.Errorf("writing passphrase prompt: %w", err)
	}

	value, err := term.ReadPassword(int(input.Fd()))
	_, _ = fmt.Fprintln(output)

	if err != nil {
		clear(value)

		return nil, fmt.Errorf("reading passphrase: %w", err)
	}

	return value, nil
}
