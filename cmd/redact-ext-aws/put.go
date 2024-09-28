package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func cmdPut() *cli.Command {
	return &cli.Command{
		Name:        "get",
		Usage:       "Gett secret from AWS Param Store",
		ArgsUsage:   "keyid=kms_key_id_or_alias param=/path/to/param",
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
				return err
			}

			fmt.Print(*getParamRet.Parameter.Value)

			return nil
		},
	}
}
