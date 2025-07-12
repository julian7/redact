//go:generate go run _generate/generate.go -output generated.go
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var (
	prefixes = map[int]string{
		DebugLevel: "DEBUG",
		InfoLevel:  "INFO",
		WarnLevel:  "WARN",
		ErrorLevel: "ERROR",
		FatalLevel: "FATAL",
	}

	NilLogger = &Discard{}
)

// Logger implements log levels on a Base interface
type Logger struct {
	Base
	level int
}

// Base is a minimal requirement for internal logger implementation.
// log.Logger implements that.
type Base interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	SetOutput(w io.Writer)
}

// New returns new logger with default config
func New() *Logger {
	return &Logger{
		Base:  log.New(os.Stderr, "", log.LstdFlags),
		level: InfoLevel,
	}
}

// With returns new Logger with an already existing log.Logger as a base
func With(base Base) *Logger {
	if base == nil {
		base = NilLogger
	}

	return &Logger{
		Base:  base,
		level: InfoLevel,
	}
}

// SetLevel sets log visibility level
func (l *Logger) SetLevel(level int) {
	l.level = level
}

// Level returns visibility level
func (l *Logger) Level() int {
	return l.level
}

// SetLevelFromString sets log visibility level based on
// their names (like debug, info, warn, error, fatal)
func (l *Logger) SetLevelFromString(level string) error {
	var logLevel int

	switch strings.ToLower(level) {
	case "debug":
		logLevel = DebugLevel
	case "info":
		logLevel = InfoLevel
	case "warn":
		logLevel = WarnLevel
	case "error":
		logLevel = ErrorLevel
	case "fatal":
		logLevel = FatalLevel
	case "":
		return nil
	default:
		return fmt.Errorf("unknown log level %q", level)
	}

	l.SetLevel(logLevel)

	return nil
}

func (l *Logger) silenced(level int) bool {
	return level < l.level
}

func (l *Logger) buildPrefix(level int, msg string) string {
	pfx, ok := prefixes[level]
	if !ok {
		pfx = prefixes[WarnLevel]
	}

	return fmt.Sprintf("%s: %v", pfx, msg)
}

// Log implements generic logging on certain level, using
// log.Print or log.Fatal.
func (l *Logger) Log(level int, attrs ...interface{}) {
	if l.silenced(level) {
		return
	}

	msg := make([]interface{}, 0, len(attrs))
	msg = append(msg, l.buildPrefix(level, attrs[0].(string)))
	msg = append(msg, attrs[1:]...)

	if level == FatalLevel {
		l.Base.Fatal(msg...)
	} else {
		l.Print(msg...)
	}
}

// Logf implements generic logging on certain level, using
// log.Printf or log.Fatalf.
func (l *Logger) Logf(level int, attrs ...interface{}) {
	if l.silenced(level) || len(attrs) < 1 {
		return
	}

	message := l.buildPrefix(level, attrs[0].(string))

	if level == FatalLevel {
		l.Base.Fatalf(message, attrs[1:]...)
	} else {
		l.Printf(message, attrs[1:]...)
	}
}
