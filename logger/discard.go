package logger

import "io"

// Discard is a nil logger, implementing Base
type Discard struct{}

// Print is a no-op
func (Discard) Print(...interface{}) {}

// Printf is a no-op
func (Discard) Printf(string, ...interface{}) {}

// Fatal is a no-op
func (Discard) Fatal(...interface{}) {}

// Fatalf is a no-op
func (Discard) Fatalf(string, ...interface{}) {}

// SetOutput is a no-op
func (Discard) SetOutput(io.Writer) {}
