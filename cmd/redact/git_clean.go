package main

import (
	"os"

	"github.com/julian7/redact/encoder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Encoding file from STDIN, to STDOUT",
	Run:   gitCleanDo,
}

func init() {
	gitCmd.AddCommand(gitCleanCmd)
}

func gitCleanDo(cmd *cobra.Command, args []string) {
	masterkey, err := basicDo()
	if err != nil {
		logrus.Fatalf("%v", err)
		return
	}
	err = masterkey.Encode(encoder.TypeAES256GCM96, masterkey.LatestKey, os.Stdin, os.Stdout)
	if err != nil {
		logrus.Fatalf("%v", err)
	}
}
