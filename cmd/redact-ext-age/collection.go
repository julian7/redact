package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

func cmdPut() *cli.Command {
	cmd := command("put", "Publish the supplied Key Exchange collection", putAction)
	cmd.Description = "Reads a Key Exchange collection from stdin and encrypts a copy for each registered age recipient."

	return cmd
}

func cmdUpdate() *cli.Command {
	cmd := command("update", "Refresh registered age recipients", putAction)
	cmd.Description = `Reads the current Key Exchange collection from stdin and replaces
registered recipients' encrypted copies.`

	return cmd
}

func putAction(_ context.Context, cmd *cli.Command) error {
	if _, err := parseSettings(cmd.Args().Slice()); err != nil {
		return err
	}

	collection, err := readCollection(os.Stdin)
	if err != nil {
		return err
	}

	return updateRecipients(collection)
}

// Keep payloads opaque; core parses the recovered collection.
func readCollection(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, failed(fmt.Errorf("reading Key Exchange collection: %w", err))
	}

	if len(data) == 0 {
		return nil, missing(ErrMissingKeyExchangeCollection)
	}

	return data, nil
}

func updateRecipients(collection []byte) error {
	recipients, failures := readRecipients()

	updated := 0

	for _, item := range recipients {
		if err := saveAgeKeyToKX(item.key, collection); err != nil {
			failures = append(failures, fmt.Errorf("updating recipient %s: %w", item.text, err))
			fmt.Fprintf(os.Stderr, "Unable to update recipient %s: %v\n", item.text, err)

			continue
		}

		updated++
	}

	if updated == 0 && len(recipients) == 0 && len(failures) == 0 {
		fmt.Fprintln(os.Stderr, "No age recipients are registered; no Key Exchange files were changed.")
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
