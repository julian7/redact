package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func cmdPut() *cli.Command {
	return &cli.Command{
		Name:        "put",
		Usage:       "Saves secret in GnuPG message",
		Description: "Reads secret from STDIN and writes to GnuPG message",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return nil
		},
	}
}
