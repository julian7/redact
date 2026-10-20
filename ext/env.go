package ext

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"

	"github.com/julian7/redact/files"
)

// privateEnv applies repository-local settings only to this extension process.
func (ext *Ext) privateEnv(cmd *exec.Cmd) error {
	// Extension names must stay within the private configuration directory.
	if ext.name == "" || strings.ContainsAny(ext.name, `/\`) {
		return fmt.Errorf("invalid extension name %q", ext.name)
	}

	if ext.repo.CommonDir() == "" {
		return nil
	}

	filename := filepath.Join(ext.repo.CommonDir(), files.DefaultKeyDir, ext.name+".env")

	fd, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("opening extension environment file %q: %w", filename, err)
	}
	defer fd.Close()

	env, err := godotenv.Parse(fd)
	if err != nil {
		// Parser errors can contain private values from the file.
		return fmt.Errorf("parsing extension environment file %q: %w", filename, ErrInvalidEnvFile)
	}

	cmd.Env = cmd.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	return nil
}
