package gitutil

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CheckAttrs fills in filter attributes for file entries
func (e FileEntries) CheckAttrs(l *logrus.Logger) error {
	cmd := exec.Command(
		"git",
		"check-attr",
		"--stdin",
		"filter",
	)
	feeder, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "getting input pipe")
	}
	receiver, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "getting output pipe")
	}
	errorstream, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "getting error pipe")
	}
	err = cmd.Start()
	if err != nil {
		return errors.Wrap(err, "starting git command")
	}
	go e.feedWithFileNames(feeder)
	go logErrors(errorstream, l)
	err = e.readCheckAttrs(receiver, l)
	if err != nil {
		l.Errorf("git command output error: %v", err)
	}
	err = cmd.Wait()
	return err
}

func (e FileEntries) feedWithFileNames(writer io.WriteCloser) {
	for _, entry := range e {
		writer.Write([]byte(entry.Name + "\n")) // nolint:errcheck
	}
	writer.Close()
}

func (e FileEntries) readCheckAttrs(reader io.ReadCloser, l *logrus.Logger) error {
	var err error
	defer reader.Close()

	idx := make(map[string]*FileEntry)
	for _, entry := range e {
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
			err = errors.Wrapf(err, `finding filter entry in line: "%s"`, line)
			break
		}
		item, ok := idx[items[0]]
		if !ok {
			l.Infof("item not found in file list: %s", items[0])
			continue
		}
		item.Filter = items[1]
	}
	return err
}

func logErrors(input io.ReadCloser, l *logrus.Logger) {
	defer input.Close()
	inbuf := bufio.NewReader(input)

	for {
		line, _, err := inbuf.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}
			l.Errorf("error reading line from error: %v", err)
			return
		}
		l.Errorf("git command error: %s", line)
	}
}
