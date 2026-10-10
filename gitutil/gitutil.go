package gitutil

import (
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
	// Common contains the absolute path of the common git dir
	Common string
	// TopLevel contains a full path of the top level directory of the git repo
	Toplevel string
}

func DetectGitRepo() (*GitRepoInfo, error) {
	out, err := exec.Command(
		"git",
		"rev-parse",
		"--path-format=absolute",
		"--show-toplevel",
		"--git-dir",
		"--git-common-dir",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("retrieving git rev-parse output: %w", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return parseRevParse(string(out), pwd)
}

func parseRevParse(out, pwd string) (*GitRepoInfo, error) {
	data := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(data) > 0 && data[0] == "--path-format=absolute" {
		data = data[1:]
	}

	if len(data) != 3 {
		return nil, ErrParsingGitRevParse
	}

	toplevel, gitDir, common := data[0], data[1], data[2]
	if common == "--git-common-dir" {
		common = gitDir
	}

	return &GitRepoInfo{
		Common:   absPath(pwd, common),
		Toplevel: absPath(pwd, toplevel),
	}, nil
}

func absPath(pwd, path string) string {
	path = filepath.FromSlash(path)
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	return filepath.Join(pwd, path)
}
