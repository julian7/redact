package main

import (
	"context"
	"fmt"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) extRemoveCmd() *cli.Command {
	return &cli.Command{
		Name:        "remove",
		Aliases:     []string{"delete", "uninstall", "del", "rm"},
		Usage:       "Delete extension",
		Description: `Delete redact extension`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Usage:    "extension name",
				Required: true,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			if err := rt.SetupRepo(); err != nil {
				return err
			}
			conf, err := ext.Load(rt.Repo)
			if err != nil {
				return err
			}
			conf.DelExt(cmd.String("name"))
			if err = conf.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			return nil
		},
	}
}
