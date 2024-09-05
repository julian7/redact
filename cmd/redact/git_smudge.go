package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

func (rt *Runtime) gitSmudgeCmd() *cli.Command {
	return &cli.Command{
		Name:      "smudge",
		Usage:     "Decoding file from STDIN, to STDOUT",
		ArgsUsage: " ",
		Before:    rt.LoadSecretKey,
		Action:    rt.gitSmudgeDo,
	}
}

func (rt *Runtime) gitSmudgeDo(_ context.Context, _ *cli.Command) error {
	err := rt.SecretKey.Decode(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
