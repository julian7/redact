package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/logger"
	"github.com/julian7/redact/sdk"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type cmdFactory func() (*cobra.Command, error)

type Runtime struct {
	*logger.Logger
	*files.SecretKey
	*viper.Viper
	Config                 string
	StrictPermissionChecks bool
}

func (rt *Runtime) AddCmdTo(cmd *cobra.Command, subcmds []cmdFactory) error {
	for _, cmdFunc := range subcmds {
		subcmd, err := cmdFunc()
		if err != nil {
			return err
		}

		subcmd.SilenceErrors = true
		subcmd.SilenceUsage = true

		cmd.AddCommand(subcmd)
	}

	return nil
}

func (rt *Runtime) FullPath() (string, error) {
	argv0 := os.Args[0]
	if argv0[0] == '.' {
		var err error
		argv0, err = filepath.Abs(argv0)

		if err != nil {
			return "", fmt.Errorf("get absolute path of argv0: %w", err)
		}
	}

	return argv0, nil
}

func (rt *Runtime) Init() {
	if rt.Config != "" {
		rt.Viper.SetConfigFile(rt.Config)
	} else {
		rt.Viper.AddConfigPath("$HOME")
		rt.Viper.SetConfigName(configName)
	}

	rt.Viper.AutomaticEnv()
	rt.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := rt.Viper.ReadInConfig(); err == nil {
		rt.Logger.Debugf("Using config file: %s", rt.Viper.ConfigFileUsed())
	}
}

func (rt *Runtime) RegisterFlags(group string, l *pflag.FlagSet) (err error) {
	nameconv := func(in string) string { return in }
	if len(group) > 0 {
		nameconv = func(in string) string { return fmt.Sprintf("%s.%s", group, in) }
	}

	l.VisitAll(func(flag *pflag.Flag) {
		if err != nil {
			return
		}

		err = rt.Viper.BindPFlag(nameconv(flag.Name), flag)
	})

	return err
}

func (rt *Runtime) RetrieveSecretKey(cmd *cobra.Command, args []string) error {
	var err error

	rt.SecretKey, err = files.NewSecretKey(rt.Logger)
	if err != nil {
		return err
	}

	return sdk.RedactRepo(rt.SecretKey, rt.StrictPermissionChecks)
}

func (rt *Runtime) SaveGitSettings() error {
	argv0, err := rt.FullPath()
	if err != nil {
		return err
	}

	err = sdk.SaveGitSettings(argv0, func(attr string) {
		rt.Logger.Debugf("Setting up filter/diff git config of %s to %s", attr, argv0)
	})
	if err != nil {
		return fmt.Errorf("setting git config: %w", err)
	}

	return nil
}

func (rt *Runtime) SetupLogging(cmd *cobra.Command, args []string) {
	logFile := rt.Viper.GetString("logfile")
	if logFile != "" {
		writer, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			rt.Logger.Warnf("cannot open log file: %v", err)
		} else {
			rt.Logger.SetOutput(writer)
		}
	}

	rt.setLogLevel(strings.ToLower(rt.Viper.GetString("verbosity")))
}

func (rt *Runtime) setLogLevel(level string) {
	err := rt.Logger.SetLevelFromString(level)
	if err != nil {
		rt.Logger.Warnf("cannot set log level: %v", err)

		return
	}

	rt.Logger.Debugf("Setting log level to %s", level)
}
