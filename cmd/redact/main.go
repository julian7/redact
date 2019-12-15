package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	rt := &Runtime{
		Logger: logrus.New(),
		Viper:  viper.New(),
	}

	cobra.OnInitialize(rt.Init)

	cmd, err := rt.rootCmd()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		rt.Logger.Fatalln("Error:", err)
	}
}
