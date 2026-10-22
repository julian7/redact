package main

import (
	"context"
	"fmt"
	"os"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

var version = "SNAPSHOT"

func main() {
	if err := app().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(ext.ExitCode(err))
	}
}

func app() *cli.Command {
	return &cli.Command{
		Name:    "redact-ext-gpg",
		Usage:   "GnuPG Key Exchange extension",
		Version: version,
		Description: `Shares Redact keys with OpenPGP recipients using the repository's .redact directory.
Run commands from within the repository. Pass settings as key=value arguments.

Examples:
  redact-ext-gpg get gpgkey=alice@example.org
  redact grant --ext gpg --option recipient=alice@example.org

Configure for a repository:
  redact ext add --name gpg`,
		Action: func(_ context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() == 0 {
				return cli.ShowRootCommandHelp(cmd)
			}

			return notApplicable(fmt.Errorf("unsupported extension command %q", cmd.Args().First()))
		},
		Commands: []*cli.Command{
			cmdList(),
			cmdGet(),
			cmdPut(),
			cmdGrant(),
			cmdRevoke(),
			cmdUpdate(),
		},
	}
}

func notApplicable(err error) error { return outcome(ext.NotApplicable, err) }

func outcome(code int, err error) error {
	if err == nil {
		return nil
	}

	return &ext.Error{Code: code, Err: err}
}

func failed(err error) error { return outcome(ext.Failure, err) }

func missing(err error) error { return outcome(ext.MissingData, err) }

func command(name, usage string, action cli.ActionFunc) *cli.Command {
	return &cli.Command{
		Name:   name,
		Usage:  usage,
		Action: action,
	}
}
