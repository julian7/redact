package logger

import "io"

// Discard is a nil logger, implementing Base
type Discard struct{}

// Print is a no-op
func (Discard) Print(...any) {}

// Printf is a no-op
func (Discard) Printf(string, ...any) {}

// Fatal is a no-op
func (Discard) Fatal(...any) {}

// Fatalf is a no-op
func (Discard) Fatalf(string, ...any) {}

// SetOutput is a no-op
func (Discard) SetOutput(io.Writer) {}
