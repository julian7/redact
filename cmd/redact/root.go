package main

import (
	"os"
	"strings"

	"github.com/julian7/redact/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	configName = ".redact"
	rootCmd    = &cobra.Command{
		Use:              "redact",
		Short:            "encrypts files in a git repository",
		PersistentPreRun: setupLogging,
	}
)

func setupLogging(cmd *cobra.Command, args []string) {
	logLevel := strings.ToLower(viper.GetString("verbosity"))
	if logLevel != "" {
		err := log.SetLogLevel(logLevel)
		if err != nil {
			log.Log().Warnf("%v", err)
		} else {
			log.Log().Debugf("Setting log level to %s", logLevel)
		}
	}
	logFile := viper.GetString("logfile")
	if logFile != "" {
		writer, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Log().Warnf("cannot open log file: %v", err)
		} else {
			log.Log().SetOutput(writer)
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	flags := rootCmd.PersistentFlags()
	flags.StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default: ~/"+configName+".yaml)",
	)
	flags.StringP("verbosity", "v", "info", "Verbosity (possible values: debug, info, warn, error, fatal)")
	flags.String("logfile", "", "log file (empty for standard out)")
	viper.BindPFlag("verbosity", flags.Lookup("verbosity"))
	viper.BindPFlag("logfile", flags.Lookup("logfile"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("$HOME")
		viper.SetConfigName(configName)
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err == nil {
		// log.Log().Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}

func cmdErrHandler(err error) {
	log.Log().Fatalf("%v", err)
}
