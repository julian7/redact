package main

import (
	"fmt"

	"github.com/julian7/redact/sdk"
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
		return fmt.Errorf("setting git config: %w", err)
	}

	if err := rt.MasterKey.Generate(); err != nil {
		return fmt.Errorf("generating master key: %w", err)
	}

	fmt.Printf("New repo key created: %v\n", rt.MasterKey)

	if err := rt.MasterKey.Save(); err != nil {
		return fmt.Errorf("saving master key: %w", err)
	}

	updatedKeys, err := sdk.UpdateMasterExchangeKeys(rt.MasterKey)
	if err != nil {
		rt.Logger.Warn(`unable to update master keys; restore original key with "redact unlock", and try again`)
		return fmt.Errorf("updating key exchange master keys: %w", err)
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
