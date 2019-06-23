package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Errorf("%v", err)
		os.Exit(1)
	}
}
