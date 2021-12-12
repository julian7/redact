package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (rt *Runtime) keyInitCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generates initial redact key",
		RunE:  rt.initDo,
	}

	return cmd, nil
}

func (rt *Runtime) initCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generates initial redact key (alias of redact key init)",
		RunE:  rt.initDo,
	}

	return cmd, nil
}

func (rt *Runtime) initDo(cmd *cobra.Command, args []string) error {
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
