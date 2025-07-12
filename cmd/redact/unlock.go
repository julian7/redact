package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) unlockCmd() *cli.Command {
	return &cli.Command{
		Name:  "unlock",
		Usage: "Unlocks repository",
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
				Name:    "ext",
				Aliases: []string{"x"},
				Usage:   "Use specific extension to read key",
				Sources: cli.EnvVars("REDACT_UNLOCK_EXT"),
			},
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
	extName := cmd.String("ext")

	if keyFile != "" && pemFile != "" {
		return fmt.Errorf("%w: --key and --exported-key are mutually exclusive", ErrOptions)
	}

	if err := rt.SetupRepo(); err != nil {
		return fmt.Errorf("building secret key: %w", err)
	}

	if extName == "" && (keyFile != "" || pemFile != "") {
		if err := rt.readSecretKey(keyFile, pemFile); err != nil {
			return err
		}
	} else {
		conf, err := ext.Load(rt.Repo)
		if err != nil {
			return err
		}
		successful := false
		for name, ext := range conf.Exts {
			if extName != "" && extName != name {
				continue
			}
			data, err := ext.LoadKey()
			if err != nil {
				continue
			}

			reader := bytes.NewReader(data)
			if err = rt.SecretKey.Import(reader); err != nil {
				return err
			}
			successful = true
			break
		}
		if !successful {
			return ErrNoSuitableKey
		}
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
