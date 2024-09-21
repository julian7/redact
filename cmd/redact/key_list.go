package main

import (
	"context"

	"github.com/julian7/redact/files"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) keyListCmd() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "Lists redact keys",
		Before: rt.LoadSecretKey,
		Action: rt.listDo,
	}
}

func (rt *Runtime) listDo(ctx context.Context, cmd *cli.Command) error {
	rt.Logger.Infof("repo key: %v", rt.SecretKey)

	return files.EachKey(rt.SecretKey.Keys, func(_ uint32, key files.KeyHandler) error {
		rt.Logger.Infof(" - %s", key)

		return nil
	})
}
