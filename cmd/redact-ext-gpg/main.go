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
		Name:      "redact-ext-gpg",
		Usage:     "GnuPG extension for react key exchange",
		ArgsUsage: "",
		Version:   version,
		Commands: []*cli.Command{
			cmdList(),
			cmdGet(),
			cmdPut(),
		},
	}
}
