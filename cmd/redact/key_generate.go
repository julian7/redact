package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen", "g"},
	Short:   "Generates redact key",
	Run:     generateDo,
}

func init() {
	keyCmd.AddCommand(generateCmd)
}

func generateDo(cmd *cobra.Command, args []string) {
	err := saveGitSettings()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "setting git config"))
		return
	}
	masterkey, err := files.NewMasterKey()
	if err != nil {
		cmdErrHandler(err)
	}
	err = masterkey.Load()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "loading master key"))
		return
	}
	masterkey.Generate()
	fmt.Printf("New repo key created: %v\n", masterkey)
	err = masterkey.Save()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "saving master key"))
	}
}
