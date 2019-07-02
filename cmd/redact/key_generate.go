package main

import (
	"fmt"

	"github.com/julian7/redact/log"
	"github.com/julian7/redact/sdk"
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
	masterkey, err := sdk.RedactRepo()
	if err != nil {
		cmdErrHandler(err)
	}
	err = sdk.SaveGitSettings()
	if err != nil {
		cmdErrHandler(errors.Wrap(err, "setting git config"))
		return
	}
	masterkey.Generate()
	fmt.Printf("New repo key created: %v\n", masterkey)
	if err := masterkey.Save(); err != nil {
		cmdErrHandler(errors.Wrap(err, "saving master key"))
	}
	updatedKeys, err := sdk.UpdateMasterExchangeKeys(masterkey)
	if err != nil {
		log.Log().Warn(`unable to update master keys; restore original key with "redact unlock", and try again`)
		cmdErrHandler(errors.Wrap(err, "updating key exchange master keys"))
	}
	if updatedKeys > 0 {
		fmt.Printf(
			"Updated %d key%s. Don't forget to commit new encrypted master keys into the repo.\n",
			updatedKeys,
			map[bool]string{false: "s", true: ""}[updatedKeys == 1],
		)
	}
}
