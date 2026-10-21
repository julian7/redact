package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/kx"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func cmdPut() *cli.Command {
	cmd := command("put", "Publish the supplied Key Exchange collection", putAction)
	cmd.Description = "Reads a Key Exchange collection from stdin and encrypts a copy for each registered GPG recipient."

	return cmd
}

func cmdUpdate() *cli.Command {
	cmd := command("update", "Refresh registered GPG recipients", putAction)
	cmd.Description = `Reads the current Key Exchange collection from stdin and replaces
registered recipients' encrypted copies.`

	return cmd
}

func putAction(_ context.Context, cmd *cli.Command) error {
	if _, err := parseSettings(cmd.Args().Slice()); err != nil {
		return err
	}

	redactRepo, err := openRepo()
	if err != nil {
		return err
	}

	key, err := readCollection(os.Stdin)
	if err != nil {
		return err
	}

	return updateRecipients(redactRepo, key)
}

func readCollection(reader io.Reader) (*files.SecretKey, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, failed(fmt.Errorf("reading Key Exchange collection: %w", err))
	}

	if len(data) == 0 {
		return nil, missing(ErrMissingKeyExchangeCollection)
	}

	key, err := files.NewSecretKey(nil)
	if err != nil {
		return nil, failed(err)
	}

	if err := key.Import(bytes.NewReader(data)); err != nil {
		return nil, failed(fmt.Errorf("reading Key Exchange collection: %w", err))
	}

	return key, nil
}

func updateRecipients(redactRepo *repo.Repo, key *files.SecretKey) error {
	recipients, failures := readRecipients(redactRepo)
	updated := 0

	for _, item := range recipients {
		if err := kx.SaveGPGKeyToKX(redactRepo, item.key, key.SaveTo); err != nil {
			failures = append(failures, fmt.Errorf("updating recipient %s: %w", item.fingerprintString(), err))
			fmt.Fprintf(os.Stderr, "Unable to update recipient %s: %v\n", item.fingerprintString(), err)

			continue
		}

		updated++
	}

	if updated == 0 && len(recipients) == 0 && len(failures) == 0 {
		fmt.Fprintln(os.Stderr, "No GPG recipients are registered; no Key Exchange files were changed.")
	} else if updated > 0 {
		fmt.Fprintf(
			os.Stderr,
			"Updated %d encrypted Key Exchange file%s. Commit and push repository changes to distribute them.\n",
			updated,
			map[bool]string{true: "", false: "s"}[updated == 1],
		)
	}

	if len(failures) > 0 {
		return failed(errors.Join(failures...))
	}

	return nil
}
