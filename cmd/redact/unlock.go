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

Alternatively, a secret key can be provided. This allows unlocking the
repository where other ways are not available.`,
		RunE: rt.unlockDo,
	}

	flags := cmd.Flags()
	flags.StringP("key", "k", "", "Use specific secret key")

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
	var err error

	keyFile := rt.Viper.GetString("key")
	if keyFile == "" {
		_ = cmd.Usage()

		return errors.New("--key is required")
	}

	if err := rt.SetupRepo(); err != nil {
		return fmt.Errorf("building secret key: %w", err)
	}

	err = rt.loadKeyFromFile(keyFile)
	if err != nil {
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

	fmt.Println("Key is unlocked.")

	return nil
}
