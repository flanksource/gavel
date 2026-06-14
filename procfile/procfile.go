// Package procfile parses Heroku/foreman-style Procfiles and supervises the
// processes they declare. A Procfile is a list of `name: command` lines; the
// supervisor runs each command via clicky.Exec, captures its output to a
// per-process log, and applies a configurable restart policy.
package procfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/utils"
)

// DefaultFilename is the conventional Procfile name discovered up the tree.
const DefaultFilename = "Procfile"

// Entry is a single `name: command` definition from a Procfile.
type Entry struct {
	Name    string `json:"name" yaml:"name"`
	Command string `json:"command" yaml:"command"`
}

// Parse reads Procfile entries from r. Blank lines and `#` comments are
// skipped. Each remaining line must be `name: command`: the name is everything
// before the first colon, the command everything after. A malformed line, an
// invalid or empty name, an empty command, or a duplicate name is a loud error
// — Procfiles are small and hand-edited, so silently dropping a process would
// hide a real mistake.
func Parse(r io.Reader) ([]Entry, error) {
	var entries []Entry
	seen := map[string]int{}
	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		name, command, ok := strings.Cut(raw, ":")
		if !ok {
			return nil, fmt.Errorf("invalid Procfile: line %d: expected \"name: command\", got %q", lineNo, raw)
		}
		name = strings.TrimSpace(name)
		command = strings.TrimSpace(command)
		if name == "" {
			return nil, fmt.Errorf("invalid Procfile: line %d: empty process name", lineNo)
		}
		if !isValidName(name) {
			return nil, fmt.Errorf("invalid Procfile: line %d: invalid process name %q (use letters, digits, _ or -)", lineNo, name)
		}
		if command == "" {
			return nil, fmt.Errorf("invalid Procfile: line %d: process %q has no command", lineNo, name)
		}
		if prev, dup := seen[name]; dup {
			return nil, fmt.Errorf("invalid Procfile: line %d: duplicate process name %q (first defined on line %d)", lineNo, name, prev)
		}
		seen[name] = lineNo
		entries = append(entries, Entry{Name: name, Command: command})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read Procfile: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no process definitions found in Procfile")
	}
	return entries, nil
}

func isValidName(name string) bool {
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

// Load parses the Procfile at path.
func Load(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open Procfile %s: %w", path, err)
	}
	defer f.Close()
	return Parse(f)
}

// Find resolves which Procfile to use. A non-empty override is resolved against
// dir when relative and returned verbatim. Otherwise the nearest "Procfile" at
// or above dir (bounded by the enclosing git root) is returned. Returns "" when
// no Procfile is found.
func Find(dir, override string) string {
	if override != "" {
		if filepath.IsAbs(override) {
			return override
		}
		return filepath.Join(dir, override)
	}
	root := utils.FindNearestProjectRoot(dir, []string{DefaultFilename})
	if root == "" {
		return ""
	}
	return filepath.Join(root, DefaultFilename)
}

// Select returns the subset of entries whose names appear in names, preserving
// Procfile order. An empty names slice returns all entries. An unknown name is
// a loud error so a typo doesn't silently start nothing.
func Select(entries []Entry, names []string) ([]Entry, error) {
	if len(names) == 0 {
		return entries, nil
	}
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[n] = true
	}
	var out []Entry
	for _, e := range entries {
		if want[e.Name] {
			out = append(out, e)
			delete(want, e.Name)
		}
	}
	if len(want) > 0 {
		missing := make([]string, 0, len(want))
		for n := range want {
			missing = append(missing, n)
		}
		return nil, fmt.Errorf("unknown process(es): %s", strings.Join(missing, ", "))
	}
	return out, nil
}
