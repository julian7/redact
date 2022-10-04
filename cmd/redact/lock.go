package main

import (
	"fmt"

	"github.com/julian7/redact/sdk"
	"github.com/urfave/cli/v2"
)

func (rt *Runtime) lockCmd() *cli.Command {
	return &cli.Command{
		Name:      "lock",
		Usage:     "Locks repository",
		ArgsUsage: " ",
		Description: `Lock repository

This command removes your secret key, and the filter configuration. It also
turns secret files into their unencrypted form. The git repo will behave
as like being not redact-aware. Locally modified or staged files can cause
leaking of secrets, and it's recommended to cancel all local modifications
beforehand.`,
		Before: rt.LoadSecretKey,
		Action: rt.lockDo,
	}
}

func (rt *Runtime) lockDo(ctx *cli.Context) error {
	err := sdk.RemoveGitSettings(func(attr string) {
		rt.Logger.Debugf("Removing filter/diff git config of %s", attr)
	})
	if err != nil {
		return fmt.Errorf("locking repo: %w", err)
	}

	if err := rt.Repo.Remove(rt.Repo.Keyfile()); err != nil {
		return fmt.Errorf("locking repo: %w", err)
	}

	if err := rt.Repo.ForceReencrypt(false, func(err error) {
		rt.Logger.Warn(err.Error())
	}); err != nil {
		return err
	}

	rt.Logger.Info("Repository locked.")

	return nil
}
