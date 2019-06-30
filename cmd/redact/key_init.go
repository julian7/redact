package main

import (
	"github.com/julian7/redact/keys"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generates initial redact key",
	Run:   initDo,
}

var rootInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generates initial redact key (alias of redact key init)",
	Run:   initDo,
}

func init() {
	keyCmd.AddCommand(initCmd)
	rootCmd.AddCommand(rootInitCmd)
}

func initDo(cmd *cobra.Command, args []string) {
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
	if err == nil {
		logrus.Fatalf("repo already has master key: %s", masterkey)
		return
	}
	masterkey.Generate()
	logrus.Infof("new repo key created: %v", masterkey)
	err = masterkey.Save()
	if err != nil {
		logrus.Fatalf("saving master key: %v", err)
	}
}
