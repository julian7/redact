package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func (rt *Runtime) keyGenerateCmd() *cli.Command {
	return &cli.Command{
		Name:        "generate",
		Aliases:     []string{"gen", "g"},
		Usage:       "Generates redact key and updates Key Exchange collections",
		Description: "Saves a new key epoch and invokes extension update. Commit and push any exchange changes.",
		Before:      rt.LoadSecretKey,
		Flags:       extensionFlags(),
		Action:      rt.generateDo,
	}
}

func (rt *Runtime) generateDo(_ context.Context, cmd *cli.Command) error {
	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	if err := rt.Generate(); err != nil {
		return fmt.Errorf("generating secret key: %w", err)
	}

	rt.Infof("New repo key created: %v", rt.SecretKey)

	if err := rt.Save(); err != nil {
		return fmt.Errorf("saving secret key: %w", err)
	}

	return rt.distributeKey(cmd, "update")
}
