package main

import (
	"fmt"

	"github.com/julian7/redact/files"
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

	masterkey, err := files.NewMasterKey(rt.Logger)
	if err != nil {
		return err
	}

	if err := masterkey.Load(rt.StrictPermissionChecks); err == nil {
		return fmt.Errorf("repo already has master key: %s", masterkey)
	}

	if err := masterkey.Generate(); err != nil {
		return fmt.Errorf("generating master key: %w", err)
	}

	rt.Logger.Infof("New repo key created: %v", masterkey)

	if err := masterkey.Save(); err != nil {
		return fmt.Errorf("saving master key: %w", err)
	}

	return nil
}
