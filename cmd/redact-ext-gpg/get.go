package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func cmdGet() *cli.Command {
	cmd := command("get", "Decrypt a Key Exchange copy", getAction)
	cmd.ArgsUsage = "[gpgkey=SELECTOR] [private-key=stdin]"
	cmd.Description = `Prints the decrypted Key Exchange copy to stdout.

  gpgkey=SELECTOR    Select an identity by fingerprint or user ID.
  private-key=stdin  Read an armored OpenPGP private key from stdin.

By default, uses the local GnuPG keyring. If multiple identities match,
retry with gpgkey=SELECTOR.

For an encrypted stdin key, set REDACT_GPG_PASSPHRASE or enter its passphrase
at the terminal prompt.

Examples:

  redact-ext-gpg get gpgkey=alice@example.org
  redact-ext-gpg get private-key=stdin < seckey.asc`

	return cmd
}

func getAction(_ context.Context, cmd *cli.Command) error {
	settings, err := parseSettings(cmd.Args().Slice(), "gpgkey")
	if err != nil {
		return err
	}

	redactRepo, err := openRepo()
	if err != nil {
		return err
	}

	selector := settings.identity()
	if source := settings["private-key"]; source != "" {
		if source != "stdin" {
			return failed(fmt.Errorf("private-key must be stdin"))
		}

		plaintext, err := loadStdinExchangeKey(redactRepo, os.Stdin, selector)
		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(plaintext)

		return failed(err)
	}

	available, err := availableKeys(redactRepo, selector)
	if err != nil {
		return err
	}

	if len(available) == 0 {
		if selector != "" {
			return notApplicable(fmt.Errorf("%w: %q", ErrIdentityHasNoKeyExchangeCopy, selector))
		}

		return notApplicable(ErrNoLocalKeyExchangeCopy)
	}

	if len(available) > 1 {
		fmt.Fprintln(os.Stderr, "Multiple local GnuPG identities have a Key Exchange copy. Retry with gpgkey=<identity>.")
		printCandidates(redactRepo, available)

		return missing(ErrGPGIdentitySelectionRequired)
	}

	plaintext, err := loadExchangeKey(redactRepo, available[0])
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(plaintext)

	return failed(err)
}

func availableKeys(redactRepo *repo.Repo, selector string) ([][]byte, error) {
	keys, warnings, err := gpgutil.GetSecretKeys(selector)
	if err != nil {
		if selector != "" && isNoGPGMatch(err) {
			return nil, notApplicable(fmt.Errorf("no local GnuPG identity matches %q", selector))
		}

		return nil, failed(fmt.Errorf("discovering local GnuPG identities: %w", err))
	}

	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
	}

	available := make([][]byte, 0, len(keys))
	for _, fingerprint := range keys {
		filename := filepath.Join(repo.DefaultKeyExchangeDir, fmt.Sprintf("%x%s", fingerprint, repo.ExtSecret))

		info, err := redactRepo.Workdir.Stat(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return nil, failed(fmt.Errorf("checking encrypted Key Exchange file %s: %w", filename, err))
		}

		if !info.Mode().IsRegular() {
			continue
		}

		available = append(available, fingerprint)
	}

	return available, nil
}

func loadExchangeKey(redactRepo *repo.Repo, fingerprint []byte) ([]byte, error) {
	filename := filepath.Join(repo.DefaultKeyExchangeDir, fmt.Sprintf("%x%s", fingerprint, repo.ExtSecret))

	ciphertext, err := redactRepo.Workdir.Open(filename)
	if err != nil {
		return nil, failed(fmt.Errorf("opening encrypted Key Exchange file: %w", err))
	}

	plaintext, err := gpgutil.DecryptWithKey(ciphertext, fingerprint)

	closeErr := ciphertext.Close()

	if err != nil {
		return nil, failed(fmt.Errorf("decrypting Key Exchange collection: %w", err))
	}

	if closeErr != nil {
		_ = plaintext.Close()

		return nil, failed(fmt.Errorf("closing encrypted Key Exchange file: %w", closeErr))
	}

	defer plaintext.Close()

	data, err := io.ReadAll(plaintext)
	if err != nil {
		return nil, failed(fmt.Errorf("reading decrypted Key Exchange collection: %w", err))
	}

	return data, nil
}

func isNoGPGMatch(err error) bool {
	var exit *exec.ExitError

	return errors.As(err, &exit) && exit.ExitCode() == 2
}

func printCandidates(redactRepo *repo.Repo, fingerprints [][]byte) {
	for _, fingerprint := range fingerprints {
		filename := filepath.Join(repo.DefaultKeyExchangeDir, fmt.Sprintf("%x%s", fingerprint, repo.ExtKeyArmor))

		file, err := redactRepo.Workdir.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read candidate %x: %v\n", fingerprint, err)

			continue
		}

		keys, readErr := gpgutil.LoadPubKey(file, true)
		closeErr := file.Close()

		if readErr != nil {
			fmt.Fprintf(os.Stderr, "Unable to read candidate %x: %v\n", fingerprint, readErr)

			continue
		}

		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "Unable to close candidate %x: %v\n", fingerprint, closeErr)

			continue
		}

		for _, key := range keys {
			if err := gpgutil.PrintKeyTo(os.Stderr, key); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to print candidate %x: %v\n", fingerprint, err)
			}
		}
	}
}
