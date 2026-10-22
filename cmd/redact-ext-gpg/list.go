package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func cmdList() *cli.Command {
	cmd := command("list", "List registered GPG recipients", func(_ context.Context, cmd *cli.Command) error {
		if _, err := parseSettings(cmd.Args().Slice()); err != nil {
			return err
		}

		redactRepo, err := openRepo()
		if err != nil {
			return err
		}

		recipients, failures := readRecipients(redactRepo)
		for _, item := range recipients {
			if err := gpgutil.PrintKeyTo(os.Stdout, item.key); err != nil {
				failures = append(failures, fmt.Errorf("printing recipient: %w", err))
			}
		}

		if len(failures) > 0 {
			return failed(errors.Join(failures...))
		}

		return nil
	})
	cmd.Description = "Prints the identities and fingerprints of GPG recipients registered in the repository."

	return cmd
}

func openRepo() (*repo.Repo, error) {
	redactRepo := &repo.Repo{}
	if err := redactRepo.SetupRepo(); err != nil {
		return nil, failed(fmt.Errorf("detecting repository: %w", err))
	}

	return redactRepo, nil
}
