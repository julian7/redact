package gitutil

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
)

// MergeFile runs a three-way merge of ours, base, and theirs files, writing
// the result into ours. Labels are used in conflict markers, in the order of
// ours, base, theirs. It returns the number of conflicts found.
func MergeFile(ours, base, theirs string, labels [3]string, markerSize int) (int, error) {
	args := []string{"merge-file", "-q"}
	if markerSize > 0 {
		args = append(args, "--marker-size="+strconv.Itoa(markerSize))
	}

	for _, label := range labels {
		args = append(args, "-L", label)
	}

	args = append(args, "--", ours, base, theirs)

	err := exec.Command("git", args...).Run() //nolint:gosec
	if err == nil {
		return 0, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// git merge-file exits with the number of conflicts (capped at
		// 127), or a negative (>= 128 as exit status) value on error.
		if code := exitErr.ExitCode(); code > 0 && code < 128 {
			return code, nil
		}
	}

	return 0, fmt.Errorf("git merge-file: %w", err)
}
