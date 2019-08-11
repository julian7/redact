package main

import (
	"os"

	"github.com/julian7/redact/sdk"
	"github.com/spf13/cobra"
)

var gitSmudgeCmd = &cobra.Command{
	Use:   "smudge",
	Args:  cobra.NoArgs,
	Short: "Decoding file from STDIN, to STDOUT",
	Run:   gitSmudgeDo,
}

func init() {
	gitCmd.AddCommand(gitSmudgeCmd)
}

func gitSmudgeDo(cmd *cobra.Command, args []string) {
	masterkey, err := sdk.RedactRepo()
	if err != nil {
		cmdErrHandler(err)
		return
	}
	err = masterkey.Decode(os.Stdin, os.Stdout)
	if err != nil {
		cmdErrHandler(err)
	}
}
