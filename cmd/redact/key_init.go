package main

import (
	"fmt"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generates initial redact key",
	Run:   initDo,
}

var rootInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generates initial redact key (alias of redact key init)",
	Run:   initDo,
}

func init() {
	keyCmd.AddCommand(initCmd)
	rootCmd.AddCommand(rootInitCmd)
}

func initDo(cmd *cobra.Command, args []string) {
	err := sdk.SaveGitSettings()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "setting git config"))
		return
	}
	masterkey, err := files.NewMasterKey()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	err = masterkey.Load()
	if err == nil {
		cmdErrHandler(errors.Errorf("repo already has master key: %s", masterkey))
		return
	}
	masterkey.Generate()
	fmt.Printf("New repo key created: %v\n", masterkey)
	err = masterkey.Save()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "saving master key"))
	}
}
