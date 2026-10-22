package main

import "github.com/julian7/redact/ext"

const (
	Success       = ext.Success
	NotApplicable = ext.NotApplicable
	MissingData   = ext.MissingData
	Failure       = ext.Failure
)

type Error = ext.Error

func ExitCode(err error) int { return ext.ExitCode(err) }
