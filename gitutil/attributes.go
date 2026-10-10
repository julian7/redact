package gitutil

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// CheckAttrs fills in filter, diff, and merge attributes for file entries
func (e *FileEntries) CheckAttrs() error {
	cmd := exec.Command(
		"git",
		"check-attr",
		"--stdin",
		"filter",
		"diff",
		"merge",
	)

	feeder, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("getting input pipe: %w", err)
	}

	receiver, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("getting output pipe: %w", err)
	}

	errorstream, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("getting error pipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("starting git command: %w", err)
	}

	go e.feedWithFileNames(feeder)

	go e.logErrors(errorstream)

	err = e.readCheckAttrs(receiver)
	if err != nil {
		e.AddError("git command output", err)
	}

	return cmd.Wait()
}

func (e *FileEntries) feedWithFileNames(writer io.WriteCloser) {
	for _, entry := range e.Items {
		_, err := writer.Write([]byte(entry.Name + "\n"))
		if err != nil {
			e.AddError(entry.Name, err)
		}
	}

	writer.Close()
}

func (e FileEntries) readCheckAttrs(reader io.ReadCloser) error {
	var err error

	defer reader.Close()

	idx := make(map[string]*FileEntry)
	for _, entry := range e.Items {
		idx[entry.Name] = entry
	}

	out, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	outbuf := bytes.NewBuffer(out)

	for {
		var line string

		line, err = outbuf.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}

			break
		}

		line = strings.TrimRight(line, "\n")

		name, attr, value, ok := parseCheckAttrLine(line)
		if !ok {
			err = fmt.Errorf(`%w: "%s"`, ErrParsingCheckAttr, line)

			break
		}

		item, ok := idx[name]
		if !ok {
			e.AddError(name, ErrNotFound)

			continue
		}

		switch attr {
		case "filter":
			item.Filter = value
		case "diff":
			item.Diff = value
		case "merge":
			item.Merge = value
		}
	}

	return err
}

func parseCheckAttrLine(line string) (name, attr, value string, ok bool) {
	rest, value, ok := cutLast(line, ": ")
	if !ok {
		return "", "", "", false
	}

	name, attr, ok = cutLast(rest, ": ")

	return name, attr, value, ok
}

func cutLast(s, sep string) (before, after string, found bool) {
	idx := strings.LastIndex(s, sep)
	if idx < 0 {
		return s, "", false
	}

	return s[:idx], s[idx+len(sep):], true
}

func (e *FileEntries) logErrors(input io.ReadCloser) {
	defer input.Close()

	inbuf := bufio.NewReader(input)

	for {
		line, _, err := inbuf.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}

			e.AddError(string(line), err)

			return
		}

		e.AddError(string(line), err)
	}
}
