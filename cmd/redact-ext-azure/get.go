package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/julian7/redact/ext"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/urfave/cli/v3"
)

func cmdGet() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "Get secrets from azure Key Vault",
		ArgsUsage:   "[key=val [key=val ...]]", //nolint:goconst
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
				var response *azcore.ResponseError
				if errors.As(err, &response) && response.StatusCode == 404 {
					return &ext.Error{Code: ext.MissingData, Err: err}
				}

				return err
			}

			if gotSecret.Value == nil || *gotSecret.Value == "" {
				return &ext.Error{Code: ext.MissingData, Err: errors.New("secret contains no exported key")}
			}

			fmt.Print(*gotSecret.Value)

			return nil
		},
	}
}
