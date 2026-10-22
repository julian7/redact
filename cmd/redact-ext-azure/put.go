package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/julian7/redact/ext"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/urfave/cli/v3"
)

func cmdPut() *cli.Command {
	return &cli.Command{
		Name:        "put",
		Aliases:     []string{"update"},
		Usage:       "Put secret to Azure Key Vault",
		ArgsUsage:   "vault=NAME_OR_URL secret=NAME", //nolint:goconst
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

			if len(key) == 0 {
				return &ext.Error{Code: ext.MissingData, Err: ErrMissingExportedKey}
			}

			client, err := config.secretsClient()
			if err != nil {
				return err
			}

			return config.save(ctx, client, string(key))
		},
	}
}

type secretStore interface {
	GetSecret(context.Context, string, string, *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
	SetSecret(
		context.Context, string, azsecrets.SetSecretParameters, *azsecrets.SetSecretOptions,
	) (azsecrets.SetSecretResponse, error)
}

func (config *Config) save(ctx context.Context, client secretStore, key string) error {
	got, err := client.GetSecret(ctx, config.SecretName, "", nil)
	if err == nil && got.Value != nil && *got.Value == key {
		return nil
	}

	contentType := "text/plain"

	params := azsecrets.SetSecretParameters{ContentType: &contentType, Value: &key}
	if _, err := client.SetSecret(ctx, config.SecretName, params, nil); err != nil {
		return fmt.Errorf("cannot write secret %s: %w", config.SecretName, err)
	}

	return nil
}
