package main

import (
	"fmt"

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
	fmt.Printf("repo key: %v\n", rt.MasterKey)
	err := files.EachKey(rt.MasterKey.Keys, func(idx uint32, key files.KeyHandler) error {
		fmt.Printf(" - %s\n", key)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
