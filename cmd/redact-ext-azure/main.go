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
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func app() *cli.Command {
	return &cli.Command{
		Name:      "redact-ext-azure",
		Usage:     "Azure Key Vault extension for react key exchange",
		ArgsUsage: "[key=val [key=val ...]]", //nolint:goconst
		Version:   version,
		Commands: []*cli.Command{
			cmdList(),
			cmdGet(),
			cmdPut(),
		},
	}
}
