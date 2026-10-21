//go:build darwin

package ageutil

import "os"

// Use XDG paths on macOS too; os.UserConfigDir chooses Application Support.
func userConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return xdgConfigDir(os.Getenv, home)
}
