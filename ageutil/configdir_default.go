//go:build !darwin

package ageutil

import "os"

func userConfigDir() (string, error) {
	return os.UserConfigDir()
}
