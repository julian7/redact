package main

import (
	"github.com/spf13/cobra"
)

func (rt *Runtime) accessCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "access",
		Short: "Key Exchange commands",
		Long: `Key Exchange commands

Key exchange allows contributors to have access to the secret key inside
the git repo by storing them in OpenPGP-encrypted format for each individual.

With Key Exchange commands you can give or revoke access to the project
for contributors by their OpenPGP keys.`,
	}

	subcommands := []cmdFactory{
		rt.accessGrantCmd,
		rt.accessListCmd,
	}

	if err := rt.AddCmdTo(cmd, subcommands); err != nil {
		return nil, err
	}

	return cmd, nil
}
