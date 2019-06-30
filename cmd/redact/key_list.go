package main

import (
	"fmt"

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
	masterkey, err := basicDo()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	fmt.Printf("repo key: %v", masterkey)
	for idx, key := range masterkey.Keys {
		fmt.Printf(" - %d: %s", idx, key)
	}
}
