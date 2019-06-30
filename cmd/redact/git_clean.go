package main

import (
	"os"

	"github.com/julian7/redact/encoder"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gitCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Encoding file from STDIN, to STDOUT",
	Run:   gitCleanDo,
}

func init() {
	flags := gitCleanCmd.Flags()
	flags.Uint32P("epoch", "e", 0, "Use specific key epoch (by default it uses the latest key)")
	viper.BindPFlags(flags)
	gitCmd.AddCommand(gitCleanCmd)
}

func gitCleanDo(cmd *cobra.Command, args []string) {
	var keyEpoch uint32
	masterkey, err := basicDo()
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
