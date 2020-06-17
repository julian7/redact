package gitutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	// LegacyCommon contains --git-dir, which is most likely a good common dir name
	LegacyCommon string
	// TopLevel contains a full path of the top level directory of the git repo
	Toplevel string
}

// GitDir detects current git repository's top level and common directories
func GitDir(info *GitRepoInfo) error {
	out, err := exec.Command(
		"git",
		"rev-parse",
		"--show-toplevel",
		"--git-dir",
		"--git-common-dir",
	).Output()
	if err != nil {
		return fmt.Errorf("retrieving git rev-parse output: %w", err)
	}

	data := strings.Split(string(out), "\n")
	if len(data) != 4 {
		return errors.New("error parsing git rev-parse")
	}

	info.Toplevel = data[0]
	info.LegacyCommon = data[1]
	info.Common = data[2]

	if info.Common == "--git-common-dir" {
		return info.FixCommon()
	}

	return nil
}

// FixCommon fixes common dir setting if legacy git CLI is in use
func (i *GitRepoInfo) FixCommon() error {
	if !filepath.IsAbs(i.LegacyCommon) {
		i.Common = i.LegacyCommon
		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(pwd, i.LegacyCommon)
	if err != nil {
		return fmt.Errorf("cannot find relative path for common dir detection: %w", err)
	}

	i.Common = relPath

	return nil
}
