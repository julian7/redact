package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Key commands",
	Run:   keyDo,
}

func keyDo(cmd *cobra.Command, args []string) {
	masterkey, err := files.NewMasterKey()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "building master key"))
		return
	}
	err = masterkey.Load()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "loading master key"))
		return
	}
	fmt.Printf("repo key: %v\n", masterkey)
}

func init() {
	rootCmd.AddCommand(keyCmd)
}
