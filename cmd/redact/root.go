package main

import (
	"github.com/spf13/cobra"
)

const configName = ".redact"

var (
	version = "SNAPSHOT"
)

func (rt *Runtime) rootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "redact",
		Short: "encrypts files in a git repository",
		Long: `redact - keep secrets in a git repository

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
		PersistentPreRun: rt.SetupLogging,
		Version:          version,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&rt.Config, "config", "", "config file (default: ~/"+configName+".yaml)")
	flags.StringP("verbosity", "v", "info", "Verbosity (possible values: debug, info, warn, error, fatal)")
	flags.String("logfile", "", "log file (empty for standard out)")
	flags.BoolVar(
		&rt.StrictPermissionChecks,
		"strict-permissions",
		true,
		"enforce file permission checks; use with caution!",
	)

	if err := rt.RegisterFlags("", flags); err != nil {
		return nil, err
	}

	subcommands := []cmdFactory{
		rt.accessCmd,
		rt.gitCmd,
		rt.initCmd,
		rt.keyCmd,
		rt.lockCmd,
		rt.statusCmd,
		rt.unlockCmd,
	}

	if err := rt.AddCmdTo(cmd, subcommands); err != nil {
		return nil, err
	}

	return cmd, nil
}
