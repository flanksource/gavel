package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/flanksource/gavel/fixtures"
)

// Entry is one service of a `kind: services` step. Exactly one resource selects its type.
type Entry struct {
	Name string `json:"name,omitempty"`

	// Deps starts a deps catalog service (postgres, valkey, …) through
	// github.com/flanksource/deps/start.
	Deps       string            `json:"deps,omitempty"`
	Version    string            `json:"version,omitempty"`
	Runtime    string            `json:"runtime,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`

	// Exec runs a supervised `sh -c` command in its own process group.
	Exec  string `json:"exec,omitempty"`
	Cwd   string `json:"cwd,omitempty"`
	Ready *Ready `json:"ready,omitempty"`

	// Git seeds a temp git repository.
	Git    *GitSeed      `json:"git,omitempty"`
	DBTest *DBTestConfig `json:"dbtest,omitempty"`

	Env  map[string]string    `json:"env,omitempty"`
	Wait *fixtures.WaitPolicy `json:"wait,omitempty"`
	// Export maps env names to templates over this entry's outputs; the key
	// "*" exports every key of the map output it references.
	Export map[string]string `json:"export,omitempty"`
}

// Ready is an exec entry's readiness probe: exactly one of HTTP, TCP or Exec.
type Ready struct {
	HTTP string `json:"http,omitempty"`
	// Status is the HTTP status the probe expects, default 200.
	Status int    `json:"status,omitempty"`
	TCP    string `json:"tcp,omitempty"`
	Exec   string `json:"exec,omitempty"`
}

// GitSeed is a git entry: Files are written verbatim (paths and contents are
// not templated, so shell and Go sources keep their `$`), Commit records them
// as one commit.
type GitSeed struct {
	Files  map[string]string `json:"files"`
	Commit bool              `json:"commit,omitempty"`
}

func (e Entry) typeName() string {
	switch {
	case e.Deps != "":
		return "deps"
	case e.Exec != "":
		return "exec"
	case e.DBTest != nil:
		return "dbtest"
	default:
		return "git"
	}
}

// rawEntry is an entry before templating: its name and generic body.
type rawEntry struct {
	name string
	body map[string]any
}

var (
	entryNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	// reservedNames are the step-level output keys that sit next to the
	// per-name aliases.
	reservedNames = map[string]bool{"services": true, "ready": true}
)

