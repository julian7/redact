package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/julian7/redact/ext/azure"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) accessGrantAzureCmd() *cli.Command {
	return &cli.Command{
		Name:      "azure",
		Usage:     "Grants access to collaborators using Azure Key Vault",
		ArgsUsage: "VAULTNAME SECRET",
		Before:    rt.LoadSecretKey,
		Action:    rt.accessGrantAzureDo,
	}
}

func (rt *Runtime) accessGrantAzureDo(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 2 {
		return errors.New("insufficient number of arguments")
	}
	args := cmd.Args()

	vaultname := args.Get(0)
	secretname := args.Get(1)

	if err := azure.SaveAzureRefToKX(rt.Repo, vaultname, secretname); err != nil {
		return fmt.Errorf("saving azure exchange config: %w", err)
	}

	if err := azure.SaveKey(ctx, rt.Repo); err != nil {
		if errors.Is(err, azure.ErrAlreadyWritten) {
			rt.Logger.Infof("secret %s already written to %s vault", secretname, vaultname)
			return nil
		}
		return fmt.Errorf("saving key to Key Vault")
	}
	rt.Logger.Infof("secret %s written to %s vault", secretname, vaultname)

	return nil
}
