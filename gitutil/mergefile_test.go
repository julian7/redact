package gitutil_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/julian7/redact/gitutil"
)

func TestMergeFile(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tt := []struct {
		name      string
		ours      string
		theirs    string
		conflicts int
		expected  []string
	}{
		{
			name:     "clean",
			ours:     "A\nb\nc\n",
			theirs:   "a\nb\nC\n",
			expected: []string{"A\nb\nC\n"},
		},
		{
			name:      "conflict",
			ours:      "a\nB1\nc\n",
			theirs:    "a\nB2\nc\n",
			conflicts: 1,
			expected:  []string{"<<<<<<<< f (ours)\nB1\n", "B2\n>>>>>>>> f (theirs)\n"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			base := writeFile(t, dir, "base", "a\nb\nc\n")
			ours := writeFile(t, dir, "ours", tc.ours)
			theirs := writeFile(t, dir, "theirs", tc.theirs)

			conflicts, err := gitutil.MergeFile(ours, base, theirs, [3]string{"f (ours)", "f (base)", "f (theirs)"}, 8)
			if err != nil {
				t.Fatal(err)
			}

			if conflicts != tc.conflicts {
				t.Errorf("expected %d conflicts, got %d", tc.conflicts, conflicts)
			}

			result, err := os.ReadFile(ours)
			if err != nil {
				t.Fatal(err)
			}

			for _, exp := range tc.expected {
				if !strings.Contains(string(result), exp) {
					t.Errorf("expected result to contain %q, got:\n%s", exp, result)
				}
			}
		})
	}
}

func TestMergeFileError(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")

	if _, err := gitutil.MergeFile(missing, missing, missing, [3]string{}, 0); err == nil {
		t.Error("expected error on missing files")
	}
}

func writeFile(t *testing.T, dir, name, contents string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}