// parseConfig decodes `services:` preserving declaration order. Two forms are
// accepted: a mapping (`services: {postgres: {...}, api: {...}}`), whose order
// is read from the JSON token stream rather than a Go map, and a list
// (`services: [{name: postgres, ...}]`) for frontends that cannot preserve
// mapping order when converting YAML to JSON.
func parseConfig(raw json.RawMessage) ([]rawEntry, error) {
	var top struct {
		Services json.RawMessage `json:"services"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&top); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	trimmed := bytes.TrimSpace(top.Services)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, fmt.Errorf("services: at least one entry is required")
	}
	var entries []rawEntry
	var err error
	switch trimmed[0] {
	case '{':
		entries, err = parseOrderedMap(trimmed)
	case '[':
		entries, err = parseList(trimmed)
	default:
		return nil, fmt.Errorf("services must be a mapping or a list, got %s", trimmed)
	}
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("services: at least one entry is required")
	}
	return entries, validateNames(entries)
}

func parseOrderedMap(raw []byte) ([]rawEntry, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	var entries []rawEntry
	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return nil, err
		}
		var body json.RawMessage
		if err := dec.Decode(&body); err != nil {
			return nil, fmt.Errorf("services.%v: %w", token, err)
		}
		entry, err := decodeBody(token.(string), body)
		if err != nil {
			return nil, err
		}
		if _, named := entry.body["name"]; named {
			return nil, fmt.Errorf("services.%s: name is only valid in the list form", entry.name)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func parseList(raw []byte) ([]rawEntry, error) {
	var bodies []json.RawMessage
	if err := json.Unmarshal(raw, &bodies); err != nil {
		return nil, fmt.Errorf("services: %w", err)
	}
	entries := make([]rawEntry, 0, len(bodies))
	for i, body := range bodies {
		entry, err := decodeBody("", body)
		if err != nil {
			return nil, fmt.Errorf("services[%d]: %w", i, err)
		}
		name, ok := entry.body["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("services[%d]: name is required in the list form", i)
		}
		delete(entry.body, "name")
		entry.name = name
		entries = append(entries, entry)
	}
	return entries, nil
}

func decodeBody(name string, raw json.RawMessage) (rawEntry, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var body map[string]any
	if err := dec.Decode(&body); err != nil {
		return rawEntry{}, fmt.Errorf("services.%s: an entry must be a mapping: %w", name, err)
	}
	return rawEntry{name: name, body: body}, nil
}

func validateNames(entries []rawEntry) error {
	seen := map[string]bool{}
	for _, entry := range entries {
		switch {
		case !entryNamePattern.MatchString(entry.name):
			return fmt.Errorf("service name %q must match %s so templates can reference it", entry.name, entryNamePattern)
		case reservedNames[entry.name]:
			return fmt.Errorf("service name %q is reserved", entry.name)
		case seen[entry.name]:
			return fmt.Errorf("service %q is declared twice", entry.name)
		}
		seen[entry.name] = true
	}
	return nil
}

// decodeEntry decodes a rendered body into an Entry and checks the
// type-specific rules.
func decodeEntry(name string, body map[string]any) (Entry, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return Entry{}, err
	}
	var entry Entry
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&entry); err != nil {
		return Entry{}, fmt.Errorf("entry %q: %w", name, err)
	}
	entry.Name = name
	return entry, entry.validate()
}

func (e Entry) validate() error {
	types := 0
	for _, set := range []bool{e.Deps != "", e.Exec != "", e.Git != nil, e.DBTest != nil} {
		if set {
			types++
		}
	}
	if types != 1 {
		return fmt.Errorf("entry %q: exactly one of deps, exec, git or dbtest is required", e.Name)
	}
	if e.Wait != nil && (e.Wait.Until != "" || e.Wait.FailWhen != "") {
		return fmt.Errorf("entry %q: wait.until/failWhen are step-level; an entry's readiness is its ready probe", e.Name)
	}
	switch e.typeName() {
	case "dbtest":
		if e.Ready != nil || e.Wait != nil || e.Cwd != "" || len(e.Env) > 0 || e.Version != "" || e.Runtime != "" || len(e.Parameters) > 0 {
			return fmt.Errorf("entry %q: a dbtest entry takes only dbtest and export settings", e.Name)
		}
		if e.DBTest.Migrations != nil && e.DBTest.Migrations.Dir == "" {
			return fmt.Errorf("entry %q: dbtest migrations.dir is required", e.Name)
		}
	case "exec":
		return e.validateExec()
	case "deps":
		if e.Ready != nil || e.Cwd != "" {
			return fmt.Errorf("entry %q: ready and cwd apply to exec entries; deps services use their catalog health check", e.Name)
		}
		if w := e.Wait; w != nil && (w.Interval != nil || w.Backoff != nil || w.Jitter != "") {
			return fmt.Errorf("entry %q: a deps entry's wait supports only timeout", e.Name)
		}
	default:
		if e.Ready != nil || e.Wait != nil || e.Cwd != "" || len(e.Env) > 0 {
			return fmt.Errorf("entry %q: a git entry takes only files and commit", e.Name)
		}
		if e.Git.Commit && len(e.Git.Files) == 0 {
			return fmt.Errorf("entry %q: git.commit needs at least one file", e.Name)
		}
	}
	return nil
}

func (e Entry) validateExec() error {
	if e.Ready == nil {
		return fmt.Errorf("entry %q: an exec entry needs a ready probe (http, tcp or exec)", e.Name)
	}
	probes := 0
	for _, set := range []bool{e.Ready.HTTP != "", e.Ready.TCP != "", e.Ready.Exec != ""} {
		if set {
			probes++
		}
	}
	if probes != 1 {
		return fmt.Errorf("entry %q: ready needs exactly one of http, tcp or exec", e.Name)
	}
	if e.Ready.Status != 0 && e.Ready.HTTP == "" {
		return fmt.Errorf("entry %q: ready.status applies only to an http probe", e.Name)
	}
	if e.Version != "" || e.Runtime != "" || len(e.Parameters) > 0 {
		return fmt.Errorf("entry %q: version, runtime and parameters apply to deps entries", e.Name)
	}
	if err := readinessPolicy(e.Wait).ValidateTiming(); err != nil {
		return fmt.Errorf("entry %q: %w", e.Name, err)
	}
	return nil
}
