package gitutil

import (
	"sync"
)

// FileEntries is a list of all files
type FileEntries struct {
	*sync.RWMutex
	Items  []*FileEntry
	Errors []*NamedError
}

// AddError adds an error into FileEntries in a multithread-safe way
func (e *FileEntries) AddError(name string, err error) {
	e.Lock()
	e.Errors = append(e.Errors, NewError(name, err))
	e.Unlock()
}

// AddFile adds a file into FileEntries in a multithread-safe way
func (e *FileEntries) AddFile(entry *FileEntry) {
	e.Lock()
	e.Items = append(e.Items, entry)
	e.Unlock()
}
