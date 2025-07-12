package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/urfave/cli/v3"
)

func cmdPut() *cli.Command {
	return &cli.Command{
		Name:        "put",
		Usage:       "Put secret to Azure Key Vault",
		ArgsUsage:   "[key=val [key=val ...]]",
		Description: "Reads secret from STDIN and writes to Key Vault",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := loadConfig(cmd.Args().Slice())
			if err != nil {
				return err
			}

			key, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			client, err := config.secretsClient()
			if err != nil {
				return err
			}

			contentType := "text/plain"
			gotSecret, err := client.GetSecret(ctx, config.SecretName, "", nil)
			if err != nil && *gotSecret.Value == string(key) {
				return ErrAlreadyWritten
			}
			params := azsecrets.SetSecretParameters{
				ContentType: &contentType,
				Value:       to.Ptr(string(key)),
			}
			_, err = client.SetSecret(ctx, config.SecretName, params, nil)
			if err != nil {
				return fmt.Errorf("cannot write secret %s: %w", config.SecretName, err)
			}

			return nil
		},
	}
}
