package runners

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/gavel/utils"
)

// FileSelection picks test names or files by exact value, glob or regexp. An
// item matches when any entry matches; the zero value is no filter.
type FileSelection struct {
	Exact []string `json:"exact,omitempty"`
	Glob  []string `json:"glob,omitempty"`
	Regex []string `json:"regex,omitempty"`
}

func (s FileSelection) Empty() bool {
	return len(s.Exact) == 0 && len(s.Glob) == 0 && len(s.Regex) == 0
}

// Merge returns the de-duplicated union of both selections.
func (s FileSelection) Merge(other FileSelection) FileSelection {
	union := func(a, b []string) []string {
		var out []string
		for _, v := range append(slices.Clone(a), b...) {
			if !slices.Contains(out, v) {
				out = append(out, v)
			}
		}
		return out
	}
	return FileSelection{Exact: union(s.Exact, other.Exact), Glob: union(s.Glob, other.Glob), Regex: union(s.Regex, other.Regex)}
}

// Pattern renders the selection as one ORed RE2 expression, the form every
// framework's name filter takes: exact values are quoted and anchored, globs
// are translated (* any run, ? one character) and regexps are kept as given.
func (s FileSelection) Pattern() (string, error) {
	var parts []string
	for _, exact := range s.Exact {
		parts = append(parts, "^"+regexp.QuoteMeta(exact)+"$")
	}
	for _, glob := range s.Glob {
		parts = append(parts, "^"+globRegexp(glob)+"$")
	}
	for _, expr := range s.Regex {
		if _, err := regexp.Compile(expr); err != nil {
			return "", fmt.Errorf("invalid test pattern %q: %w", expr, err)
		}
		parts = append(parts, expr)
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	for i, part := range parts {
		parts[i] = "(?:" + part + ")"
	}
	return strings.Join(parts, "|"), nil
}

// Matches reports whether value is selected. Globs use doublestar semantics so
// a path selection like "src/**/*.test.ts" matches the same files Expand does.
func (s FileSelection) Matches(value string) (bool, error) {
	if slices.Contains(s.Exact, value) {
		return true, nil
	}
	for _, glob := range s.Glob {
		ok, err := doublestar.Match(glob, value)
		if err != nil {
			return false, fmt.Errorf("invalid glob %q: %w", glob, err)
		}
		if ok {
			return true, nil
		}
	}
	for _, expr := range s.Regex {
		re, err := regexp.Compile(expr)
		if err != nil {
			return false, fmt.Errorf("invalid pattern %q: %w", expr, err)
		}
		if re.MatchString(value) {
			return true, nil
		}
	}
	return false, nil
}

// Expand resolves a file selection to the concrete files under root, as
// slash-separated root-relative paths. Exact entries must exist; globs and
// regexps are matched against every non-ignored file. Files matched by skip
// are dropped. A selection that resolves to nothing is an error naming it.
func (s FileSelection) Expand(root string, skip FileSelection) ([]string, error) {
	var files []string
	add := func(rel string) error {
		skipped, err := skip.Matches(rel)
		if err != nil || skipped {
			return err
		}
		if !slices.Contains(files, rel) {
			files = append(files, rel)
		}
		return nil
	}
	for _, exact := range s.Exact {
		rel := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(exact, "./")))
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			return nil, fmt.Errorf("test file %s: %w", exact, err)
		}
		if err := add(rel); err != nil {
			return nil, err
		}
	}
	if len(s.Glob) > 0 || len(s.Regex) > 0 {
		pattern := FileSelection{Glob: s.Glob, Regex: s.Regex}
		err := utils.WalkGitIgnoredBounded(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != root && nodeSkipDirs[d.Name()] {
					return fs.SkipDir
				}
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			ok, err := pattern.Matches(rel)
			if err != nil || !ok {
				return err
			}
			return add(rel)
		})
		if err != nil {
			return nil, err
		}
	}
	if len(files) == 0 && !s.Empty() {
		return nil, fmt.Errorf("no test files under %s match %s", root, s)
	}
	slices.Sort(files)
	return files, nil
}

func (s FileSelection) String() string {
	var parts []string
	parts = append(parts, s.Exact...)
	for _, glob := range s.Glob {
		parts = append(parts, "glob:"+glob)
	}
	for _, expr := range s.Regex {
		parts = append(parts, "regex:"+expr)
	}
	return strings.Join(parts, ", ")
}

func globRegexp(glob string) string {
	var out strings.Builder
	for _, r := range glob {
		switch r {
		case '*':
			out.WriteString(".*")
		case '?':
			out.WriteString(".")
		default:
			out.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	return out.String()
}
