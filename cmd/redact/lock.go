package main

import (
	"fmt"

	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/sdk"
	"github.com/spf13/cobra"
)

func (rt *Runtime) lockCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Locks repository",
		Long: `Lock repository

This command removes your secret key, and the filter configuration. It also
turns secret files into their unencrypted form. The git repo will behave
as like being not redact-aware. Locally modified or staged files can cause
leaking of secrets, and it's recommended to cancel all local modifications
beforehand.`,
		PreRunE: rt.LoadSecretKey,
		RunE:    rt.lockDo,
	}

	return cmd, nil
}

func (rt *Runtime) lockDo(cmd *cobra.Command, args []string) error {
	err := sdk.RemoveGitSettings(func(attr string) {
		rt.Logger.Debugf("Removing filter/diff git config of %s", attr)
	})
	if err != nil {
		return fmt.Errorf("locking repo: %w", err)
	}

	if err := rt.SecretKey.Remove(rt.Repo.Keyfile()); err != nil {
		return fmt.Errorf("locking repo: %w", err)
	}

	if err := gitutil.Renormalize(rt.Repo.Toplevel, false); err != nil {
		return err
	}

	rt.Logger.Info("Repository locked.")

	return nil
}
