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
		Name:      "redact-ext-azure",
		Usage:     "Azure Key Vault extension for react key exchange",
		ArgsUsage: "[key=val [key=val ...]]",
		Version:   version,
		Commands: []*cli.Command{
			cmdList(),
			cmdGet(),
			cmdPut(),
		},
	}
}
