package gitutil

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

// CheckAttrs fills in filter attributes for file entries
func (e *FileEntries) CheckAttrs() error {
	cmd := exec.Command(
		"git",
		"check-attr",
		"--stdin",
		"filter",
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

	out, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	outbuf := bytes.NewBuffer(out)

	for {
		var line string

		line, err = outbuf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}

			break
		}

		line = strings.TrimRight(line, "\n")
		items := strings.Split(line, ": filter: ")

		if len(items) != 2 {
			err = fmt.Errorf(`finding filter entry in line: "%s": %w`, line, err)
			break
		}

		item, ok := idx[items[0]]
		if !ok {
			e.AddError(items[0], errors.New("not found"))
		}

		item.Filter = items[1]
	}

	return err
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
