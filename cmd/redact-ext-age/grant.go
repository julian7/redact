package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"filippo.io/age"

	"github.com/urfave/cli/v3"

	"github.com/julian7/redact/ageutil"
)

func cmdGrant() *cli.Command {
	grant := command("grant", "Grant access to the Key Exchange collection", grantAction)

	grant.ArgsUsage = "[recipient=RECIPIENT] [file=PATH]"
	grant.Description = `At least one recipient source is required:

  recipient=RECIPIENT  An age recipient (public key) in bech32 form.
  file=PATH            Read an age recipients file (one recipient per line).

Repeat a setting or provide comma-separated values to grant access to multiple
recipients. Through redact, pass these as --option KEY=VALUE. Examples:

  redact grant --ext age --option recipient=age1v0p8q6z4x2n5m3k7h9j0f2d4s6a8c0e2t4l6k8m0q2s4u6w0y1a3c5r7zg
  redact grant --ext age --option file=recipients.txt`

	return grant
}

func grantAction(_ context.Context, cmd *cli.Command) error {
	settings, err := parseSettings(cmd.Args().Slice(), "recipient", "file")
	if err != nil {
		return err
	}

	if len(settings.values("recipient"))+len(settings.values("file")) == 0 {
		return showGrantCandidates()
	}

	collection, err := readCollection(os.Stdin)
	if err != nil {
		return err
	}

	recipients, failures := loadGrantRecipients(settings)
	if len(recipients) == 0 {
		if len(failures) > 0 {
			return failed(errors.Join(failures...))
		}

		return missing(ErrNoRecipientPublicKeys)
	}

	saved := 0

	for _, item := range recipients {
		if err := ageutil.PrintRecipientTo(os.Stderr, item.text); err != nil {
			failures = append(failures, fmt.Errorf("printing recipient: %w", err))

			continue
		}

		if err := saveAgeKeyToKX(item.key, collection); err != nil {
			failures = append(failures, fmt.Errorf("encrypting Key Exchange collection for %s: %w", item.text, err))
			fmt.Fprintf(os.Stderr, "Unable to grant access to %s: %v\n", item.text, err)

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

type grantRecipient struct {
	key  age.Recipient
	text string
}

func loadGrantRecipients(settings settings) ([]grantRecipient, []error) {
	var failures []error

	keys := make([]age.Recipient, 0)

	for _, selector := range settings.values("recipient") {
		parsed, err := ageutil.ParseRecipient(selector)
		if err != nil {
			failures = append(failures, fmt.Errorf("parsing age recipient %q: %w", selector, err))
			fmt.Fprintf(os.Stderr, "Unable to parse recipient %q: %v\n", selector, err)

			continue
		}

		keys = append(keys, parsed)
	}

	for _, filename := range settings.values("file") {
		parsed, err := ageutil.ParseRecipientsFile(filename)
		if err != nil {
			failures = append(failures, fmt.Errorf("loading recipient file %q: %w", filename, err))
			fmt.Fprintf(os.Stderr, "Unable to load recipient file %q: %v\n", filename, err)

			continue
		}

		keys = append(keys, parsed...)
	}

	unique, textFailures := dedupeRecipients(keys)
	failures = append(failures, textFailures...)

	return unique, failures
}

func dedupeRecipients(keys []age.Recipient) ([]grantRecipient, []error) {
	unique := make([]grantRecipient, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))

	var failures []error

	for _, key := range keys {
		text, err := ageutil.RecipientText(key)
		if err != nil {
			failures = append(failures, err)

			continue
		}

		if _, ok := seen[text]; ok {
			continue
		}

		seen[text] = struct{}{}
		unique = append(unique, grantRecipient{key: key, text: text})
	}

	return unique, failures
}

func showGrantCandidates() error {
	identities, err := loadIdentities(nil, promptPassphrase)
	if err != nil {
		return err
	}

	if len(identities) == 0 {
		fmt.Fprintln(os.Stderr, "No age identities were found. Create one with 'age-keygen' or provide file=PATH.")
		fmt.Fprintln(os.Stderr, "Retry with --option recipient=<recipient>, or provide file=PATH.")

		return missing(ErrGrantRecipientInputRequired)
	}

	fmt.Fprintln(os.Stderr, "age recipients available from local identities:")

	for _, item := range identities {
		if _, err := fmt.Fprintln(os.Stderr, item.text); err != nil {
			return failed(fmt.Errorf("printing age recipient candidate: %w", err))
		}
	}

	fmt.Fprintln(os.Stderr, "Retry with --option recipient=<recipient>, or provide file=PATH.")

	return missing(ErrGrantRecipientInputRequired)
}
