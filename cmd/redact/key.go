package main

import (
	"github.com/urfave/cli/v2"
)

func (rt *Runtime) keyCmd() *cli.Command {
	return &cli.Command{
		Name:      "key",
		Usage:     "Key commands",
		ArgsUsage: " ",
		Description: `Secret Key management

Key commands let you manage secret key. The secret key consists of multiple
encryption keys, of which the last one is considered to be the active one.
All the other keys are kept for archival purposes.

A simple "redact key" will return the secret key location, and its active
key's epoch and signature. For more detailed look, please see
"redact key list".`,
		Action: rt.keyDo,
		Subcommands: []*cli.Command{
			rt.keyGenerateCmd(),
			rt.keyInitCmd(),
			rt.keyListCmd(),
			rt.keyExportCmd(),
		},
	}
}

func (rt *Runtime) keyDo(ctx *cli.Context) error {
	err := rt.LoadSecretKey(ctx)
	if err != nil {
		return err
	}

	rt.Logger.Infof("repo key: %v", rt.SecretKey)

	return nil
}
