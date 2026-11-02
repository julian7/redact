package main

import (
	"strings"
	"testing"

	"github.com/julian7/redact/ext"
)

// The label is added to the wording users already read, never substituted for it:
// someone who has typed this passphrase before must not have to work out what the
// changed request is asking for.
func TestPassphrasePrompt(t *testing.T) {
	for _, one := range []struct {
		name  string
		label string
		want  string
	}{
		{name: "no label", label: "", want: defaultPassphrasePrompt},
		{name: "a label", label: "thekey.sh", want: `OpenPGP private-key passphrase for "thekey.sh":`},
	} {
		t.Run(one.name, func(t *testing.T) {
			if got := passphrasePrompt(one.label); got != one.want {
				t.Errorf("passphrasePrompt(%q) = %q, want %q", one.label, got, one.want)
			}
		})
	}
}

// The label comes from a repository's committed extension config and goes to the
// terminal as-is, so it must carry neither what a terminal would act on nor enough
// text to bury the request below it.
func TestPromptSetting(t *testing.T) {
	for _, one := range []struct {
		name     string
		label    string
		want     string
		wantErr  string
		wantCode int
	}{
		{name: "unset", label: "", want: ""},
		{name: "blank", label: "   ", want: ""},
		{name: "a label to be trimmed", label: "  thekey.sh  ", want: "thekey.sh"},
		{
			name:     "a newline that could forge another prompt",
			label:    "key\nEnter admin password:",
			wantErr:  "printable",
			wantCode: ext.Failure,
		},
		{name: "a tab", label: "key\tsource", wantErr: "printable", wantCode: ext.Failure},
		{name: "an escape sequence", label: "key\x1b[2J", wantErr: "printable", wantCode: ext.Failure},
		{
			name:     "a label over the length limit",
			label:    strings.Repeat("k", maxPromptLength+1),
			wantErr:  "at most 80 characters",
			wantCode: ext.Failure,
		},
		{
			name:  "a label at the length limit",
			label: strings.Repeat("k", maxPromptLength),
			want:  strings.Repeat("k", maxPromptLength),
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			got, err := settings{"prompt": one.label}.prompt()
			if err != nil {
				if one.wantErr == "" {
					t.Fatalf("prompt(%q): %v", one.label, err)
				}

				if !strings.Contains(err.Error(), one.wantErr) {
					t.Fatalf("prompt(%q): %q, want an error mentioning %q", one.label, err, one.wantErr)
				}

				if code := ext.ExitCode(err); code != one.wantCode {
					t.Errorf("prompt(%q): exit code %d, want %d", one.label, code, one.wantCode)
				}

				return
			}

			if one.wantErr != "" {
				t.Fatalf("prompt(%q) = %q, want an error mentioning %q", one.label, got, one.wantErr)
			}

			if got != one.want {
				t.Errorf("prompt(%q) = %q, want %q", one.label, got, one.want)
			}
		})
	}
}
