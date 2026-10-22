package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/julian7/redact/ext"

	"github.com/urfave/cli/v3"
)

var version = "SNAPSHOT"

func main() {
	if err := app().Run(context.Background(), os.Args); err != nil {
		if !errors.Is(err, ErrNotApplicable) {
			fmt.Fprintln(os.Stderr, err.Error())
		}

		os.Exit(ext.ExitCode(err))
	}
}

func app() *cli.Command {
	return &cli.Command{
		Name:    "redact-ext-azure",
		Usage:   "Azure Key Vault extension for redact key exchange",
		Version: version,
		Description: `Stores Redact keys as secrets in Azure Key Vault.
Pass vault=NAME_OR_URL and secret=NAME as key=value arguments.
Uses Azure credentials from the environment or an authenticated Azure CLI session.

Example:
  redact-ext-azure get vault=project secret=redact

Configure for a repository:
  redact ext add --name azure --option vault=project --option secret=redact`,
		Action: func(_ context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return cli.ShowRootCommandHelp(cmd)
			}

			return &ext.Error{Code: ext.NotApplicable, Err: fmt.Errorf("%w: %s", ErrNotSupported, cmd.Args().First())}
		},
		Commands: []*cli.Command{
			cmdList(),
			cmdGet(),
			cmdPut(),
			cmdNA("grant"),
			cmdNA("revoke"),
			cmdNA("update"),
		},
	}
}

func cmdNA(name string) *cli.Command {
	return &cli.Command{
		Name:        name,
		Usage:       "not available",
		Description: "this command is not available in this extension",
		Action: func(_ context.Context, _ *cli.Command) error {
			return &ext.Error{Code: ext.NotApplicable, Err: ErrNotApplicable}
		},
	}
}
