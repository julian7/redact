package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
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
		Name:   "init",
		Usage:  "Generates initial redact key (alias of redact key init)",
		Action: rt.initDo,
	}
}

func (rt *Runtime) initDo(_ context.Context, _ *cli.Command) error {
	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	err := rt.SetupRepo()
	if err != nil {
		return err
	}

	if err := rt.Load(rt.StrictPermissionChecks); err == nil {
		return fmt.Errorf("%w: %s", ErrKeyAlreadyExists, rt.SecretKey)
	}

	if err := rt.Generate(); err != nil {
		return fmt.Errorf("generating secret key: %w", err)
	}

	rt.Infof("New repo key created: %v", rt.SecretKey)

	if err := rt.Save(); err != nil {
		return fmt.Errorf("saving secret key: %w", err)
	}

	extConfig, err := ext.Load(rt.Repo)
	if err != nil {
		return fmt.Errorf("loading extension config: %w", err)
	}

	if extConfig != nil {
		buf := &bytes.Buffer{}
		if err := rt.Export(buf); err != nil {
			return err
		}

		if err := extConfig.SaveKey(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}
