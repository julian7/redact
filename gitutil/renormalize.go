package gitutil

import (
	"os/exec"
)

// Cat runs git add --renormalize on the repository
func Renormalize(toplevel string, forceRekey bool) error {
	cmd := exec.Command(
		"git",
		"add",
		"--renormalize",
		toplevel,
	)

	if forceRekey {
		cmd.Env = append(cmd.Env, "REDACT_GIT_CLEAN_EPOCH=0")
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}
