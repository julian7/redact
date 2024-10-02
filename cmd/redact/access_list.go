package main

import (
	"context"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) accessListCmd() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "List collaborators to secrets in git repo",
		Action: rt.accessListDo,
	}
}

func (rt *Runtime) accessListDo(ctx context.Context, cmd *cli.Command) error {
	_ = rt.LoadSecretKey(ctx, cmd)

	extConfig, err := ext.Load(rt.Repo)
	if extConfig != nil {
		extConfig.List()
	}

	if err != nil {
		return err
	}

	return nil
}
