package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var version = "SNAPSHOT"

func main() {
	if err := app().Run(context.Background(), os.Args); err != nil {
		fmt.Println(err.Error())
	}
}

func app() *cli.Command {
	return &cli.Command{
		Name:      "redact-ext-aws",
		Usage:     "AWS KMS+param store extension for react key exchange",
		ArgsUsage: "keyid=kms_key_id_or_alias param=/path/to/param",
		Version:   version,
		Commands: []*cli.Command{
			cmdList(),
			cmdGet(),
			cmdPut(),
		},
	}
}
