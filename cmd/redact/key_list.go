package main

import (
	"github.com/julian7/redact/files"
	"github.com/spf13/cobra"
)

func (rt *Runtime) keyListCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Lists redact keys",
		PreRunE: rt.RetrieveMasterKey,
		RunE:    rt.listDo,
	}

	return cmd, nil
}

func (rt *Runtime) listDo(cmd *cobra.Command, args []string) error {
	rt.Logger.Infof("repo key: %v", rt.MasterKey)

	return files.EachKey(rt.MasterKey.Keys, func(idx uint32, key files.KeyHandler) error {
		rt.Logger.Infof(" - %s", key)
		return nil
	})
}
