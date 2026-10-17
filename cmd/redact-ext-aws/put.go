package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/julian7/redact/ext"

	"github.com/urfave/cli/v3"
)

func cmdPut() *cli.Command {
	return &cli.Command{
		Name:        "put",
		Aliases:     []string{"update"},
		Usage:       "Put secret to AWS Param Store",
		ArgsUsage:   "keyid=kms_key_id_or_alias param=/path/to/param", //nolint:goconst
		Description: "Reads secret from STDIN and writes to Param Store",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := loadConfig(cmd.Args().Slice())
			if err != nil {
				return err
			}

			key, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			if len(key) == 0 {
				return &ext.Error{Code: ext.MissingData, Err: errors.New("missing exported key on stdin")}
			}

			keyStr := string(key)

			ssmClient, err := ssmClient(ctx)
			if err != nil {
				return err
			}

			return cfg.save(ctx, ssmClient, keyStr)
		},
	}
}

type parameterStore interface {
	GetParameter(context.Context, *ssm.GetParameterInput, ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
	PutParameter(context.Context, *ssm.PutParameterInput, ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
}

func (config *Config) save(ctx context.Context, client parameterStore, key string) error {
	got, err := config.get(ctx, client)
	if err == nil && got.Parameter != nil && got.Parameter.Value != nil && *got.Parameter.Value == key {
		return nil
	}

	if _, err := config.put(ctx, client, key); err != nil {
		return fmt.Errorf("writing secret %s: %w", config.ParamPath, err)
	}

	return nil
}
