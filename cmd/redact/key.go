package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (rt *Runtime) keyCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Key commands",
		Long: `Master Key management

Key commands let you manage master key. The master key consists of multiple
encryption keys, of which the last one is considered to be the active one.
All the other keys are kept for archival purposes.

A simple "redact key" will return the master key location, and its active
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
	masterkey, err := files.NewMasterKey(rt.Logger)
	if err != nil {
		return errors.Wrap(err, "building master key")
	}
	err = masterkey.Load()
	if err != nil {
		return errors.Wrap(err, "loading master key")
	}
	fmt.Printf("repo key: %v\n", masterkey)

	return nil
}
