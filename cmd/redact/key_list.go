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

func (rt *Runtime) listDo(_ context.Context, _ *cli.Command) error {
	rt.Infof("repo key: %v", rt.SecretKey)

	return files.EachKey(rt.Keys, func(_ uint32, key files.KeyHandler) error {
		rt.Infof(" - %s", key)

		return nil
	})
}
