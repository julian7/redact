package gitutil

import "testing"

func TestParseCheckAttrLine(t *testing.T) {
	tt := []struct {
		line  string
		name  string
		attr  string
		value string
		ok    bool
	}{
		{line: "a.key: filter: redact", name: "a.key", attr: "filter", value: "redact", ok: true},
		{line: "dir/b.key: merge: unspecified", name: "dir/b.key", attr: "merge", value: "unspecified", ok: true},
		{line: "odd: name.key: diff: redact", name: "odd: name.key", attr: "diff", value: "redact", ok: true},
		{line: "garbage", ok: false},
		{line: "only: one", ok: false},
	}

	for _, tc := range tt {
		t.Run(tc.line, func(t *testing.T) {
			name, attr, value, ok := parseCheckAttrLine(tc.line)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}

			if !ok {
				return
			}

			if name != tc.name || attr != tc.attr || value != tc.value {
				t.Errorf("expected (%q, %q, %q), got (%q, %q, %q)", tc.name, tc.attr, tc.value, name, attr, value)
			}
		})
	}
}
