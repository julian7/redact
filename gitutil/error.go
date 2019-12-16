package gitutil

import "fmt"

// NamedError is an error related to a name (like file name)
type NamedError struct {
	Name string
	Orig error
}

// Error describes the NamedError, exposing name and original error.
func (e *NamedError) Error() string {
	return fmt.Sprintf("%v: %s", e.Orig, e.Name)
}

// NewError builds a new NamedError
func NewError(name string, err error) *NamedError {
	return &NamedError{Name: name, Orig: err}
}
