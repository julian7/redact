package main

import "github.com/urfave/cli/v3"

func (rt *Runtime) accessGrantCmd() *cli.Command {
	return &cli.Command{
		Name:  "grant",
		Usage: "Grants access to collaborators",
		Commands: []*cli.Command{
			rt.accessGrantGPGCmd(),
			rt.accessGrantAzureCmd(),
		},
		Description: `Access provision in Key Exchange

Access Grant allows contributors to get access to the secret key by securely
sharing it in public.`,
	}
}
