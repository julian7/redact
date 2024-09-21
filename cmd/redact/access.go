package main

import "github.com/urfave/cli/v3"

func (rt *Runtime) accessCmd() *cli.Command {
	return &cli.Command{
		Name:  "access",
		Usage: "Key Exchange commands",
		Commands: []*cli.Command{
			rt.accessGrantCmd(),
			rt.accessListCmd(),
		},
		Description: `Key Exchange commands

Key exchange allows contributors to have access to the secret key inside
the git repo by storing them in OpenPGP-encrypted format for each individual.

With Key Exchange commands you can give or revoke access to the project
for contributors by their OpenPGP keys.`,
	}
}
