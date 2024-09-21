package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func cmdList() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "Shows Azure Key Vault configuration",
		ArgsUsage: "[key=val [key=val ...]]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := loadConfig(cmd.Args().Slice())
			if err != nil {
				return err
			}

			fmt.Printf("Azure key vault %s in secret %q\n", config.KeyvaultUrl, config.SecretName)

			return nil
		},
	}
}
