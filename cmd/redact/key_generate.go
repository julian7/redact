package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/julian7/redact/sdk"
	"github.com/urfave/cli/v2"
)

func (rt *Runtime) keyGenerateCmd() *cli.Command {
	return &cli.Command{
		Name:      "generate",
		Aliases:   []string{"gen", "g"},
		Usage:     "Generates redact key",
		ArgsUsage: " ",
		Before:    rt.LoadSecretKey,
		Action:    rt.generateDo,
	}
}

func (rt *Runtime) generateDo(ctx *cli.Context) error {
	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	if err := rt.SecretKey.Generate(); err != nil {
		return fmt.Errorf("generating secret key: %w", err)
	}

	rt.Logger.Infof("New repo key created: %v", rt.SecretKey)

	if err := rt.SecretKey.Save(); err != nil {
		return fmt.Errorf("saving secret key: %w", err)
	}

	updatedKeys, err := sdk.UpdateSecretExchangeKeys(rt.Repo, func(w io.Writer) {
		if err := rt.SecretKey.SaveTo(w); err != nil {
			rt.Logger.Warn(err)
		}
	})
	if err != nil {
		var exchangedir *sdk.ExchangeDirError
		if errors.As(err, &exchangedir) {
			return nil
		}

		rt.Logger.Warn(`unable to update secret keys; restore original key with "redact unlock", and try again`)

		return fmt.Errorf("updating key exchange secret keys: %w", err)
	}

	if updatedKeys > 0 {
		fmt.Printf(
			"Updated %d key%s. Don't forget to commit new encrypted secret keys into the repo.\n",
			updatedKeys,
			map[bool]string{false: "s", true: ""}[updatedKeys == 1],
		)
	}

	return nil
}
