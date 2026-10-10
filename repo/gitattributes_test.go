package repo_test

import (
	"testing"

	"github.com/julian7/redact/repo"
)

func TestFixGitAttributesContents(t *testing.T) {
	tt := []struct {
		name     string
		input    string
		expected string
		changed  int
	}{
		{
			name:     "complete",
			input:    "*.key filter=redact diff=redact merge=redact\n",
			expected: "*.key filter=redact diff=redact merge=redact\n",
		},
		{
			name:     "missing merge",
			input:    "*.key filter=redact diff=redact\n",
			expected: "*.key filter=redact diff=redact merge=redact\n",
			changed:  1,
		},
		{
			name:     "missing both, tabs, no final newline",
			input:    "*.key\tfilter=redact",
			expected: "*.key\tfilter=redact diff=redact merge=redact",
			changed:  1,
		},
		{
			name:     "CRLF",
			input:    "*.key filter=redact diff=redact\r\n*.txt text\r\n",
			expected: "*.key filter=redact diff=redact merge=redact\r\n*.txt text\r\n",
			changed:  1,
		},
		{
			name:     "explicit other merge is kept",
			input:    "*.key filter=redact diff=redact -merge\n*.pem filter=redact merge=binary\n",
			expected: "*.key filter=redact diff=redact -merge\n*.pem filter=redact merge=binary diff=redact\n",
			changed:  1,
		},
		{
			name:     "comments, other filters, and blank lines are left alone",
			input:    "# *.key filter=redact\n\n*.bin filter=lfs diff=lfs merge=lfs\n",
			expected: "# *.key filter=redact\n\n*.bin filter=lfs diff=lfs merge=lfs\n",
		},
		{
			name:     "quoted pattern",
			input:    "\"my secret.key\" filter=redact diff=redact\n",
			expected: "\"my secret.key\" filter=redact diff=redact merge=redact\n",
			changed:  1,
		},
		{
			name:     "macro definition",
			input:    "[attr]secret filter=redact\n*.key secret\n",
			expected: "[attr]secret filter=redact diff=redact merge=redact\n*.key secret\n",
			changed:  1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, changed := repo.FixGitAttributesContents([]byte(tc.input))
			if string(out) != tc.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tc.expected, out)
			}

			if changed != tc.changed {
				t.Errorf("expected %d changed lines, got %d", tc.changed, changed)
			}
		})
	}
}
