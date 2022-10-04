package main

import (
	"github.com/julian7/redact/files"
	"github.com/urfave/cli/v2"
)

func (rt *Runtime) keyListCmd() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "Lists redact keys",
		ArgsUsage: " ",
		Before:    rt.LoadSecretKey,
		Action:    rt.listDo,
	}
}

func (rt *Runtime) listDo(ctx *cli.Context) error {
	rt.Logger.Infof("repo key: %v", rt.SecretKey)

	return files.EachKey(rt.SecretKey.Keys, func(idx uint32, key files.KeyHandler) error {
		rt.Logger.Infof(" - %s", key)

		return nil
	})
}
