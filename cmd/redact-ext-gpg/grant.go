package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/kx"
	"github.com/urfave/cli/v3"
)

func cmdGrant() *cli.Command {
	cmd := command("grant", "Register GPG recipients and publish the supplied collection", grantAction)
	cmd.ArgsUsage = "[recipient=SELECTOR] [file=PATH] [armor=PATH]"
	cmd.Description = `At least one recipient source is required:

  recipient=SELECTOR  Export a recipient from the local GnuPG keyring.
  file=PATH           Read a binary OpenPGP public key.
  armor=PATH          Read an ASCII-armored OpenPGP public key.

Repeat a setting or provide comma-separated values to grant access to multiple
recipients. Through redact, pass these as --option KEY=VALUE. Examples:

  redact grant --ext gpg --option recipient=alice@example.org
  redact grant --ext gpg --option file=recipient.gpg
  redact grant --ext gpg --option armor=recipient.asc`

	return cmd
}

func grantAction(_ context.Context, cmd *cli.Command) error {
	settings, err := parseSettings(cmd.Args().Slice(), "recipient", "file", "armor")
	if err != nil {
		return err
	}

	if len(settings.values("recipient"))+len(settings.values("file"))+len(settings.values("armor")) == 0 {
		return showGrantCandidates()
	}

	key, err := readCollection(os.Stdin)
	if err != nil {
		return err
	}

	redactRepo, err := openRepo()
	if err != nil {
		return err
	}

	keys, failures := loadGrantKeys(settings)
	if len(keys) == 0 {
		if len(failures) > 0 {
			return failed(errors.Join(failures...))
		}

		return missing(ErrNoRecipientPublicKeys)
	}

	saved := 0

	for _, recipient := range keys {
		if err := gpgutil.PrintKeyTo(os.Stderr, recipient); err != nil {
			failures = append(failures, fmt.Errorf("printing recipient: %w", err))
		}

		if err := kx.SaveGPGKeyToKX(redactRepo, recipient, key.SaveTo); err != nil {
			failures = append(failures, fmt.Errorf(
				"encrypting Key Exchange collection for %x: %w",
				recipient.PrimaryKey.Fingerprint,
				err,
			))
			fmt.Fprintf(os.Stderr, "Unable to grant access to %x: %v\n", recipient.PrimaryKey.Fingerprint, err)

			continue
		}

		if err := kx.SaveGPGPubkeyToKX(redactRepo, recipient); err != nil {
			failures = append(failures, fmt.Errorf("saving recipient %x: %w", recipient.PrimaryKey.Fingerprint, err))
			fmt.Fprintf(os.Stderr, "Unable to register recipient %x: %v\n", recipient.PrimaryKey.Fingerprint, err)

			continue
		}

		saved++
	}

	if saved > 0 {
		fmt.Fprintf(os.Stderr, "Granted access to %d recipient%s. Commit and push repository changes to distribute them.\n",
			saved, map[bool]string{true: "", false: "s"}[saved == 1])
	}

	if len(failures) > 0 {
		return failed(errors.Join(failures...))
	}

	return nil
}

func showGrantCandidates() error {
	data, err := gpgutil.ExportKey(nil)
	if err != nil {
		return failed(fmt.Errorf("listing local GnuPG public keys: %w", err))
	}

	if len(data) == 0 {
		fmt.Fprintln(os.Stderr, "No public keys found in the local GnuPG keyring.")
		fmt.Fprintln(os.Stderr, "Import a recipient key or provide file=PATH or armor=PATH.")

		return missing(ErrGrantRecipientInputRequired)
	}

	keys, err := gpgutil.LoadPubKey(bytes.NewReader(data), true)
	if err != nil {
		return failed(fmt.Errorf("reading local GnuPG public keys: %w", err))
	}

	fmt.Fprintln(os.Stderr, "Potential recipients in the local GnuPG keyring:")

	for _, key := range keys {
		if err := gpgutil.PrintKeyTo(os.Stderr, key); err != nil {
			return failed(fmt.Errorf("printing GnuPG recipient candidate: %w", err))
		}
	}

	fmt.Fprintln(os.Stderr, "Retry with --option recipient=<fingerprint>, or provide file=PATH or armor=PATH.")

	return missing(ErrGrantRecipientInputRequired)
}

func loadGrantKeys(settings settings) ([]*openpgp.Entity, []error) {
	keys := make([]*openpgp.Entity, 0)

	var failures []error

	for _, selector := range settings.values("recipient") {
		data, err := gpgutil.ExportKey([]string{selector})
		if err != nil {
			failures = append(failures, fmt.Errorf("exporting GnuPG recipient %q: %w", selector, err))
			fmt.Fprintf(os.Stderr, "Unable to export recipient %q: %v\n", selector, err)

			continue
		}

		loaded, err := gpgutil.LoadPubKey(bytes.NewReader(data), true)
		if err != nil {
			failures = append(failures, fmt.Errorf("loading GnuPG recipient %q: %w", selector, err))
			fmt.Fprintf(os.Stderr, "Unable to load recipient %q: %v\n", selector, err)

			continue
		}

		keys = append(keys, loaded...)
	}

	for _, filename := range settings.values("file") {
		loaded, err := readPublicKeys(filename, false)
		if err != nil {
			failures = append(failures, fmt.Errorf("loading binary public key file %q: %w", filename, err))
			fmt.Fprintf(os.Stderr, "Unable to load public key file %q: %v\n", filename, err)

			continue
		}

		keys = append(keys, loaded...)
	}

	for _, filename := range settings.values("armor") {
		loaded, err := readPublicKeys(filename, true)
		if err != nil {
			failures = append(failures, fmt.Errorf("loading armored public key file %q: %w", filename, err))
			fmt.Fprintf(os.Stderr, "Unable to load armored public key file %q: %v\n", filename, err)

			continue
		}

		keys = append(keys, loaded...)
	}

	return dedupeKeys(keys), failures
}

func dedupeKeys(keys []*openpgp.Entity) []*openpgp.Entity {
	seen := make(map[string]struct{}, len(keys))

	result := make([]*openpgp.Entity, 0, len(keys))
	for _, key := range keys {
		fingerprint := fmt.Sprintf("%x", key.PrimaryKey.Fingerprint)
		if _, ok := seen[fingerprint]; ok {
			continue
		}

		seen[fingerprint] = struct{}{}

		result = append(result, key)
	}

	return result
}
