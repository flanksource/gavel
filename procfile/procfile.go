// Package procfile parses Procfiles and supervises the processes they declare.
// A Procfile is YAML: each top-level key is a process name whose value is either
// a command string (`web: npm start`) or a mapping with `command` plus optional
// `default`, `autoRestart`, `cpu`, `mem`, `profiles`, `env`, and `maxRestarts`.
// Declaration order is preserved. The supervisor runs each command via clicky's
// SupervisedProcess, captures output to a per-process log, and applies the
// configured restart policy, resource limits, and profile/default selection.
package procfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
	yamlv3 "gopkg.in/yaml.v3"
)

// DefaultFilename is the conventional Procfile name discovered up the tree.
const DefaultFilename = "Procfile"

// Entry is a single process definition from a Procfile.
type Entry struct {
	Name    string
	Command string
	// Default reports whether the process starts as part of the default set
	// (nil means true). default:false makes it start only on explicit request.
	Default *bool
	// AutoRestart overrides the global restart policy for this process.
	AutoRestart verify.RestartPolicy
	// CPU / Mem override the global resource limits for this process.
	CPU float64
	Mem string
	// Profiles gate auto-start: empty runs in every profile; otherwise the
	// process auto-starts only when one of these is the active profile.
	Profiles []string
	// Env is injected into this process on top of the global/.env layers.
	Env map[string]string
	// MaxRestarts overrides the global cap (nil = use global).
	MaxRestarts *int
}

// procEntry is the YAML object form of an entry (the value when it is a mapping
// rather than a bare command string).
type procEntry struct {
	Command     string               `yaml:"command"`
	Default     *bool                `yaml:"default"`
	AutoRestart verify.RestartPolicy `yaml:"autoRestart"`
	CPU         float64              `yaml:"cpu"`
	Mem         string               `yaml:"mem"`
	Profiles    stringOrSlice        `yaml:"profiles"`
	Env         map[string]string    `yaml:"env"`
	MaxRestarts *int                 `yaml:"maxRestarts"`
}

// stringOrSlice decodes a YAML scalar or sequence of strings into a slice, so
// `profiles: dev` and `profiles: [dev, ci]` are both accepted.
type stringOrSlice []string

func (s *stringOrSlice) UnmarshalYAML(node *yamlv3.Node) error {
	var one string
	if err := node.Decode(&one); err == nil {
		*s = []string{one}
		return nil
	}
	var many []string
	if err := node.Decode(&many); err != nil {
		return err
	}
	*s = many
	return nil
}

// Parse reads Procfile entries from r as YAML, preserving declaration order. A
// malformed document, a non-mapping top level, an invalid or duplicate name, an
// object entry missing its command, or an empty document is a loud error —
// Procfiles are small and hand-edited, so silently dropping a process would hide
// a real mistake.
func Parse(r io.Reader) ([]Entry, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read Procfile: %w", err)
	}
	var doc yamlv3.Node
	if err := yamlv3.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid Procfile: %w", err)
	}
	if doc.Kind == 0 || len(doc.Content) == 0 {
		return nil, fmt.Errorf("no process definitions found in Procfile")
	}
	root := doc.Content[0]
	if root.Kind != yamlv3.MappingNode {
		return nil, fmt.Errorf("invalid Procfile: expected a mapping of `name: command` entries")
	}

	var entries []Entry
	seen := map[string]bool{}
	for i := 0; i+1 < len(root.Content); i += 2 {
		name := root.Content[i].Value
		val := root.Content[i+1]
		if name == "" {
			return nil, fmt.Errorf("invalid Procfile: empty process name")
		}
		if !isValidName(name) {
			return nil, fmt.Errorf("invalid Procfile: invalid process name %q (use letters, digits, _ or -)", name)
		}
		if seen[name] {
			return nil, fmt.Errorf("invalid Procfile: duplicate process name %q", name)
		}
		seen[name] = true

		e := Entry{Name: name}
		switch val.Kind {
		case yamlv3.ScalarNode:
			e.Command = strings.TrimSpace(val.Value)
		case yamlv3.MappingNode:
			var obj procEntry
			if err := val.Decode(&obj); err != nil {
				return nil, fmt.Errorf("invalid Procfile: process %q: %w", name, err)
			}
			e.Command = strings.TrimSpace(obj.Command)
			e.Default = obj.Default
			e.AutoRestart = obj.AutoRestart
			e.CPU = obj.CPU
			e.Mem = obj.Mem
			e.Profiles = []string(obj.Profiles)
			e.Env = obj.Env
			e.MaxRestarts = obj.MaxRestarts
		default:
			return nil, fmt.Errorf("invalid Procfile: process %q: expected a command string or a mapping", name)
		}
		if e.Command == "" {
			return nil, fmt.Errorf("invalid Procfile: process %q has no command", name)
		}
		entries = append(entries, e)
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
