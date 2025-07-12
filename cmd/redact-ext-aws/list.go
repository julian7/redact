package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func cmdList() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "Shows AWS param store configuration",
		ArgsUsage: "keyid=kms_key_id_or_alias param=/path/to/param",
		Action: func(_ context.Context, cmd *cli.Command) error {
			config, err := loadConfig(cmd.Args().Slice())
			if err != nil {
				return err
			}

			fmt.Printf("AWS Param Store key %q with key %q\n", config.ParamPath, config.KeyID)

			return nil
		},
	}
}
