package main

import (
	"os"

	"github.com/julian7/redact/log"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Log().Errorf("%v", err)
		os.Exit(1)
	}
}
