package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/kx"
	"github.com/urfave/cli/v3"
)

func (rt *Runtime) keySaveCmd() *cli.Command {
	return &cli.Command{
		Name:    "save",
		Aliases: []string{"s"},
		Usage:   "Saves redact key in Key Exchange",
		Before:  rt.LoadSecretKey,
		Action:  rt.keySaveDo,
	}
}

func (rt *Runtime) keySaveDo(ctx context.Context, cmd *cli.Command) error {
	if err := rt.SaveGitSettings(); err != nil {
		return err
	}

	if err := rt.SecretKey.Save(); err != nil {
		return fmt.Errorf("saving secret key: %w", err)
	}

	extConfig, err := ext.Load(rt.Repo)
	if extConfig != nil {
		buf := &bytes.Buffer{}
		rt.Repo.SecretKey.Export(buf)
		if err = extConfig.SaveKey(buf.Bytes()); err != nil {
			return err
		}
	}

	updatedKeys, err := kx.UpdateGPGKeysInKX(rt.Repo, func(w io.Writer) {
		if err := rt.SecretKey.SaveTo(w); err != nil {
			rt.Logger.Warn(err)
		}
	})
	if err != nil {
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
