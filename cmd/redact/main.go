package main

import (
	"os"

	"github.com/julian7/redact/logger"
)

func main() {
	rt := &Runtime{
		Logger: logger.New(),
	}

	if err := rt.app().Run(os.Args); err != nil {
		rt.Logger.Fatal(err.Error())
	}
}

func openFileToRead(filename string) (*os.File, error) {
	if filename == "" || filename == "-" {
		return os.Stdin, nil
	}

	return os.Open(filename)
}
