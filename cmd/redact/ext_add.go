package main

import (
	"context"
	"fmt"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) extAddCmd() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add extension",
		Description: `Add redact extension

A redact extension is usually following the name of "redact-ext-" + name
format. That can be overwritten with the --cmd flag.

Provide all configuration parameters required with the --option NAME=VALUE
flag (multiple parameters can be comma-separated, or multiple --option flags
can be provided).`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Usage:    "extension name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "cmd",
				Aliases:  []string{"c"},
				Usage:    "extension command name (use full path if it is not in $PATH)",
				Required: false,
			},
			&cli.StringMapFlag{
				Name:     "option",
				Aliases:  []string{"o"},
				Usage:    "Configuration option for extension",
				Required: false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := rt.SetupRepo(); err != nil {
				return err
			}
			conf, err := ext.Load(rt.Repo)
			if err != nil {
				return err
			}
			ext := ext.Ext{
				Command: cmd.String("cmd"),
				Config:  cmd.StringMap("option"),
			}
			if err = conf.AddExt(cmd.String("name"), ext); err != nil {
				return fmt.Errorf("adding extension: %w", err)
			}
			newExt, ok := conf.Ext(cmd.String("name"))
			if !ok {
				return ErrExtensionNotFound
			}
			if err = newExt.List(); err != nil {
				return fmt.Errorf("extension pre-check: %w", err)
			}
			if err = conf.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			return nil
		},
	}
}
