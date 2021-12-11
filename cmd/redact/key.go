package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/spf13/cobra"
)

func (rt *Runtime) keyCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Key commands",
		Long: `Secret Key management

Key commands let you manage secret key. The secret key consists of multiple
encryption keys, of which the last one is considered to be the active one.
All the other keys are kept for archival purposes.

A simple "redact key" will return the secret key location, and its active
key's epoch and signature. For more detailed look, please see
"redact key list".`,
		RunE: rt.keyDo,
	}
	subcommands := []cmdFactory{
		rt.keyGenerateCmd,
		rt.keyInitCmd,
		rt.keyListCmd,
	}

	if err := rt.AddCmdTo(cmd, subcommands); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) keyDo(cmd *cobra.Command, args []string) error {
	secretkey, err := files.NewSecretKey(rt.Logger)
	if err != nil {
		return fmt.Errorf("building secret key: %w", err)
	}

	if err := secretkey.Load(rt.StrictPermissionChecks); err != nil {
		return fmt.Errorf("loading secret key: %w", err)
	}

	rt.Logger.Infof("repo key: %v", secretkey)

	return nil
}
