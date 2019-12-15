package main

import (
	"fmt"

	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (rt *Runtime) lockCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Locks repository",
		Long: `Lock repository

This command removes your master key, and the filter configuration. It also
turns secret files into their unencrypted form. The git repo will behave
as like being not redact-aware. Locally modified or staged files can cause
leaking of secrets, and it's recommended to cancel all local modifications
beforehand.`,
		PreRunE: rt.RetrieveMasterKey,
		RunE: rt.lockDo,
	}

	return cmd, nil
}

func (rt *Runtime) lockDo(cmd *cobra.Command, args []string) error {
	if err := sdk.RemoveGitSettings(); err != nil {
		return errors.Wrap(err, "locking repo")
	}
	if err := rt.MasterKey.Remove(rt.MasterKey.KeyFile()); err != nil {
		return errors.Wrap(err, "locking repo")
	}
	if err := sdk.TouchUp(rt.MasterKey); err != nil {
		return err
	}
	fmt.Println("Repository locked.")

	return nil
}
