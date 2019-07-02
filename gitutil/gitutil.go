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

// GitRepoInfo provides the most basic information about a git repository
type GitRepoInfo struct {
	// Common contains internal git dir inside a workspace
	Common string
	// TopLevel contains a full path of the top level directory of the git repo
	Toplevel string
}

// GitDir detects current git repository's top level and common directories
func GitDir(info *GitRepoInfo) error {
	out, err := exec.Command(
		"git",
		"rev-parse",
		"--show-toplevel",
		"--git-common-dir",
	).Output()
	if err != nil {
		return errors.Wrap(err, "retrieving git rev-parse output")
	}
	data := strings.Split(string(out), "\n")
	if len(data) != 3 {
		return errors.New("error parsing git rev-parse")
	}
	info.Toplevel = data[0]
	info.Common = data[1]
	return nil
}
