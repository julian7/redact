package gitutil

import (
	"fmt"
)

// FileEntry contains a single file entry in a git repository
type FileEntry struct {
	Filter string
	Mode   int64
	Name   string
	Status byte
	SHA1   [20]byte
	Stage  int64
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
