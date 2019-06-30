package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/gitutil"
	"github.com/julian7/redact/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// AttrName defines name used in .gitattribute file's attribute
	// like: `*.key filter=AttrName diff=AttrName`
	AttrName = "redact"
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
	viper.BindPFlag("verbosity", flags.Lookup("verbosity"))
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
		log.Log().Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}

func saveGitSettings() error {
	argv0 := os.Args[0]
	argv0, err := filepath.Abs(argv0)
	if err != nil {
		return errors.Wrap(err, "get absolute path of argv0")
	}
	configItems := map[string]string{
		"filter.%s.clean":  `"%s" git clean`,
		"filter.%s.smudge": `"%s" git smudge`,
		"diff.%s.textconv": `"%s" git diff`,
	}
	for key, val := range configItems {
		if err := gitutil.GitConfig(
			fmt.Sprintf(key, AttrName),
			fmt.Sprintf(val, argv0),
		); err != nil {
			return err
		}
	}
	return nil
}

func basicDo() (*files.MasterKey, error) {
	err := saveGitSettings()
	if err != nil {
		return nil, errors.Wrap(err, "setting git config")
	}
	masterkey, err := files.NewMasterKey()
	if err != nil {
		return nil, errors.Wrap(err, "creating master key object")
	}
	err = masterkey.Load()
	if err != nil {
		return nil, errors.Wrap(err, "loading master key")
	}
	return masterkey, nil
}

func cmdErrHandler(err error) {
	log.Log().Fatalf("%v", err)
}
