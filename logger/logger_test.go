package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func captureOutput(t *testing.T) *bytes.Buffer {
	t.Helper()

	output := &bytes.Buffer{}
	SetOutput(output)
	SetLevel(DebugLevel)
	t.Cleanup(func() {
		SetOutput(os.Stderr)
		SetLevel(InfoLevel)
	})

	return output
}

func TestSetLevel(t *testing.T) {
	SetLevel(WarnLevel)

	if got := Level(); got != WarnLevel {
		t.Fatalf("Level() = %d, want %d", got, WarnLevel)
	}

	SetLevel(InfoLevel)
}

func TestSetLevelFromString(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		want    int
		wantErr bool
	}{
		{"debug", "debug", DebugLevel, false},
		{"info", "info", InfoLevel, false},
		{"warn", "warn", WarnLevel, false},
		{"error", "error", ErrorLevel, false},
		{"fatal", "fatal", FatalLevel, false},
		{"unknown", "unknown", InfoLevel, true},
		{"empty", "", InfoLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(InfoLevel)

			err := SetLevelFromString(tt.level)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SetLevelFromString(%q) error = %v, wantErr %v", tt.level, err, tt.wantErr)
			}

			if !tt.wantErr && Level() != tt.want {
				t.Errorf("Level() = %d, want %d", Level(), tt.want)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name    string
		level   int
		message string
	}{
		{"debug", DebugLevel, "DEBUG: onetwo"},
		{"info", InfoLevel, "INFO: onetwo"},
		{"warn", WarnLevel, "WARN: onetwo"},
		{"error", ErrorLevel, "ERROR: onetwo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(t)

			SetLevel(tt.level)

			Log(tt.level, "one", "two")

			if got := output.String(); !strings.Contains(got, tt.message) {
				t.Errorf("log output %q does not contain %q", got, tt.message)
			}
		})
	}
}

func TestLogfLevels(t *testing.T) {
	tests := []struct {
		name    string
		level   int
		message string
	}{
		{"debug", DebugLevel, "DEBUG: one,two"},
		{"info", InfoLevel, "INFO: one,two"},
		{"warn", WarnLevel, "WARN: one,two"},
		{"error", ErrorLevel, "ERROR: one,two"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(t)

			SetLevel(tt.level)

			Logf(tt.level, "one,%s", "two")

			if got := output.String(); !strings.Contains(got, tt.message) {
				t.Errorf("log output %q does not contain %q", got, tt.message)
			}
		})
	}
}

func TestLogfWithoutArguments(t *testing.T) {
	output := captureOutput(t)

	Logf(FatalLevel)

	if got := output.String(); got != "" {
		t.Errorf("Logf(FatalLevel) output = %q, want empty", got)
	}
}
