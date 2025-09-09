package main

import "github.com/urfave/cli/v3"

func (rt *Runtime) gpgCmd() *cli.Command {
	return &cli.Command{
		Name:    "openpgp",
		Aliases: []string{"gpg"},
		Usage:   "OpenPGP Key Exchange commands",
		Commands: []*cli.Command{
			rt.gpgGrantCmd(),
			rt.gpgListCmd(),
		},
		Description: `OpenPGP Key Exchange commands

Key exchange allows contributors to have access to the secret key inside
the git repo by storing them in OpenPGP-encrypted format for each individual.

With Key Exchange commands you can give or revoke access to the project
for contributors by their OpenPGP keys.`,
	}
}
