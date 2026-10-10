package repo

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/go-git/go-billy/v5/util"

	"github.com/julian7/redact/gitutil"
)

var requiredAttrs = []string{"diff", "merge"}

// FixGitAttributes adds missing diff=redact and merge=redact attributes to
// lines with filter=redact in all .gitattributes files of the working tree.
// Lines explicitly setting diff or merge attributes to anything else are
// left alone. The callback is called with each modified file name and the
// number of lines changed.
func (r *Repo) FixGitAttributes(cb func(string, int)) error {
	names, err := gitutil.ListAttributesFiles()
	if err != nil {
		return err
	}

	for _, name := range names {
		if path.Dir(name) == DefaultKeyExchangeDir {
			continue
		}

		changed, err := r.fixGitAttributesFile(name)
		if err != nil {
			return err
		}

		if changed > 0 && cb != nil {
			cb(name, changed)
		}
	}

	return nil
}

func (r *Repo) fixGitAttributesFile(name string) (int, error) {
	st, err := r.Workdir.Stat(name)
	if err != nil {
		return 0, fmt.Errorf("checking %s: %w", name, err)
	}

	if !st.Mode().IsRegular() {
		return 0, nil
	}

	f, err := r.Workdir.Open(name)
	if err != nil {
		return 0, fmt.Errorf("opening %s: %w", name, err)
	}

	data, err := io.ReadAll(f)
	f.Close()

	if err != nil {
		return 0, fmt.Errorf("reading %s: %w", name, err)
	}

	fixed, changed := FixGitAttributesContents(data)
	if changed == 0 {
		return 0, nil
	}

	if err := util.WriteFile(r.Workdir, name, fixed, st.Mode().Perm()); err != nil {
		return 0, fmt.Errorf("writing %s: %w", name, err)
	}

	return changed, nil
}

// FixGitAttributesContents adds missing diff=redact and merge=redact
// attributes to lines with filter=redact. It returns the new contents, and
// the number of lines changed.
func FixGitAttributesContents(data []byte) ([]byte, int) {
	lines := bytes.SplitAfter(data, []byte("\n"))
	changed := 0

	for idx, line := range lines {
		body, eol := splitEOL(string(line))

		missing := missingRequiredAttrs(body)
		if len(missing) == 0 {
			continue
		}

		lines[idx] = []byte(body + " " + strings.Join(missing, " ") + eol)
		changed++
	}

	return bytes.Join(lines, nil), changed
}

func splitEOL(line string) (string, string) {
	trimmed := strings.TrimRight(line, "\r\n")

	return trimmed, line[len(trimmed):]
}

func missingRequiredAttrs(line string) []string {
	attrs := lineAttrs(line)
	if attrs == nil || attrs["filter"] != AttrName {
		return nil
	}

	missing := []string{}

	for _, attr := range requiredAttrs {
		if _, ok := attrs[attr]; !ok {
			missing = append(missing, attr+"="+AttrName)
		}
	}

	return missing
}

func lineAttrs(line string) map[string]string {
	line = strings.TrimSpace(line)
	if line == "" || line[0] == '#' {
		return nil
	}

	var rest string

	if line[0] == '"' {
		end := closingQuote(line)
		if end < 0 {
			return nil
		}

		rest = line[end+1:]
	} else {
		_, rest, _ = strings.Cut(strings.ReplaceAll(line, "\t", " "), " ")
	}

	attrs := map[string]string{}

	for _, field := range strings.Fields(rest) {
		switch {
		case strings.HasPrefix(field, "-"), strings.HasPrefix(field, "!"):
			attrs[field[1:]] = field[:1]
		default:
			key, val, _ := strings.Cut(field, "=")
			attrs[key] = val
		}
	}

	return attrs
}

func closingQuote(line string) int {
	for i := 1; i < len(line); i++ {
		switch line[i] {
		case '\\':
			i++
		case '"':
			return i
		}
	}

	return -1
}
