package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/julian7/redact/ext"
	"github.com/urfave/cli/v3"
)

const (
	grantOperation  = "grant"
	revokeOperation = "revoke"
)

func descrAppendix(msg string) string {
	return fmt.Sprintf(`%s

Use --ext to select one extension, including an unconfigured extension.
Provide recipient inputs with --option key=value; invocation settings override
configured defaults. Repository changes require a commit and push.`,
		msg)
}

func (rt *Runtime) grantCmd() *cli.Command {
	return &cli.Command{
		Name:        grantOperation,
		Usage:       "Grant access",
		Description: descrAppendix("Grant access and distribute the current Key Exchange collection."),
		Flags:       extensionFlags(),
		Before:      rt.LoadSecretKey,
		Action: func(_ context.Context, cmd *cli.Command) error {
			return rt.collectionDo(cmd, grantOperation)
		},
	}
}

func (rt *Runtime) revokeCmd() *cli.Command {
	return &cli.Command{
		Name:        revokeOperation,
		Usage:       "Revoke access",
		Description: descrAppendix("Remove access from the Key Exchange collection."),
		Flags:       extensionFlags(),
		Action: func(_ context.Context, cmd *cli.Command) error {
			if err := rt.SetupRepo(); err != nil {
				return err
			}

			return rt.collectionDo(cmd, revokeOperation)
		},
	}
}

func (rt *Runtime) collectionDo(cmd *cli.Command, operation string) error {
	if cmd.Args().Len() != 0 {
		return fmt.Errorf("%w: provide extension inputs with --option key=value", ErrOptions)
	}

	items, err := rt.selectExtensions(cmd)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("%w: configure an extension or select one with --ext", ErrExtensionNotFound)
	}

	buf := &bytes.Buffer{}
	if operation == grantOperation {
		if err := rt.Export(buf); err != nil {
			return err
		}
	}

	return ext.RunAll(items, operation, buf.Bytes())
}
