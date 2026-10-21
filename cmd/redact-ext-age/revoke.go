package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/julian7/redact/ageutil"
)

func cmdRevoke() *cli.Command {
	cmd := command("revoke", "Remove matching age recipient registrations", revokeAction)
	cmd.ArgsUsage = "[recipient=RECIPIENT] [file=PATH]"
	cmd.Description = `Removes encrypted Key Exchange copies for the selected recipients.

  recipient=RECIPIENT  An age recipient or fingerprint; partial fingerprints are accepted.
  file=PATH            Read an age recipients file (one recipient per line).

At least one recipient source is required. Repeat a setting or provide
comma-separated values to revoke multiple recipients.`

	return cmd
}

func revokeAction(_ context.Context, cmd *cli.Command) error {
	settings, err := parseSettings(cmd.Args().Slice(), "recipient", "file")
	if err != nil {
		return err
	}

	if len(settings.values("recipient"))+len(settings.values("file")) == 0 {
		return missing(ErrRevokeRecipientInputRequired)
	}

	selectors, failures := revokeSelectors(settings)

	recipients, scanFailures := readRecipients()
	failures = append(failures, scanFailures...)

	matched := matchingRecipients(recipients, selectors)

	if len(matched) == 0 {
		if len(failures) > 0 {
			return failed(errors.Join(failures...))
		}

		return missing(fmt.Errorf("no registered age recipient matches %q", selectors))
	}

	removed, removeFailures := removeRecipients(matched)
	failures = append(failures, removeFailures...)

	fmt.Fprintf(os.Stderr,
		"Removed %d age recipient registration%s. Commit and push repository changes to distribute revocation.\n",
		removed, map[bool]string{true: "", false: "s"}[removed == 1],
	)

	if len(failures) > 0 {
		return failed(errors.Join(failures...))
	}

	return nil
}

func revokeSelectors(settings settings) ([]string, []error) {
	var failures []error

	selectors := settings.values("recipient")

	for _, filename := range settings.values("file") {
		parsed, err := ageutil.ParseRecipientsFile(filename)
		if err != nil {
			failures = append(failures, fmt.Errorf("loading recipient file %q: %w", filename, err))
			fmt.Fprintf(os.Stderr, "Unable to load recipient file %q: %v\n", filename, err)

			continue
		}

		for _, key := range parsed {
			text, err := ageutil.RecipientText(key)
			if err != nil {
				failures = append(failures, fmt.Errorf("identifying recipient from file %q: %w", filename, err))
				fmt.Fprintf(os.Stderr, "Unable to identify recipient from file %q: %v\n", filename, err)

				continue
			}

			selectors = append(selectors, text)
		}
	}

	return selectors, failures
}

func removeRecipients(recipients []recipient) (int, []error) {
	removed := 0

	var failures []error

	for _, item := range recipients {
		if err := removeIfPresent(item.stub + ExtAgeSecret); err != nil {
			failures = append(failures, fmt.Errorf("removing encrypted copy for %s: %w", item.text, err))

			continue
		}

		if err := removeIfPresent(item.filename); err != nil {
			failures = append(failures, fmt.Errorf("removing encrypted recipient copy %s: %w", item.filename, err))
			fmt.Fprintf(os.Stderr, "Unable to remove encrypted copy for %s: %v\n", item.text, err)

			continue
		}

		removed++
	}

	return removed, failures
}

func removeIfPresent(filename string) error {
	err := os.Remove(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}
