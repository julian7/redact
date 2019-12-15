package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var logLevels = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
	"fatal": logrus.FatalLevel,
}

// Log returns default logger
func Log() *logrus.Logger {
	return logger
}

// SetLogLevel sets log level to predefined values
func SetLogLevel(logLevel string) error {
	if _, ok := logLevels[logLevel]; ok {
		logger.SetLevel(logLevels[logLevel])
	} else {
		return fmt.Errorf("unknown log level: %s", logLevel)
	}

	return nil
}

func init() {
	logger = logrus.New()
}
