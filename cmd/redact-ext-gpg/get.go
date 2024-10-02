package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func cmdGet() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "Get secret from GPG key",
		Description: "Prints secret from GPG message to STDOUT",
		Action: func(ctx context.Context, cmd *cli.Command) error {

			return nil
		},
	}
}
