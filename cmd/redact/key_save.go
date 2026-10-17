package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func (rt *Runtime) keySaveCmd() *cli.Command {
	return &cli.Command{
		Name:    "save",
		Aliases: []string{"s"},
		Usage:   "Saves redact key in Key Exchange",
		Before:  rt.LoadSecretKey,
		Flags:   extensionFlags(),
		Action:  rt.keySaveDo,
	}
}

func (rt *Runtime) keySaveDo(_ context.Context, cmd *cli.Command) error {
	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	if err := rt.Save(); err != nil {
		return fmt.Errorf("saving secret key: %w", err)
	}

	return rt.distributeKey(cmd, "put")
}
