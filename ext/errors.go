package ext

import "errors"

const (
	Success = iota
	NotApplicable
	MissingData
	Failure
)

var (
	ErrExtAlreadyExists = errors.New("extension already exists")
	ErrExtNotFound      = errors.New("extension not found")
	ErrInvalidEnvFile   = errors.New("invalid dotenv syntax")
	ErrReservedName     = errors.New("_order_ is reserved for extension ordering")
	ErrCmdRequiresExt   = errors.New("--cmd requires --ext")
)

// Error describes a standardized extension outcome. Its cause is retained for
// diagnostics; stderr from the extension is also passed directly to the user.
type Error struct {
	Code int
	Err  error
}

func (err *Error) Error() string { return err.Err.Error() }
func (err *Error) Unwrap() error { return err.Err }

// ExitCode returns the highest severity in a batch. Unclassified errors,
// launch failures, signals, and nonstandard process exits are failures.
func ExitCode(err error) int {
	if err == nil {
		return Success
	}

	// Inspect this node; errors.As could hide a more severe sibling.
	if outcome, ok := err.(*Error); ok { //nolint:errorlint
		if outcome.Code >= NotApplicable && outcome.Code <= Failure {
			return outcome.Code
		}

		return Failure
	}

	if batch, ok := err.(interface{ Unwrap() []error }); ok {
		code := Success
		for _, item := range batch.Unwrap() {
			code = max(code, ExitCode(item))
		}

		return code
	}

	if cause := errors.Unwrap(err); cause != nil {
		return ExitCode(cause)
	}

	return Failure
}
