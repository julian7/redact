package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists redact keys",
	Run:   listDo,
}

func init() {
	keyCmd.AddCommand(listCmd)
}

func listDo(cmd *cobra.Command, args []string) {
	masterkey, err := basicDo()
	if err != nil {
		logrus.Fatalf("%v", err)
		return
	}
	logrus.Infof("repo key: %v", masterkey)
	for idx, key := range masterkey.Keys {
		logrus.Infof(" - %d: %s", idx, key)
	}
}
