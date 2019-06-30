package gitutil

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// FileEntry contains a single file entry in a git repository
type FileEntry struct {
	Status byte
	Mode   int64
	SHA1   [20]byte
	Stage  int64
	Filter string
	Name   string
}

// FileEntries is a list of all files
type FileEntries []*FileEntry

// LsFiles returns files in the repository, possibly filtered by names
func LsFiles(files []string) (FileEntries, error) {
	args := []string{
		"ls-files",
		"--cached",
		"--others",
		"--stage",
		"-t",
		"-z",
		"--exclude-standard",
	}

	if len(files) != 0 {
		args = append(args, "--")
		args = append(args, files...)
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, errors.Wrap(err, "listing git files")
	}
	reader := bytes.NewBuffer(out)
	var allEntries []*FileEntry
	for {
		entry, err := reader.ReadString(0)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrap(err, "reading git command output")
		}
		entry = strings.TrimRight(entry, "\000")
		fileEntry := &FileEntry{Status: entry[0]}
		if fileEntry.Status == StatusOther {
			// unknown files: "? <SPACE> <file name>"
			fileEntry.Name = entry[2:]
		} else {
			// known files: "<metadata <TAB> <file name>"
			contents := strings.Split(entry, "\t")
			if len(contents) != 2 {
				return nil, errors.Errorf("invalid output from git command: %v", entry)
			}
			fileEntry.Name = contents[1]

			// metadata: "<status> <SPACE> <file mode> <SPACE> <sha1> <SPACE> <stage>"
			meta := strings.Split(contents[0], " ")

			mode, err := strconv.ParseInt(meta[1], 8, 64)
			if err != nil {
				return nil, errors.Wrapf(err, `parsing mode for file entry "%s"`, entry)
			}
			fileEntry.Mode = mode

			sha1, err := hex.DecodeString(meta[2])
			if err != nil {
				return nil, errors.Wrapf(err, `parsing SHA1 entry for file entry "%s"`, entry)
			}
			copy(fileEntry.SHA1[:], sha1)

			stage, err := strconv.ParseInt(meta[3], 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err, `parsing stage for file entry "%s"`, entry)
			}
			fileEntry.Stage = stage
		}
		allEntries = append(allEntries, fileEntry)
	}
	_ = out
	return allEntries, nil
}

func (entry FileEntry) String() string {

	return fmt.Sprintf(
		"%c %o (%x) stage %d %s [%s]",
		entry.Status,
		entry.Mode,
		entry.SHA1[0:4],
		entry.Stage,
		entry.Name,
		entry.Filter,
	)
}