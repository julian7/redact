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
	version    = "SNAPSHOT"
	rootCmd    = &cobra.Command{
		Use:   "redact",
		Short: "encrypts files in a git repository",
		Long: `redact - keep secrets in a git repository

This application uses gitattributes(5) to encrypt and decrypt files behind
the scenes (see filter and diff attributes). This process requires a master
key, what you can generate with "redact init" command. The master key can
hold multiple key versions, supporting key rotation and retrieval of old
secrets.

Master keys can be distributed inside the repository in the key exchange
directory ($GIT_DIR/.redact), encrypted by contributors' OpenPGP keys.
Contributors can unlock the repo by running "redact unlock".

To make files to be managed by adding the file pattern into a .gitattributes
file like this:

	*.secret.txt filter=redact diff=redact

The subsequent "git add" command will encrypt files matching this pattern.`,
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
	rootCmd.Version = version
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
	if err := viper.BindPFlags(flags); err != nil {
		panic(err)
	}
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
		log.Log().Debugf("Using config file: %s", viper.ConfigFileUsed())
	}
}

func cmdErrHandler(err error) {
	log.Log().Fatalf("%v", err)
}
