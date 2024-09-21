package main

import (
	"context"

	"github.com/urfave/cli/v3"
)

func (rt *Runtime) keyCmd() *cli.Command {
	return &cli.Command{
		Name:  "key",
		Usage: "Key commands",
		Description: `Secret Key management

Key commands let you manage secret key. The secret key consists of multiple
encryption keys, of which the last one is considered to be the active one.
All the other keys are kept for archival purposes.

A simple "redact key" will return the secret key location, and its active
key's epoch and signature. For more detailed look, please see
"redact key list".`,
		Action: rt.keyDo,
		Commands: []*cli.Command{
			rt.keyGenerateCmd(),
			rt.keySaveCmd(),
			rt.keyInitCmd(),
			rt.keyListCmd(),
			rt.keyExportCmd(),
		},
	}
}

func (rt *Runtime) keyDo(ctx context.Context, cmd *cli.Command) error {
	if err := rt.LoadSecretKey(ctx, cmd); err != nil {
		return err
	}

	rt.Logger.Infof("repo key: %v", rt.SecretKey)

	return nil
}
