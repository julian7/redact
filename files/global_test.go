package files

import "testing"

func checkError(t *testing.T, expected string, receivedError error) bool {
	if receivedError != nil {
		received := receivedError.Error()
		if expected == "" {
			t.Errorf("Unexpected error: %s", received)
			return false
		}
		if received != expected {
			t.Errorf(
				`Returned error doesn't match.\nExpected: "%s"\nReceived: "%s"`,
				expected,
				received,
			)
			return false
		}
	} else if expected != "" {
		t.Errorf("Unexpected success. Expected error: %s", expected)
		return false
	}
	return true
}
