package main

import "os"

const gpgExtensionName = "gpg"

// gpgExtensionMode selects extension-backed GPG collection handling when
// FF_GIT_EXT is present.
func gpgExtensionMode() bool {
	_, enabled := os.LookupEnv("FF_GIT_EXT")

	return enabled
}
