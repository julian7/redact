package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/julian7/redact/files"
	"github.com/julian7/redact/sdk"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type cmdFactory func() (*cobra.Command, error)

type Runtime struct {
	*logrus.Logger
	*files.MasterKey
	*viper.Viper
	Config string
}

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

func (rt *Runtime) AddCmdTo(cmd *cobra.Command, subcmds []cmdFactory) error {
	for _, cmdFunc := range subcmds {
		subcmd, err := cmdFunc()
		if err != nil {
			return err
		}

		cmd.AddCommand(subcmd)
	}

	return nil
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
	l.VisitAll(func(flag *pflag.Flag) {
		if err != nil {
			return
		}

		name := flag.Name
		if len(group) > 0 {
			name = fmt.Sprintf("%s.%s", group, name)
		}

		err = rt.Viper.BindPFlag(name, flag)
	})

	return err
}

func (rt *Runtime) RetrieveMasterKey(cmd *cobra.Command, args []string) error {
	var err error
	rt.MasterKey, err = sdk.RedactRepo(rt.Logger)

	if err != nil {
		return err
	}

	return nil
}

func (rt *Runtime) SetupLogging(cmd *cobra.Command, args []string) {
	rt.setLogLevel(strings.ToLower(rt.Viper.GetString("verbosity")))

	logFile := rt.Viper.GetString("logfile")
	if logFile != "" {
		writer, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			rt.Logger.Warnf("cannot open log file: %v", err)
		} else {
			rt.Logger.SetOutput(writer)
		}
	}
}

func (rt *Runtime) setLogLevel(level string) {
	var logLevel logrus.Level

	switch level {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "fatal":
		logLevel = logrus.FatalLevel
	case "":
		return
	default:
		rt.Logger.Warnf("unknown log level: %s", level)
		return
	}

	rt.Logger.SetLevel(logLevel)
	rt.Logger.Debugf("Setting log level to %s", level)
}
