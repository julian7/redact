package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func (rt *Runtime) keyInitCmd() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Generates initial redact key",
		ArgsUsage: " ",
		Action:    rt.initDo,
	}
}

func (rt *Runtime) initCmd() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Generates initial redact key (alias of redact key init)",
		ArgsUsage: " ",
		Action:    rt.initDo,
	}
}

func (rt *Runtime) initDo(_ *cli.Context) error {
	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	err := rt.SetupRepo()
	if err != nil {
		return err
	}

	if err := rt.SecretKey.Load(rt.StrictPermissionChecks); err == nil {
		return fmt.Errorf("repo already has secret key: %s", rt.SecretKey)
	}

	if err := rt.SecretKey.Generate(); err != nil {
		return fmt.Errorf("generating secret key: %w", err)
	}

	rt.Logger.Infof("New repo key created: %v", rt.SecretKey)

	if err := rt.SecretKey.Save(); err != nil {
		return fmt.Errorf("saving secret key: %w", err)
	}

	return nil
}
