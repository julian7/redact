package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

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
		Flags: append(extensionFlags(), []cli.Flag{
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
		}...),
		Commands: []*cli.Command{
			rt.unlockGpgCmd(),
		},
	}
}

func (rt *Runtime) unlockDo(_ context.Context, cmd *cli.Command) error {
	keyFile := cmd.String("key")
	pemFile := cmd.String("exported-key")
	extName := cmd.String("ext")

	extensionSettings := extName != "" || cmd.String(extCommandFlag) != "" || len(cmd.StringMap(extOptionFlag)) > 0
	if (keyFile != "" && pemFile != "") || ((keyFile != "" || pemFile != "") && extensionSettings) {
		return fmt.Errorf("%w: key files and extension selection/settings are mutually exclusive", ErrOptions)
	}

	if err := rt.SetupRepo(); err != nil {
		return fmt.Errorf("building secret key: %w", err)
	}

	if extName == "" && (keyFile != "" || pemFile != "") {
		if err := rt.readSecretKey(keyFile, pemFile); err != nil {
			return err
		}
	} else {
		err := rt.loadSecretKeyFromExt(extName, cmd.String(extCommandFlag), cmd.StringMap(extOptionFlag))
		if errors.Is(err, ErrNoSuitableKey) && extName == "" && cmd.String(extCommandFlag) == "" && !gpgExtensionMode() {
			err = rt.loadSecretKeyFromGPG(cmd.StringMap(extOptionFlag)["gpgkey"])
			if errors.Is(err, io.EOF) {
				return nil
			}
		}

		if err != nil {
			return err
		}
	}

	return rt.finishUnlock()
}

func (rt *Runtime) finishUnlock() error {
	if err := rt.Save(); err != nil {
		return err
	}

	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	if err := rt.ForceReencrypt(false, func(err error) {
		rt.Warn(err.Error())
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
		err = rt.Read(f)
	} else {
		err = rt.Import(f)
	}

	return err
}

func (rt *Runtime) loadSecretKeyFromExt(name, command string, options map[string]string) error {
	conf, err := ext.Load(rt.Repo)
	if err != nil {
		return fmt.Errorf("loading extension config: %w", err)
	}

	items, err := conf.Select(name, command, options)
	if err != nil {
		return err
	}

	if name == "" && !gpgExtensionMode() {
		items = withoutGPGExtension(items)
	}

	for _, item := range items {
		data, err := item.LoadKey()
		if ext.ExitCode(err) == ext.NotApplicable && name == "" {
			continue
		}

		if err != nil {
			return err
		}

		if err := rt.Import(bytes.NewReader(data)); err != nil {
			return fmt.Errorf("importing extension key: %w", &ext.Error{Code: ext.Failure, Err: err})
		}

		return nil
	}

	return ErrNoSuitableKey
}
