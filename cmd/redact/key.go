package main

import (
	"github.com/julian7/redact/files"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Key commands",
	Run:   keyDo,
}

func keyDo(cmd *cobra.Command, args []string) {
	masterkey, err := files.NewMasterKey()
	if err != nil {
		logrus.Fatalf("building master key: %v", err)
		return
	}
	err = masterkey.Load()
	if err != nil {
		logrus.Fatalf("loading master key: %v", err)
		return
	}
	logrus.Infof("repo key: %v", masterkey)
}

func init() {
	rootCmd.AddCommand(keyCmd)
}
