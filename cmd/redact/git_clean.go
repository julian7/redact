package main

import (
	"os"

	"github.com/julian7/redact/encoder"
	"github.com/julian7/redact/sdk"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gitCleanCmd = &cobra.Command{
	Use:   "clean",
	Args:  cobra.NoArgs,
	Short: "Encoding file from STDIN, to STDOUT",
	Run:   gitCleanDo,
}

func init() {
	flags := gitCleanCmd.Flags()
	flags.Uint32P("epoch", "e", 0, "Use specific key epoch (by default it uses the latest key)")
	gitCmd.AddCommand(gitCleanCmd)
	if err := viper.BindPFlags(flags); err != nil {
		panic(err)
	}
}

func gitCleanDo(cmd *cobra.Command, args []string) {
	var keyEpoch uint32
	masterkey, err := sdk.RedactRepo()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	keyEpoch, err = cast.ToUint32E(viper.Get("epoch"))
	if err != nil {
		cmdErrHandler(err)
		return
	}
	if keyEpoch == 0 {
		keyEpoch = masterkey.LatestKey
	}
	err = masterkey.Encode(encoder.TypeAES256GCM96, keyEpoch, os.Stdin, os.Stdout)
	if err != nil {
		cmdErrHandler(err)
	}
}
