package main

import (
	"context"
	"fmt"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) extUpdateCmd() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Aliases: []string{"modify", "mod", "change"},
		Usage:   "Update extension configuration",
		Description: `Update redact extension

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
			&cli.BoolFlag{
				Name:     "nocmd",
				Aliases:  []string{"C"},
				Usage:    "removes command name",
				Value:    false,
				Required: false,
			},
			&cli.StringFlag{
				Name:     "cmd",
				Aliases:  []string{"c"},
				Usage:    "sets extension command name (use full path if it is not in $PATH)",
				Required: false,
			},
			&cli.StringMapFlag{
				Name:     "option",
				Aliases:  []string{"o"},
				Usage:    "Add configuration option",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:     "nooption",
				Aliases:  []string{"O"},
				Usage:    "Remove configuration option",
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
			name := cmd.String("name")
			extension, ok := conf.Ext(name)
			if !ok {
				return ext.ErrExtNotFound
			}
			if cmd.Bool("nocmd") {
				extension.Command = ""
			} else if newCmd := cmd.String("cmd"); newCmd != "" {
				extension.Command = newCmd
			}

			for _, item := range cmd.StringSlice("nooption") {
				delete(extension.Config, item)
			}

			for key, val := range cmd.StringMap("option") {
				extension.Config[key] = val
			}

			if err = conf.UpdateExt(name, extension); err != nil {
				return fmt.Errorf("updating extension config: %w", err)
			}

			if err = extension.List(); err != nil {
				return fmt.Errorf("extension pre-check: %w", err)
			}

			if err = conf.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			return nil
		},
	}
}
