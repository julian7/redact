package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func (rt *Runtime) unlockCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "unlock",
		Args:  cobra.NoArgs,
		Short: "Unlocks repository",
		Long: `Unlock repository

This group of commands able to unlock a repository, or to obtain a new version
of the secret key.

Use a subcommand for different ways of unlocking the repository from key
exchange.

Alternatively, a secret key file can be provided. This allows unlocking the
repository where other ways are not available. Providing '-' reads the key
from standard input.`,
		RunE: rt.unlockDo,
	}

	flags := cmd.Flags()
	flags.StringP("key", "k", "", "Use specific raw secret key file")
	flags.StringP("exported-key", "e", "", "Use specific exported secret key file")

	if err := rt.RegisterFlags("", flags); err != nil {
		return nil, err
	}

	subcommands := []cmdFactory{
		rt.unlockGpgCmd,
	}

	if err := rt.AddCmdTo(cmd, subcommands); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) unlockDo(cmd *cobra.Command, args []string) error {
	keyFile := rt.Viper.GetString("key")
	pemFile := rt.Viper.GetString("exported-key")

	if keyFile == "" && pemFile == "" {
		_ = cmd.Usage()

		return errors.New("--key or --exported-key is required")
	}

	if keyFile != "" && pemFile != "" {
		_ = cmd.Usage()

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
