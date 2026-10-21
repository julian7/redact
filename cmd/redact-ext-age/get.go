package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"filippo.io/age"

	"github.com/urfave/cli/v3"

	"github.com/julian7/redact/ageutil"
)

func cmdGet() *cli.Command {
	cmd := command("get", "Decrypt a Key Exchange copy", getAction)
	cmd.ArgsUsage = "[agekey=RECIPIENT] [keyfile=PATH] [private-key=stdin]"
	cmd.Description = `Prints the decrypted Key Exchange copy to stdout.

  agekey=RECIPIENT    Select an identity by its fingerprint or age recipient.
  keyfile=PATH        Read an age identities file; repeat for multiple files.
  private-key=stdin   Read plain or encrypted age identities from stdin.

Without keyfile, uses AGE_IDENTITIES_FILE or the user's age identities file.
For encrypted identities, set REDACT_AGE_PASSPHRASE or enter the passphrase
at the terminal prompt.

Examples:

  redact-ext-age get keyfile=identities.txt
  redact-ext-age get private-key=stdin < identity.txt`

	return cmd
}

type localIdentity struct {
	identity age.Identity
	text     string
}

func getAction(_ context.Context, cmd *cli.Command) error {
	settings, err := parseSettings(cmd.Args().Slice(), "agekey", "keyfile")
	if err != nil {
		return err
	}

	selector := settings.identity()

	available, err := availableKeys(settings, selector, promptPassphrase)
	if err != nil {
		return fmt.Errorf("cannot find available key: %w", err)
	}

	if len(available) == 0 {
		if selector != "" {
			return notApplicable(fmt.Errorf("%w: %q", ErrIdentityHasNoKeyExchangeCopy, selector))
		}

		return notApplicable(ErrNoLocalKeyExchangeCopy)
	}

	if len(available) > 1 {
		fmt.Fprintln(os.Stderr, "Multiple local age identities have a Key Exchange copy. Retry with agekey=<recipient>.")
		printCandidates(available)

		return missing(ErrAgeIdentitySelectionRequired)
	}

	collection, err := loadExchangeKey(available[0])
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(collection); err != nil {
		return failed(fmt.Errorf("exporting Key Exchange collection: %w", err))
	}

	return nil
}

// Skip unreadable files so other identities or extensions remain usable.
// A rejected passphrase stops dispatch because another extension cannot fix it.
func loadIdentities(settings settings, prompt ageutil.PassphraseFunc) ([]localIdentity, error) {
	if source := settings["private-key"]; source != "" {
		if source != "stdin" || len(settings.values("keyfile")) > 0 {
			return nil, failed(fmt.Errorf("use private-key=stdin without keyfile"))
		}

		parsed, err := ageutil.ReadIdentities(os.Stdin, "stdin", prompt)
		if errors.Is(err, ageutil.ErrPassphraseUnavailable) {
			return nil, missing(err)
		}

		if err != nil {
			return nil, failed(fmt.Errorf("reading age identities from stdin: %w", err))
		}

		return identitiesOf(parsed), nil
	}

	filenames := settings.values("keyfile")
	if len(filenames) == 0 {
		if def := ageutil.DefaultIdentityFile(); def != "" {
			filenames = []string{def}
		}
	}

	identities := make([]localIdentity, 0)
	locked := ""

	for _, filename := range filenames {
		parsed, err := ageutil.ParseIdentitiesFile(filename, prompt)

		switch {
		case err == nil:
		case errors.Is(err, ageutil.ErrIncorrectPassphrase), errors.Is(err, ageutil.ErrNoPassphraseEntered):
			return nil, failed(fmt.Errorf("using age identities file %s: %w", filename, err))
		case errors.Is(err, ageutil.ErrPassphraseUnavailable):
			fmt.Fprintf(os.Stderr, "Ignoring age identities file %s: no terminal to type its passphrase. "+
				"Run redact from a terminal, or use an identities file without a passphrase.\n", filename)
			locked = filename

			continue
		default:
			fmt.Fprintf(os.Stderr, "Ignoring age identities file %s: %v\n", filename, err)

			continue
		}

		identities = append(identities, identitiesOf(parsed)...)
	}

	if locked != "" && len(identities) == 0 {
		return nil, notApplicable(fmt.Errorf("age identities file %s: %w", locked, ageutil.ErrPassphraseUnavailable))
	}

	return identities, nil
}

// Skip unsupported identities because encrypted copies are indexed by recipient.
func identitiesOf(parsed []age.Identity) []localIdentity {
	identities := make([]localIdentity, 0, len(parsed))

	for _, identity := range parsed {
		text, err := ageutil.RecipientString(identity)
		if err != nil {
			continue
		}

		identities = append(identities, localIdentity{identity: identity, text: text})
	}

	return identities
}

func availableKeys(settings settings, selector string, prompt ageutil.PassphraseFunc) ([]localIdentity, error) {
	directory, err := exchangeDirectory()
	if err != nil {
		return nil, failed(err)
	}

	identities, err := loadIdentities(settings, prompt)
	if err != nil {
		return nil, fmt.Errorf("loading identities: %w", err)
	}

	if selector != "" {
		matched := make([]localIdentity, 0, len(identities))
		for _, item := range identities {
			if matchesRecipient(item.text, selector) {
				matched = append(matched, item)
			}
		}

		if len(matched) == 0 {
			return nil, notApplicable(fmt.Errorf("no local age identity matches %q", selector))
		}

		identities = matched
	}

	available := make([]localIdentity, 0, len(identities))

	for _, item := range identities {
		filename := filepath.Join(directory, ExchangeAgeSecretKeyFile(item.text))

		info, err := os.Stat(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return nil, failed(fmt.Errorf("checking encrypted Key Exchange file %s: %w", filename, err))
		}

		if !info.Mode().IsRegular() {
			continue
		}

		available = append(available, item)
	}

	return available, nil
}

func printCandidates(candidates []localIdentity) {
	for _, item := range candidates {
		if err := ageutil.PrintRecipientTo(os.Stderr, item.text); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to print candidate %s: %v\n", item.text, err)
		}
	}
}

func loadExchangeKey(item localIdentity) ([]byte, error) {
	directory, err := exchangeDirectory()
	if err != nil {
		return nil, failed(err)
	}

	filename := filepath.Join(directory, ExchangeAgeSecretKeyFile(item.text))

	ciphertext, err := os.ReadFile(filename)
	if err != nil {
		return nil, failed(fmt.Errorf("opening encrypted Key Exchange file: %w", err))
	}

	plaintext := bytes.Buffer{}
	if err := ageutil.Decrypt(bytes.NewReader(ciphertext), &plaintext, item.identity); err != nil {
		return nil, failed(fmt.Errorf("decrypting Key Exchange collection: %w", err))
	}

	return plaintext.Bytes(), nil
}
