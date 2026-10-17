package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/julian7/redact/ext"
	"github.com/julian7/redact/kx"
	"github.com/julian7/redact/logger"
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

	if cmd.String("ext") == "" && !gpgExtensionMode() {
		items = withoutGPGExtension(items)
	}

	extErr := ext.RunAll(items, operation, buf.Bytes())
	// An explicit extension selection must only call that extension.
	if cmd.String("ext") != "" {
		return extErr
	}

	if ext.ExitCode(extErr) >= ext.MissingData {
		return extErr
	}

	if !gpgExtensionMode() {
		logger.Debugf("Refreshing GPG collection in-tree because FF_GIT_EXT is unset")

		if err := rt.refreshGPGKeys(); err != nil {
			return err
		}

		return nil
	}

	return extErr
}

func withoutGPGExtension(items []ext.Ext) []ext.Ext {
	filtered := make([]ext.Ext, 0, len(items))
	for _, item := range items {
		if item.Name() != gpgExtensionName {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func (rt *Runtime) refreshGPGKeys() error {
	updatedKeys, err := kx.UpdateGPGKeysInKX(rt.Repo, func(w io.Writer) error {
		return rt.SaveTo(w)
	})
	if err != nil {
		logger.Warn(`unable to update secret keys; restore original key with "redact unlock", and try again`)

		return fmt.Errorf("updating key exchange secret keys: %w", err)
	}

	if updatedKeys > 0 {
		fmt.Printf(
			"Updated %d key%s. Don't forget to commit and push new encrypted secret keys.\n",
			updatedKeys,
			map[bool]string{false: "s", true: ""}[updatedKeys == 1],
		)
	}

	return nil
}
