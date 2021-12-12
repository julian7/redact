package main

import (
	"os"

	"github.com/spf13/cobra"
)

func (rt *Runtime) gitSmudgeCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "smudge",
		Args:    cobra.NoArgs,
		Short:   "Decoding file from STDIN, to STDOUT",
		PreRunE: rt.LoadSecretKey,
		RunE:    rt.gitSmudgeDo,
	}

	return cmd, nil
}

func (rt *Runtime) gitSmudgeDo(cmd *cobra.Command, args []string) error {
	err := rt.SecretKey.Decode(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
