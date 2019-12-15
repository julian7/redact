package gitutil

import (
	"os/exec"

	"github.com/pkg/errors"
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
		return errors.Wrapf(err, "can't set config %s", key)
	}

	return nil
}
