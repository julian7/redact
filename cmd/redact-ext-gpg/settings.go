package main

import (
	"fmt"
	"strings"
)

type settings map[string]string

func parseSettings(args []string, allowed ...string) (settings, error) {
	allow := make(map[string]struct{}, len(allowed))
	// Private-key input applies only to get, but configured settings accompany
	// every operation. Other operations must not consume that input.
	allow["private-key"] = struct{}{}

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
