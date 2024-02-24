package main

import (
	"os"

	"github.com/urfave/cli/v2"
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

func (rt *Runtime) gitSmudgeDo(_ *cli.Context) error {
	err := rt.SecretKey.Decode(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
