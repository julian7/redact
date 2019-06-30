package gitutil

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

const (
	// StatusCached means the file is in the index
	StatusCached = 'H'
	// StatusSkipWorktree represents an entry, which is not stored in Git
	StatusSkipWorktree = 'S'
	// StatusUnmerged means the file is unmerged
	StatusUnmerged = 'M'
	// StatusRemoved represents a file which has been removed
	StatusRemoved = 'R'
	// StatusChanged represents a file which has been changed
	StatusChanged = 'C'
	// StatusKilled represents a file to be killed
	StatusKilled = 'K'
	// StatusOther represents an unknown file, or a file which has an unknown status
	StatusOther = '?'
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
