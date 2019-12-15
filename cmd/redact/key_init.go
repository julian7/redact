package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/sdk"
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
	err := sdk.SaveGitSettings()
	if err != nil {
		return fmt.Errorf("setting git config: %w", err)
	}

	masterkey, err := files.NewMasterKey(rt.Logger)
	if err != nil {
		return err
	}

	if err := masterkey.Load(); err == nil {
		return fmt.Errorf("repo already has master key: %s", masterkey)
	}

	if err := masterkey.Generate(); err != nil {
		return fmt.Errorf("generating master key: %w", err)
	}

	fmt.Printf("New repo key created: %v\n", masterkey)

	if err := masterkey.Save(); err != nil {
		return fmt.Errorf("saving master key: %w", err)
	}

	return nil
}
