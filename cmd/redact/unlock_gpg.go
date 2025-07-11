package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/julian7/redact/gpgutil"
	"github.com/julian7/redact/kx"
	"github.com/julian7/redact/repo"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) unlockGpgCmd() *cli.Command {
	return &cli.Command{
		Name:  "gpg",
		Usage: "Unlocks repository",
		Description: `Unlock repository

This command unlocks the repository using a GPG key.

By default, it detects your GnuPG keys by running gpg -K, and tries to match
them to the available encrypted keys in the key exchange directory. This
process won't make decisions for you, if you have multiple keys available. In
this case, you have to provide the appropriate key with the --gpgkey option.`,
		Action: rt.unlockGpgDo,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "gpgkey",
				Aliases: []string{"k"},
				Usage:   "Use specific GPG key",
				Sources: cli.EnvVars("REDACT_UNLOCK_GPG_KEY"),
			},
		},
	}
}

func (rt *Runtime) unlockGpgDo(_ context.Context, cmd *cli.Command) error {
	var err error

	if err := rt.SetupRepo(); err != nil {
		return fmt.Errorf("building secret key: %w", err)
	}

	var key string

	key, err = rt.loadKeyFromGPG(cmd.String("gpgkey"))
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		if key != "" {
			fmt.Printf("Hint: try to unlock by hand:\n\n")
			fmt.Printf(
				"  gpg -o key -u %s -d .redact/%s.key\n",
				key,
				key,
			)
			fmt.Println("  redact unlock key")
			fmt.Println("  rm key")
		}

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

func (rt *Runtime) selectKey(keyname string) (*[]byte, error) {
	keys, warns, err := gpgutil.GetSecretKeys(keyname)
	if err != nil {
		return nil, err
	}

	if len(warns) > 0 {
		for _, item := range warns {
			rt.Warn(item)
		}
	}

	availableKeys := make([]int, 0, len(keys))

	for idx, key := range keys {
		stub, err := rt.GetExchangeFilenameStubFor(key, rt.Logger)
		if err != nil {
			rt.Warnf("cannot get exchange filename for %x: %v", key, err)

			continue
		}

		secretKeyFilename := repo.ExchangeSecretKeyFile(stub)

		st, err := rt.Workdir.Stat(secretKeyFilename)
		if err != nil || st.IsDir() {
			continue
		}

		availableKeys = append(availableKeys, idx)
	}

	if len(availableKeys) > 1 {
		fmt.Println("Multiple keys found. Please specify one:")

		for _, idx := range availableKeys {
			pubKey, err := kx.LoadGPGPubkeysFromKX(rt.Repo, keys[idx])
			if err != nil {
				rt.Warnf("%v", err)

				continue
			}

			for _, entity := range pubKey {
				gpgutil.PrintKey(entity)
			}
		}

		return nil, io.EOF
	}

	if len(availableKeys) < 1 {
		return nil, ErrNoSuitableKey
	}

	return &keys[availableKeys[0]], nil
}

func (rt *Runtime) loadKeyFromGPG(gpgkey string) (string, error) {
	key, err := rt.selectKey(gpgkey)
	if err != nil {
		return "", err
	}

	reader, err := kx.SecretKeyFromExchange(rt.Repo, *key)
	if err != nil {
		return "", err
	}

	defer reader.Close()

	return fmt.Sprintf("%x", *key), kx.LoadSecretKeyFromReader(rt.SecretKey, reader)
}
