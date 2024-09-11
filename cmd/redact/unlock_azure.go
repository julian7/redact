package main

import (
	"context"
	"fmt"

	"github.com/julian7/redact/ext/azure"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) unlockAzureCmd() *cli.Command {
	return &cli.Command{
		Name:      "azure",
		Usage:     "Unlocks repository from Azure Key Vault",
		ArgsUsage: " ",
		Description: `Unlock repository from Azure Key Vault

This command unlocks the repository using a key stored in Azure Key Vault secret.

This command reads Azure config in key exchange directory, and based on
its settings, it reads the appropriate key from Key Vault.`,
		Action: rt.unlockAzureDo,
	}
}

func (rt *Runtime) unlockAzureDo(ctx context.Context, cmd *cli.Command) error {
	var err error

	if err := rt.SetupRepo(); err != nil {
		return fmt.Errorf("building secret key: %w", err)
	}

	err = azure.LoadKey(ctx, rt.Repo)
	if err != nil {
		return err
	}

	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	if err := rt.Repo.ForceReencrypt(false, func(err error) {
		rt.Logger.Warn(err.Error())
	}); err != nil {
		return err
	}

	fmt.Println("Repo is unlocked.")

	return nil
}
