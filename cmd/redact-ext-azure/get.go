package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func cmdGet() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "Get secrets from azure Key Vault",
		ArgsUsage:   "[key=val [key=val ...]]",
		Description: "Prints Key Vault secret to STDOUT",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := loadConfig(cmd.Args().Slice())
			if err != nil {
				return err
			}

			client, err := config.secretsClient()
			if err != nil {
				return err
			}

			gotSecret, err := client.GetSecret(ctx, config.SecretName, "", nil)
			if err != nil {
				return err
			}
			fmt.Printf(*gotSecret.Value)
			return nil
		},
	}
}
