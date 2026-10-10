package gitutil

import (
	"fmt"
	"os/exec"
	"strings"
)

// ListAttributesFiles returns .gitattributes files in the repository (tracked
// or not ignored), with paths relative to the top level directory.
func ListAttributesFiles() ([]string, error) {
	out, err := exec.Command(
		"git",
		"ls-files",
		"-z",
		"--cached",
		"--others",
		"--exclude-standard",
		"--full-name",
		"--",
		":(top,glob)**/.gitattributes",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("listing .gitattributes files: %w", err)
	}

	seen := map[string]bool{}
	result := []string{}

	for name := range strings.SplitSeq(strings.TrimRight(string(out), "\000"), "\000") {
		if name == "" || seen[name] {
			continue
		}

		seen[name] = true

		result = append(result, name)
	}

	return result, nil
}
