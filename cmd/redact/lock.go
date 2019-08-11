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
	Long: `Lock repository

This command removes your master key, and the filter configuration. It also
turns secret files into their unencrypted form. The git repo will behave
as like being not redact-aware. Locally modified or staged files can cause
leaking of secrets, and it's recommended to cancel all local modifications
beforehand.`,
	Run: lockDo,
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
