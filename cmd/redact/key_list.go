package main

import (
	"github.com/julian7/redact/files"
	"github.com/spf13/cobra"
)

func (rt *Runtime) keyListCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Lists redact keys",
		PreRunE: rt.LoadSecretKey,
		RunE:    rt.listDo,
	}

	return cmd, nil
}

func (rt *Runtime) listDo(cmd *cobra.Command, args []string) error {
	rt.Logger.Infof("repo key: %v", rt.SecretKey)

	return files.EachKey(rt.SecretKey.Keys, func(idx uint32, key files.KeyHandler) error {
		rt.Logger.Infof(" - %s", key)

		return nil
	})
}
