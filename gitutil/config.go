package gitutil

import (
	"errors"
	"fmt"
	"os/exec"
)

const gitConfigExitNotFound = 5

// GitConfig sets configuration data
func GitConfig(key, val string) error {
	err := exec.Command(
		"git",
		"config",
		key,
		val,
	).Run()
	if err != nil {
		var exitErr *exec.ExitError
		if key == "--unset" && errors.As(err, &exitErr) && exitErr.ExitCode() == gitConfigExitNotFound {
			return fmt.Errorf("unsetting config %s: %w", val, ErrConfigKeyNotFound)
		}

		return fmt.Errorf("setting config %s: %w", key, err)
	}

	return nil
}
