package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"
)

func (rt *Runtime) unlockCmd() *cli.Command {
	return &cli.Command{
		Name:      "unlock",
		Usage:     "Unlocks repository",
		ArgsUsage: " ",
		Description: `Unlock repository

This group of commands able to unlock a repository, or to obtain a new version
of the secret key.

Use a subcommand for different ways of unlocking the repository from key
exchange.

Alternatively, a secret key file can be provided. This allows unlocking the
repository where other ways are not available. Providing '-' reads the key
from standard input.`,
		Action: rt.unlockDo,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
				Usage:   "Use specific raw secret key file",
				Sources: cli.EnvVars("REDACT_UNLOCK_KEY"),
			},
			&cli.StringFlag{
				Name:    "exported-key",
				Aliases: []string{"e"},
				Usage:   "Use specific exported secret key file",
				Sources: cli.EnvVars("REDACT_UNLOCK_EXPORTED_KEY"),
			},
		},
		Commands: []*cli.Command{
			rt.unlockGpgCmd(),
		},
	}
}

func (rt *Runtime) unlockDo(ctx context.Context, cmd *cli.Command) error {
	keyFile := cmd.String("key")
	pemFile := cmd.String("exported-key")

	if keyFile == "" && pemFile == "" {
		return errors.New("--key or --exported-key is required")
	}

	if keyFile != "" && pemFile != "" {
		return errors.New("--key and --exported-key are mutully exclusive")
	}

	if err := rt.SetupRepo(); err != nil {
		return fmt.Errorf("building secret key: %w", err)
	}

	if err := rt.readSecretKey(keyFile, pemFile); err != nil {
		return err
	}

	if err := rt.SecretKey.Save(); err != nil {
		return err
	}

	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	if err := rt.Repo.ForceReencrypt(false, func(err error) {
		rt.Logger.Warn(err.Error())
	}); err != nil {
		return err
	}

	fmt.Println("Repo is unlocked.")

	return nil
}

func (rt *Runtime) readSecretKey(keyFile, pemFile string) error {
	var fname string
	if keyFile != "" {
		fname = keyFile
	} else {
		fname = pemFile
	}

	f, err := openFileToRead(fname)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", fname, err)
	}
	defer f.Close()

	if keyFile != "" {
		err = rt.SecretKey.Read(f)
	} else {
		err = rt.SecretKey.Import(f)
	}

	return err
}
