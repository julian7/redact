package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/julian7/redact/ext"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/urfave/cli/v3"
)

func cmdGet() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "Get secret from AWS Param Store",
		ArgsUsage:   "keyid=kms_key_id_or_alias param=/path/to/param", //nolint:goconst
		Description: "Prints Param Store secret to STDOUT",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := loadConfig(cmd.Args().Slice())
			if err != nil {
				return err
			}

			ssmClient, err := ssmClient(ctx)
			if err != nil {
				return err
			}

			getParamRet, err := cfg.get(ctx, ssmClient)
			if err != nil {
				var missing *types.ParameterNotFound
				if errors.As(err, &missing) {
					return &ext.Error{Code: ext.MissingData, Err: err}
				}

				return err
			}

			if getParamRet.Parameter == nil || getParamRet.Parameter.Value == nil || *getParamRet.Parameter.Value == "" {
				return &ext.Error{Code: ext.MissingData, Err: ErrNoExportedKey}
			}

			fmt.Print(*getParamRet.Parameter.Value)

			return nil
		},
	}
}
