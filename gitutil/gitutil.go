package gitutil

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// GitDir detects current git repository's git repo directory
func GitDir() (string, error) {
	out, err := exec.Command(
		"git",
		"rev-parse",
		"--git-common-dir",
	).Output()
	if err != nil {
		return "", errors.Wrap(err, "parsing git dir")
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}
