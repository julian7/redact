package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

func cmdGet() *cli.Command {
	return &cli.Command{
		Name:        "put",
		Usage:       "Put secret to AWS Param Store",
		ArgsUsage:   "keyid=kms_key_id_or_alias param=/path/to/param",
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
			keyStr := string(key)

			ssmClient, err := ssmClient(ctx)
			if err != nil {
				return err
			}
			getParamRet, err := cfg.get(ctx, ssmClient)
			if err == nil {
				gotSecret := getParamRet.Parameter.Value
				if gotSecret != nil && *gotSecret == keyStr {
					return ErrAlreadyWritten
				}
			}

			_, err = cfg.put(ctx, ssmClient, keyStr)
			if err != nil {
				return fmt.Errorf("writing secret %s: %w", cfg.ParamPath, err)
			}

			return nil
		},
	}
}
