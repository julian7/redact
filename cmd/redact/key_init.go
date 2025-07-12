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

func (rt *Runtime) initDo(ctx context.Context, cmd *cli.Command) error {
	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	err := rt.SetupRepo()
	if err != nil {
		return err
	}

	if err := rt.SecretKey.Load(rt.StrictPermissionChecks); err == nil {
		return fmt.Errorf("%w: %s", ErrKeyAlreadyExists, rt.SecretKey)
	}

	if err := rt.SecretKey.Generate(); err != nil {
		return fmt.Errorf("generating secret key: %w", err)
	}

	rt.Logger.Infof("New repo key created: %v", rt.SecretKey)

	if err := rt.SecretKey.Save(); err != nil {
		return fmt.Errorf("saving secret key: %w", err)
	}

	extConfig, err := ext.Load(rt.Repo)
	if extConfig != nil {
		buf := &bytes.Buffer{}
		rt.Repo.SecretKey.Export(buf)
		if err = extConfig.SaveKey(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}
