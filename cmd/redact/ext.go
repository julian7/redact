package main

import "github.com/urfave/cli/v3"

var extNameFlag = cli.StringFlag{
	Name:     "name",
	Aliases:  []string{"n"},
	Usage:    "extension name",
	Required: true,
}

func (rt *Runtime) extCmd() *cli.Command {
	return &cli.Command{
		Name:  "ext",
		Usage: "Extension handling",
		Description: `Extensions

Redact can use extensions for key management. Extensions are external
executables for secret key storage and Key Exchange management. Each extension
should support the following commands:

redact-ext-NAME put:  store a secret key.
redact-ext-NAME get:  retrieve a secret key.
redact-ext-NAME list: shows information about the secret or Key Exchange collection.
redact-ext-NAME grant: grants access using key=value inputs and a Key Exchange collection on stdin.
redact-ext-NAME revoke: removes access using key=value inputs, without rotation.
redact-ext-NAME update: refreshes the Key Exchange collection from stdin (key generate).

Use top-level grant/revoke for Key Exchange access. "ext update" edits configuration.
Unsupported commands return 1; missing data returns 2; failures return 3.
Diagnostics belong on stderr; successful get stdout contains only exported key data.
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
