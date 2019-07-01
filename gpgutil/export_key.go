package gpgutil

import "os/exec"

// ExportKey exports GPG key in ASCII armor
func ExportKey(keyIDs []string) ([]byte, error) {
	args := []string{
		"--armor",
		"--export",
	}
	args = append(args, keyIDs...)
	return exec.Command("gpg", args...).Output()
}
