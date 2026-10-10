package gitutil

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const gitConfigExitNotFound = 5

// GitConfigGet retrieves configuration data. It returns
// ErrConfigKeyNotFound if the key is not set.
func GitConfigGet(key string) (string, error) {
	out, err := exec.Command(
		"git",
		"config",
		"--get",
		key,
	).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", fmt.Errorf("getting config %s: %w", key, ErrConfigKeyNotFound)
		}

		return "", fmt.Errorf("getting config %s: %w", key, err)
	}

	return strings.TrimRight(string(out), "\n"), nil
}

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
