package ageutil

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Reject relative paths so identity discovery cannot depend on repository contents.
func xdgConfigDir(env func(string) string, home string) (string, error) {
	if config := env("XDG_CONFIG_HOME"); config != "" {
		if !filepath.IsAbs(config) {
			return "", fmt.Errorf("XDG_CONFIG_HOME is not an absolute path: %q", config)
		}

		return config, nil
	}

	if home == "" {
		return "", errors.New("neither XDG_CONFIG_HOME nor a home directory is set")
	}

	return filepath.Join(home, ".config"), nil
}
