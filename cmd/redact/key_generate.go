package main

import (
	"github.com/julian7/redact/keys"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen", "g"},
	Short:   "Generates redact key",
	Run:     generateDo,
}

func init() {
	keyCmd.AddCommand(generateCmd)
}

func generateDo(cmd *cobra.Command, args []string) {
	err := saveGitSettings()
	if err != nil {
		logrus.Fatalf("setting git config: %v", err)
		return
	}
	masterkey, err := keys.NewMasterKey()
	if err != nil {
		logrus.Fatalf("%v", err)
	}
	err = masterkey.Load()
	if err != nil {
		logrus.Fatalf("loading master key: %v", err)
		return
	}
	masterkey.Generate()
	logrus.Infof("new repo key created: %v", masterkey)
	err = masterkey.Save()
	if err != nil {
		logrus.Fatalf("saving master key: %v", err)
	}
}
