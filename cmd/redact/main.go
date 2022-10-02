package main

import (
	"fmt"
	"os"

	"github.com/julian7/redact/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	rt := &Runtime{
		Logger: logger.New(),
		Viper:  viper.New(),
	}

	cobra.OnInitialize(rt.Init)

	cmd, err := rt.rootCmd()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		rt.Logger.Fatal(err.Error())
	}
}

func openFileToRead(filename string) (*os.File, error) {
	if filename == "" || filename == "-" {
		return os.Stdin, nil
	}

	return os.Open(filename)
}
