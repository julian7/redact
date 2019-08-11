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
	Long: `Master Key management

Key commands let you manage master key. The master key consists of multiple
encryption keys, of which the last one is considered to be the active one.
All the other keys are kept for archival purposes.

A simple "redact key" will return the master key location, and its active
key's epoch and signature. For more detailed look, please see
"redact key list".`,
	Run: keyDo,
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
