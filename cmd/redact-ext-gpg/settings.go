package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	// maxPromptLength caps the prompt label, which a repository's extension
	// config supplies, so that a crafted config cannot fill the terminal.
	maxPromptLength = 80
)

type settings map[string]string

func parseSettings(args []string, allowed ...string) (settings, error) {
	allow := make(map[string]struct{}, len(allowed))
	// Private-key input and its prompt label apply only to get, but configured
	// settings accompany every operation. Other operations must not consume them.
	allow["private-key"] = struct{}{}
	allow["prompt"] = struct{}{}

	for _, key := range allowed {
		allow[key] = struct{}{}
	}

	result := make(settings, len(args))
	for _, arg := range args {
		key, value, ok := strings.Cut(arg, "=")
		if !ok || key == "" {
			return nil, failed(fmt.Errorf("invalid setting %q; expected key=value", arg))
		}

		if _, ok := allow[key]; !ok {
			return nil, failed(fmt.Errorf("unknown setting %q", key))
		}

		if prior := result[key]; prior != "" {
			result[key] = prior + "," + value
		} else {
			result[key] = value
		}
	}

	return result, nil
}

func (settings settings) values(key string) []string {
	value := settings[key]
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")

	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}

	return values
}

func (settings settings) identity() string {
	return strings.TrimSpace(settings["gpgkey"])
}

// prompt names the source of the private key read from standard input, so that the
// passphrase request tells which key is being asked for. A repository supplies the
// label through its extension config, where it can name the command that produced
// the key. An empty label keeps the default wording.
//
// The label goes to the terminal as-is, so only printable text is accepted: a
// control character would be acted on by the terminal instead of being read, which
// lets a committed config forge a different prompt or hide what was asked.
func (settings settings) prompt() (string, error) {
	label := strings.TrimSpace(settings["prompt"])
	if label == "" {
		return "", nil
	}

	if utf8.RuneCountInString(label) > maxPromptLength {
		return "", failed(fmt.Errorf("prompt must be at most %d characters", maxPromptLength))
	}

	if strings.IndexFunc(label, func(r rune) bool { return !unicode.IsPrint(r) }) >= 0 {
		return "", failed(errors.New("prompt must be printable text"))
	}

	return label, nil
}
