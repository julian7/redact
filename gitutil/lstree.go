package gitutil

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type TreeEntry struct {
	Access   int
	Type     string
	ObjectID []byte
	Filename string
}

func LsTree(treeish string, paths []string) ([]*TreeEntry, error) {
	args := []string{
		"ls-tree",
		"-z",
		treeish,
	}
	args = append(args, paths...)

	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("listing git files: %w", err)
	}

	reader := bytes.NewBuffer(out)
	entries := []*TreeEntry{}

	for {
		entry, err := reader.ReadString(0)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("reading git command output: %w", err)
		}

		entry = strings.TrimRight(entry, "\000")

		treeEntry, err := parseTreeEntry(entry)
		if err != nil {
			return nil, err
		}

		entries = append(entries, treeEntry)
	}

	return entries, nil
}

func parseTreeEntry(entry string) (*TreeEntry, error) {
	mainParts := strings.SplitN(entry, "\t", 2)
	if len(mainParts) != 2 {
		return nil, fmt.Errorf("invalid line: %q", entry)
	}

	statParts := strings.Split(mainParts[0], " ")
	if len(statParts) != 3 {
		return nil, fmt.Errorf("invalid stat parts of line: %q", entry)
	}

	access, err := strconv.Atoi(statParts[0])
	if err != nil {
		return nil, fmt.Errorf("parsing line %q: %w", entry, err)
	}

	objectID, err := hex.DecodeString(statParts[2])
	if err != nil {
		return nil, fmt.Errorf("parsing objectid from line %q: %w", entry, err)
	}

	return &TreeEntry{
		Access:   access,
		Type:     statParts[1],
		ObjectID: objectID,
		Filename: mainParts[1],
	}, nil
}
