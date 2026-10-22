package main

import (
	"bytes"
	"fmt"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

const (
	extCommandFlag = "cmd"
	extOptionFlag  = "option"
)

func extensionFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name: "ext", Aliases: []string{"extension", "x"},
			Usage: "Select one extension, including an unconfigured extension",
		},
		&cli.StringFlag{
			Name: extCommandFlag, Aliases: []string{"c"},
			Usage: "Executable for the selected extension (requires --ext)",
		},
		&cli.StringMapFlag{
			Name: extOptionFlag, Aliases: []string{"o"},
			Usage: "Extension key=value setting; overrides configured defaults for this invocation",
		},
	}
}

func (rt *Runtime) selectExtensions(cmd *cli.Command) ([]ext.Ext, error) {
	conf, err := ext.Load(rt.Repo)
	if err != nil {
		return nil, fmt.Errorf("loading extension config: %w", err)
	}

	return conf.Select(cmd.String("ext"), cmd.String(extCommandFlag), cmd.StringMap(extOptionFlag))
}

// distributeKey retains writes completed before a stopping extension outcome.
func (rt *Runtime) distributeKey(cmd *cli.Command, operation string) error {
	items, err := rt.selectExtensions(cmd)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := rt.Export(buf); err != nil {
		return err
	}

	return ext.RunAll(items, operation, buf.Bytes())
}
