//go:generate go run _generate/generate.go -output generated.go
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel

	DebugLabel = "DEBUG"
	InfoLabel  = "INFO"
	WarnLabel  = "WARN"
	ErrorLabel = "ERROR"
	FatalLabel = "FATAL"
)

var (
	prefixes = map[int]string{
		DebugLevel: DebugLabel,
		InfoLevel:  InfoLabel,
		WarnLevel:  WarnLabel,
		ErrorLevel: ErrorLabel,
		FatalLevel: FatalLabel,
	}

	standard = log.New(os.Stderr, "", log.LstdFlags)
	level    atomic.Int64
)

func init() {
	level.Store(InfoLevel)
}

// SetOutput changes the destination for all log messages.
func SetOutput(w io.Writer) {
	standard.SetOutput(w)
}

// SetLevel sets the minimum visible log level.
func SetLevel(newLevel int) {
	level.Store(int64(newLevel))
}

// Level returns the minimum visible log level.
func Level() int {
	return int(level.Load())
}

// SetLevelFromString sets the minimum visible log level based on its name.
func SetLevelFromString(name string) error {
	var logLevel int

	switch strings.ToUpper(name) {
	case DebugLabel:
		logLevel = DebugLevel
	case InfoLabel:
		logLevel = InfoLevel
	case WarnLabel:
		logLevel = WarnLevel
	case ErrorLabel:
		logLevel = ErrorLevel
	case FatalLabel:
		logLevel = FatalLevel
	case "":
		return nil
	default:
		return fmt.Errorf("unknown log level %q", name)
	}

	SetLevel(logLevel)

	return nil
}

func silenced(logLevel int) bool {
	return logLevel < Level()
}

func buildPrefix(logLevel int, message string) string {
	prefix, ok := prefixes[logLevel]
	if !ok {
		prefix = prefixes[WarnLevel]
	}

	return fmt.Sprintf("%s: %v", prefix, message)
}

// Log writes a message at the requested level.
func Log(logLevel int, attrs ...any) {
	if silenced(logLevel) {
		return
	}

	message := buildPrefix(logLevel, attrs[0].(string))
	args := append([]any{message}, attrs[1:]...)

	if logLevel == FatalLevel {
		standard.Fatal(args...)
	} else {
		standard.Print(args...)
	}
}

// Logf writes a formatted message at the requested level.
func Logf(logLevel int, attrs ...any) {
	if silenced(logLevel) || len(attrs) < 1 {
		return
	}

	message := buildPrefix(logLevel, attrs[0].(string))
	if logLevel == FatalLevel {
		standard.Fatalf(message, attrs[1:]...)
	} else {
		standard.Printf(message, attrs[1:]...)
	}
}
