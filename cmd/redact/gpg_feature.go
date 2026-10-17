package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

const gpgExtensionName = "gpg"

// gpgExtensionMode selects the GPG extension for the transition period.
// Presence of FF_GIT_EXT enables the extension-backed path.
func gpgExtensionMode() bool {
	_, enabled := os.LookupEnv("FF_GIT_EXT")

	return enabled
}

func requireInTreeGPGMode(ctx context.Context, _ *cli.Command) (context.Context, error) {
	if gpgExtensionMode() {
		return ctx, ErrInTreeGPGDisabled
	}

	return ctx, nil
}
