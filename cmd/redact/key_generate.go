package main

import (
	"fmt"

	"github.com/julian7/redact/sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (rt *Runtime) keyGenerateCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen", "g"},
		Short:   "Generates redact key",
		PreRunE: rt.RetrieveMasterKey,
		RunE:    rt.generateDo,
	}

	return cmd, nil
}

func (rt *Runtime) generateDo(cmd *cobra.Command, args []string) error {
	if err := sdk.SaveGitSettings(); err != nil {
		return errors.Wrap(err, "setting git config")
	}

	if err := rt.MasterKey.Generate(); err != nil {
		return errors.Wrap(err, "generating master key")
	}

	fmt.Printf("New repo key created: %v\n", rt.MasterKey)

	if err := rt.MasterKey.Save(); err != nil {
		return errors.Wrap(err, "saving master key")
	}

	updatedKeys, err := sdk.UpdateMasterExchangeKeys(rt.MasterKey)
	if err != nil {
		rt.Logger.Warn(`unable to update master keys; restore original key with "redact unlock", and try again`)
		return errors.Wrap(err, "updating key exchange master keys")
	}

	if updatedKeys > 0 {
		fmt.Printf(
			"Updated %d key%s. Don't forget to commit new encrypted master keys into the repo.\n",
			updatedKeys,
			map[bool]string{false: "s", true: ""}[updatedKeys == 1],
		)
	}

	return nil
}
