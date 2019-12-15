package main

import (
	"os"

	"github.com/julian7/redact/encoder"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (rt *Runtime) gitCleanCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "clean",
		Args:    cobra.NoArgs,
		Short:   "Encoding file from STDIN, to STDOUT",
		PreRunE: rt.RetrieveMasterKey,
		RunE:    rt.gitCleanDo,
	}

	flags := cmd.Flags()
	flags.Uint32P("epoch", "e", 0, "Use specific key epoch (by default it uses the latest key)")

	if err := rt.RegisterFlags("git.clean", flags); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (rt *Runtime) gitCleanDo(cmd *cobra.Command, args []string) error {
	keyEpoch, err := cast.ToUint32E(viper.Get("git.clean.epoch"))
	if err != nil {
		return err
	}
	if keyEpoch == 0 {
		keyEpoch = rt.MasterKey.LatestKey
	}
	err = rt.MasterKey.Encode(encoder.TypeAES256GCM96, keyEpoch, os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
