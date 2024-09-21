package main

import "github.com/urfave/cli/v3"

func (rt *Runtime) extCmd() *cli.Command {
	return &cli.Command{
		Name:  "ext",
		Usage: "Extension handling",
		Description: `Extensions

Redact can use extensions for key management. Extensions are external
executables for secret key storage, listing, and retrieval. Each extension
should support the following commands:

azure-ext-NAME put:  store a secret key.
azure-ext-NAME get:  retrieve a secret key.
azure-ext-NAME list: shows information about the secret.
`,
		Commands: commands(
			rt.extAddCmd(),
			rt.extListCmd(),
			rt.extRunCmd(),
			rt.extRemoveCmd(),
			rt.extUpdateCmd(),
		),
	}
}
