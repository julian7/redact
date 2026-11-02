package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"golang.org/x/term"
)

const passphraseEnv = "REDACT_GPG_PASSPHRASE" //nolint:gosec // Environment variable name.

// defaultPassphrasePrompt asks for the passphrase of a private key whose source
// has no name of its own. A repository's extension config can name that source
// with prompt=, so that a key a filter command supplies is named after the command.
const defaultPassphrasePrompt = "OpenPGP private-key passphrase:" //nolint:gosec // Prompt wording, not a credential.

var (
	ErrPassphraseUnavailable = errors.New("no terminal available; set " + passphraseEnv)
	promptPassphrase         = terminalPassphrase
)

// passphrasePrompt words the request for a private-key passphrase, naming the
// source the extension was told to read the key from.
func passphrasePrompt(label string) string {
	if label == "" {
		return defaultPassphrasePrompt
	}

	return fmt.Sprintf("OpenPGP private-key passphrase for %q:", label)
}

func privateKeyPrompt(label string) openpgp.PromptFunction {
	attempted := false

	return func(keys []openpgp.Key, _ bool) ([]byte, error) {
		if attempted || len(keys) == 0 {
			return nil, failed(fmt.Errorf("unable to decrypt with the supplied private key"))
		}

		attempted = true

		passphrase, err := privateKeyPassphrase(label)
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

func privateKeyPassphrase(label string) ([]byte, error) {
	if value, ok := os.LookupEnv(passphraseEnv); ok {
		return []byte(value), nil
	}

	value, err := promptPassphrase(passphrasePrompt(label))
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
