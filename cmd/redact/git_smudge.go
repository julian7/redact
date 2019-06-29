package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitSmudgeCmd = &cobra.Command{
	Use:   "smudge",
	Short: "Decoding file from STDIN, to STDOUT",
	Run:   gitSmudgeDo,
}

func init() {
	gitCmd.AddCommand(gitSmudgeCmd)
}

func gitSmudgeDo(cmd *cobra.Command, args []string) {
	masterkey, err := basicDo()
	if err != nil {
		logrus.Fatalf("%v", err)
		return
	}
	err = masterkey.Decode(os.Stdin, os.Stdout)
	if err != nil {
		logrus.Fatalf("%v", err)
	}
}
