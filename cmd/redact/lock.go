package main

import (
	"fmt"

	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Locks repository",
	Run:   lockDo,
}

func init() {
	rootCmd.AddCommand(lockCmd)
}

func lockDo(cmd *cobra.Command, args []string) {
	masterkey, err := sdk.RedactRepo()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	if err := sdk.RemoveGitSettings(); err != nil {
		cmdErrHandler(errors.Wrap(err, "locking repo"))
	}
	if err := masterkey.Remove(masterkey.KeyFile()); err != nil {
		cmdErrHandler(errors.Wrap(err, "locking repo"))
		return
	}
	if err := sdk.TouchUp(masterkey); err != nil {
		cmdErrHandler(err)
	}
	fmt.Println("Repository locked.")
}
