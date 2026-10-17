package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func cmdRevoke() *cli.Command {
	return command("revoke", "Remove matching GPG recipient registrations", revokeAction)
}

func revokeAction(_ context.Context, cmd *cli.Command) error {
	settings, err := parseSettings(cmd.Args().Slice(), "recipient", "file", "armor")
	if err != nil {
		return err
	}

	if len(settings.values("recipient"))+len(settings.values("file"))+len(settings.values("armor")) == 0 {
		return missing(ErrRevokeRecipientInputRequired)
	}

	selectors, failures := revokeSelectors(settings)

	redactRepo, err := openRepo()
	if err != nil {
		return err
	}

	recipients, scanFailures := readRecipients(redactRepo)
	failures = append(failures, scanFailures...)

	matched := matchingRecipients(recipients, selectors)
	if len(matched) == 0 {
		if len(failures) > 0 {
			return failed(errors.Join(failures...))
		}

		return missing(fmt.Errorf("no registered GPG recipient matches %q", selectors))
	}

	removed, removeFailures := removeRecipients(redactRepo, matched)
	failures = append(failures, removeFailures...)

	fmt.Fprintf(os.Stderr,
		"Removed %d GPG recipient registration%s. Commit and push repository changes to distribute revocation.\n",
		removed, map[bool]string{true: "", false: "s"}[removed == 1],
	)

	if len(failures) > 0 {
		return failed(errors.Join(failures...))
	}

	return nil
}

func revokeSelectors(settings settings) ([]string, []error) {
	selectors := settings.values("recipient")

	var failures []error

	for _, source := range []struct {
		setting string
		armor   bool
	}{{setting: "file"}, {setting: "armor", armor: true}} {
		for _, filename := range settings.values(source.setting) {
			keys, err := readPublicKeys(filename, source.armor)
			if err != nil {
				failures = append(failures, fmt.Errorf("loading public key file %q: %w", filename, err))
				fmt.Fprintf(os.Stderr, "Unable to load public key file %q: %v\n", filename, err)

				continue
			}

			selectors = appendFingerprints(selectors, keys)
		}
	}

	return selectors, failures
}

func removeRecipients(redactRepo *repo.Repo, recipients []recipient) (int, []error) {
	removed := 0

	var failures []error

	for _, item := range recipients {
		secretFilename := item.stub + repo.ExtSecret
		if err := removeIfPresent(redactRepo, secretFilename); err != nil {
			failures = append(failures, fmt.Errorf("removing encrypted recipient copy %s: %w", secretFilename, err))
			fmt.Fprintf(os.Stderr, "Unable to remove encrypted copy for %s: %v\n", item.fingerprintString(), err)
		}

		if err := removeIfPresent(redactRepo, item.filename); err != nil {
			failures = append(failures, fmt.Errorf("removing recipient registration %s: %w", item.filename, err))
			fmt.Fprintf(os.Stderr, "Unable to remove registration for %s: %v\n", item.fingerprintString(), err)

			continue
		}

		removed++
	}

	return removed, failures
}

func appendFingerprints(selectors []string, keys []*openpgp.Entity) []string {
	for _, key := range keys {
		selectors = append(selectors, hex.EncodeToString(key.PrimaryKey.Fingerprint))
	}

	return selectors
}

func removeIfPresent(redactRepo *repo.Repo, filename string) error {
	err := redactRepo.Workdir.Remove(filename)
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
