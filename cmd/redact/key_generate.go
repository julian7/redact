package main

import (
	"fmt"

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
	masterkey, err := basicDo()
	if err != nil {
		cmdErrHandler(err)
	}
	err = saveGitSettings()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "setting git config"))
		return
	}
	masterkey.Generate()
	fmt.Printf("New repo key created: %v\n", masterkey)
	err = masterkey.Save()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "saving master key"))
	}
}
