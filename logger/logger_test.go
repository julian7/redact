//go:generate go run _generate/generate.go -output generated.go
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	got := New()
	if got.Base == nil {
		t.Errorf("internal logger is nil")
	}

	lvl := got.Level()
	if lvl != InfoLevel {
		t.Errorf("default level is %s, not INFO", prefixes[lvl])
	}
}

func TestWith(t *testing.T) {
	stdlogger := log.New(os.Stdout, "", log.LstdFlags)
	tests := []struct {
		name string
		base Base
		want *Logger
	}{
		{"empty logger", nil, &Logger{Base: NilLogger, level: InfoLevel}},
		{"stdlogger", stdlogger, &Logger{Base: stdlogger, level: InfoLevel}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := With(tt.base); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("With() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogger_SetLevel(t *testing.T) {
	type args struct {
		level int
	}

	tests := []struct {
		name   string
		logger *Logger
		args   args
		level  int
	}{
		{"sets level", New(), args{5}, 5},
		{"another level", New(), args{9}, 9},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.logger.SetLevel(tt.args.level)
			got := tt.logger.Level()
			if got != tt.level {
				t.Errorf("Logger.SetLevel() -> %d, wants %d", got, tt.level)
			}
		})
	}
}

func TestLogger_SetLevelFromString(t *testing.T) {
	type args struct {
		level string
	}

	tests := []struct {
		name    string
		logger  *Logger
		args    args
		level   int
		wantErr bool
	}{
		{"debug", New(), args{"debug"}, DebugLevel, false},
		{"info", New(), args{"info"}, InfoLevel, false},
		{"warn", New(), args{"warn"}, WarnLevel, false},
		{"error", New(), args{"error"}, ErrorLevel, false},
		{"fatal", New(), args{"fatal"}, FatalLevel, false},
		{"unknown", New(), args{"unknown"}, 0, true},
		{"empty", New(), args{""}, InfoLevel, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.logger.SetLevelFromString(tt.args.level); (err != nil) != tt.wantErr {
				t.Errorf("Logger.SetLevelFromString() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got := tt.logger.Level()
				if got != tt.level {
					t.Errorf("Logger.SetLevelFromString() -> %d, wants %d", got, tt.level)
				}
			}
		})
	}
}

type logRecorder struct {
	Logs []string
}

func (r *logRecorder) Print(v ...interface{}) {
	r.Logs = append(r.Logs, fmt.Sprint(v...))
}

func (r *logRecorder) Printf(format string, v ...interface{}) {
	r.Logs = append(r.Logs, fmt.Sprintf(format, v...))
}

func (r *logRecorder) Fatal(v ...interface{}) {
	r.Logs = append(r.Logs, fmt.Sprint(v...))
}

func (r *logRecorder) Fatalf(format string, v ...interface{}) {
	r.Logs = append(r.Logs, fmt.Sprintf(format, v...))
}

func (r *logRecorder) SetOutput(io.Writer) {}

func (r *logRecorder) Verify(items []string) error {
	if len(items) != len(r.Logs) {
		return fmt.Errorf("recorded %d logs, wants %d", len(r.Logs), len(items))
	}

	for _, item := range items {
		found := false

		for _, rec := range r.Logs {
			if item == rec {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("message %q not found, got %v", item, r.Logs)
		}
	}

	return nil
}

func TestLogger_Log(t *testing.T) {
	type args struct {
		level int
		attrs []interface{}
	}

	tests := []struct {
		name  string
		args  args
		items []string
		until int
	}{
		{"debug", args{DebugLevel, []interface{}{"one", "two"}}, []string{"DEBUG: onetwo"}, DebugLevel},
		{"info", args{InfoLevel, []interface{}{"one", "two"}}, []string{"INFO: onetwo"}, InfoLevel},
		{"warn", args{WarnLevel, []interface{}{"one", "two"}}, []string{"WARN: onetwo"}, WarnLevel},
		{"error", args{ErrorLevel, []interface{}{"one", "two"}}, []string{"ERROR: onetwo"}, ErrorLevel},
		{"fatal", args{FatalLevel, []interface{}{"one", "two"}}, []string{"FATAL: onetwo"}, FatalLevel},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for level := DebugLevel; level <= FatalLevel; level++ {
				level := level
				t.Run(prefixes[level], func(t *testing.T) {
					recorder := &logRecorder{}
					l := With(recorder)
					l.SetLevel(level)
					l.Log(tt.args.level, tt.args.attrs...)

					if level > tt.until {
						tt.items = []string{}
					}

					if err := recorder.Verify(tt.items); err != nil {
						t.Error(err)
					}
				})
			}
		})
	}
}

func TestLogger_Logf(t *testing.T) {
	type args struct {
		level int
		attrs []interface{}
	}

	tests := []struct {
		name  string
		args  args
		items []string
		until int
	}{
		{"no args", args{FatalLevel, []interface{}{}}, []string{}, FatalLevel},
		{"debug", args{DebugLevel, []interface{}{"one,%s", "two"}}, []string{"DEBUG: one,two"}, DebugLevel},
		{"info", args{InfoLevel, []interface{}{"one,%s", "two"}}, []string{"INFO: one,two"}, InfoLevel},
		{"warn", args{WarnLevel, []interface{}{"one,%s", "two"}}, []string{"WARN: one,two"}, WarnLevel},
		{"error", args{ErrorLevel, []interface{}{"one,%s", "two"}}, []string{"ERROR: one,two"}, ErrorLevel},
		{"fatal", args{FatalLevel, []interface{}{"one,%s", "two"}}, []string{"FATAL: one,two"}, FatalLevel},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for level := DebugLevel; level <= FatalLevel; level++ {
				level := level
				t.Run(prefixes[level], func(t *testing.T) {
					recorder := &logRecorder{}
					l := With(recorder)
					l.SetLevel(level)
					l.Logf(tt.args.level, tt.args.attrs...)

					if level > tt.until {
						tt.items = []string{}
					}

					if err := recorder.Verify(tt.items); err != nil {
						t.Error(err)
					}
				})
			}
		})
	}
}
