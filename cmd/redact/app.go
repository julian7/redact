package main

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

var (
	version = "SNAPSHOT"
)

func (rt *Runtime) app() *cli.App {
	return &cli.App{
		Name:      "redact",
		Usage:     "encrypts files in a git repository",
		ArgsUsage: " ",
		Description: `redact - keep secrets in a git repository

This application uses gitattributes(5) to encrypt and decrypt files behind
the scenes (see filter and diff attributes). This process requires a secret
key, what you can generate with "redact init" command. The secret key can
hold multiple key versions, supporting key rotation and retrieval of old
secrets.

Secret keys can be distributed inside the repository in the key exchange
directory ($GIT_DIR/.redact), encrypted by contributors' OpenPGP keys.
Contributors can unlock the repo by running "redact unlock".

To make files to be managed by adding the file pattern into a .gitattributes
file like this:

	*.secret.txt filter=redact diff=redact

The subsequent "git add" command will encrypt files matching this pattern.`,
		Before:  rt.GlobalConfig,
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "verbosity",
				Aliases: []string{"V"},
				Value:   "info",
				Usage:   "Verbosity (possible values: debug, info, warn, error, fatal)",
				EnvVars: []string{"REDACT_LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    "logfile",
				EnvVars: []string{"REDACT_LOG_FILE"},
				Usage:   "Write logs to `FILE` (standard output by default)",
			},
			&cli.BoolFlag{
				Name:    "strict-permissions",
				EnvVars: []string{"REDACT_STRICT"},
				Value:   true,
				Usage:   "enforce file permission checks; use with caution!",
			},
		},
		Commands: []*cli.Command{
			rt.accessCmd(),
			rt.gitCmd(),
			rt.initCmd(),
			rt.keyCmd(),
			rt.lockCmd(),
			rt.statusCmd(),
			rt.unlockCmd(),
		},
	}
}

func (rt *Runtime) GlobalConfig(ctx *cli.Context) error {
	if logFile := ctx.Path("logfile"); logFile != "" {
		writer, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			rt.Logger.Warnf("cannot open log file: %v", err)
		} else {
			rt.Logger.SetOutput(writer)
		}
	}

	rt.setLogLevel(strings.ToLower(ctx.String("verbosity")))

	rt.StrictPermissionChecks = ctx.Bool("strict-permissions")

	return nil
}

func (rt *Runtime) setLogLevel(level string) {
	err := rt.Logger.SetLevelFromString(level)
	if err != nil {
		rt.Logger.Warnf("cannot set log level: %v", err)

		return
	}

	rt.Logger.Debugf("Setting log level to %s", level)
}
