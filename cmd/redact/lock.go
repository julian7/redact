package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func (rt *Runtime) lockCmd() *cli.Command {
	return &cli.Command{
		Name:  "lock",
		Usage: "Locks repository",
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

func (rt *Runtime) lockDo(_ context.Context, _ *cli.Command) error {
	err := rt.RemoveGitSettings(func(attr string) {
		rt.Debugf("Removing filter/diff git config of %s", attr)
	})
	if err != nil {
		return fmt.Errorf("locking repo: %w", err)
	}

	if err := rt.Remove(); err != nil {
		return fmt.Errorf("locking repo: %w", err)
	}

	if err := rt.ForceReencrypt(false, func(err error) {
		rt.Warn(err.Error())
	}); err != nil {
		return err
	}

	rt.Info("Repository locked.")

	return nil
}
