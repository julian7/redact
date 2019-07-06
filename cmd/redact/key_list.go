package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/sdk"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists redact keys",
	Run:   listDo,
}

func init() {
	keyCmd.AddCommand(listCmd)
}

func listDo(cmd *cobra.Command, args []string) {
	masterkey, err := sdk.RedactRepo()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	fmt.Printf("repo key: %v\n", masterkey)
	err = files.EachKey(masterkey.Keys, func(idx uint32, key files.KeyHandler) error {
		fmt.Printf(" - %s\n", key)
		return nil
	})
	if err != nil {
		cmdErrHandler(err)
	}
}
