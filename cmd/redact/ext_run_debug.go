//go:build debug

package main

import (
	"context"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) extRunCmd() *cli.Command {
	return &cli.Command{
		Name:        "run",
		Usage:       "Run extension commands",
		Description: `Run commands in redact extensions`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := rt.SetupRepo(); err != nil {
				return err
			}
			conf, err := ext.Load(rt.Repo)
			if err != nil {
				return err
			}
			for _, ext := range conf.Exts {
				ext.Exec(cmd.Args().First())
			}
			return nil
		},
	}
}
