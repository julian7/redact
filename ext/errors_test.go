package ext_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/julian7/redact/ext"
)

func TestBatchExitSeverity(t *testing.T) {
	missing := &ext.Error{Code: ext.MissingData, Err: errors.New("selection missing")}

	failed := &ext.Error{Code: ext.Failure, Err: errors.New("destination failed")}
	for _, err := range []error{
		errors.Join(missing, failed),
		fmt.Errorf("publishing: %w", errors.Join(missing, failed)),
		errors.Join(missing, errors.New("legacy destination failed")),
	} {
		if got := ext.ExitCode(err); got != ext.Failure {
			t.Fatalf("batch severity = %d, want 3: %v", got, err)
		}
	}
}
