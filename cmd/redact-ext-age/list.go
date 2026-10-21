package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/julian7/redact/ageutil"
)

func cmdList() *cli.Command {
	cmd := command("list", "List registered age recipients", func(_ context.Context, cmd *cli.Command) error {
		if _, err := parseSettings(cmd.Args().Slice()); err != nil {
			return err
		}

		recipients, failures := readRecipients()

		for _, item := range recipients {
			if err := ageutil.PrintRecipientTo(os.Stdout, item.text); err != nil {
				failures = append(failures, fmt.Errorf("printing recipient: %w", err))
			}
		}

		if len(failures) > 0 {
			return failed(errors.Join(failures...))
		}

		return nil
	})
	cmd.Description = "Prints the age recipients registered in the repository."

	return cmd
}
