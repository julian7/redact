package gitutil

import (
	"fmt"
	"os/exec"
)

// GitConfig sets configuration data
func GitConfig(key, val string) error {
	err := exec.Command(
		"git",
		"config",
		key,
		val,
	).Run()
	if err != nil {
		return fmt.Errorf("setting config %s: %w", key, err)
	}

	return nil
}
