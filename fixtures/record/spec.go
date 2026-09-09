// Package record captures diagnostic evidence from a fixture run — how the
// terminal actually rendered, what HTTP calls the child made, what SQL it
// issued — as artifacts on disk that CEL expressions can assert on and the
// dashboard can replay. A failing fixture otherwise offers only stdout, stderr
// and an exit code.
//
// The package deliberately imports nothing from the rest of gavel. Package
// fixtures already sits at the bottom of the report → linters → verify →
// todos/types → fixtures dependency chain, so anything a FixtureResult embeds
// must be importable without bending that chain back around.
package record

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/goccy/go-yaml"
)

// Kind names one recorder. It doubles as the YAML key under `record:` and as
// the suffix in an artifact's filename.
type Kind string

const (
	KindANSI    Kind = "ansi"
	KindHTTP    Kind = "http"
	KindSQL     Kind = "sql"
	KindClients Kind = "clients"
)

// Scope decides whether one recorder is shared by every test in a markdown file
// or started fresh per test. File scope is the default because a proxy or a
// pgwire listener per test would multiply listeners by the fixture count; the
// cost is that attributing an entry to a single test becomes a time-slice
// heuristic (fixtures in a file run concurrently and overlap).
type Scope string

const (
	ScopeFile Scope = "file"
	ScopeTest Scope = "test"
)

// HTTPMode selects how much of the child's HTTP traffic is visible.
// connect sees only the CONNECT tunnels for TLS; mitm terminates TLS with a
// generated CA and is opt-in because trusting that CA is per-runtime
// best-effort (see HTTPOptions.RequireEntries).
type HTTPMode string

const (
	HTTPOff     HTTPMode = "off"
	HTTPConnect HTTPMode = "connect"
	HTTPMITM    HTTPMode = "mitm"
)

// SQLMode selects where SQL is observed. proxy sits between a child process and
// postgres; inprocess hooks gavel's own gorm logger and therefore sees nothing
// a child process does.
type SQLMode string

const (
	SQLOff       SQLMode = "off"
	SQLProxy     SQLMode = "proxy"
	SQLInProcess SQLMode = "inprocess"
)

// Spec is the parsed `record:` block. A nil Spec means no recorder starts —
// no goroutine, no listener, no file. That is load-bearing: every fixture in a
// run shares one parallel task group, so an always-on recorder would multiply
// across the whole run.
type Spec struct {
	ANSI    *ANSIOptions   `yaml:"ansi,omitempty" json:"ansi,omitempty"`
	HTTP    *HTTPOptions   `yaml:"http,omitempty" json:"http,omitempty"`
	SQL     *SQLOptions    `yaml:"sql,omitempty" json:"sql,omitempty"`
	Clients *ClientOptions `yaml:"clients,omitempty" json:"clients,omitempty"`
}

// ANSIOptions configures the asciinema cast recorded from the fixture's PTY.
// Enabling it implies `terminal: pty` — there is no ANSI to record from a pipe.
type ANSIOptions struct {
	Width    int           `yaml:"width,omitempty" json:"width,omitempty"`
	Height   int           `yaml:"height,omitempty" json:"height,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	MaxBytes Size          `yaml:"maxBytes,omitempty" json:"max_bytes,omitempty"`
}

// HTTPOptions configures the proxy the child process is pointed at.
type HTTPOptions struct {
	Mode  HTTPMode `yaml:"mode,omitempty" json:"mode,omitempty"`
	Hosts []string `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	// Bodies caps how much of each request/response body is written to the HAR.
	// Zero in connect mode: there are no bodies to capture through a tunnel.
	Bodies Size `yaml:"bodies,omitempty" json:"bodies,omitempty"`
	// Redact names extra headers to blank out on top of the built-in denylist.
	Redact []string `yaml:"redact,omitempty" json:"redact,omitempty"`
	// RequireEntries fails the fixture when fewer than N entries were captured.
	// This exists because the dangerous MITM failure is silent: the child cannot
	// verify the generated CA, its TLS handshake fails, the recorder captures
	// nothing, and the fixture passes with an empty HAR.
	RequireEntries int   `yaml:"requireEntries,omitempty" json:"require_entries,omitempty"`
	Scope          Scope `yaml:"scope,omitempty" json:"scope,omitempty"`
}

// SQLOptions configures SQL capture.
type SQLOptions struct {
	Mode SQLMode `yaml:"mode,omitempty" json:"mode,omitempty"`
	// DSN is the upstream postgres the proxy forwards to. Empty means the
	// recorder reads it from the fixture's own environment.
	DSN string `yaml:"dsn,omitempty" json:"dsn,omitempty"`
	// Params keeps bind parameter values in the artifact. Off by default: bind
	// params carry the row data, which is the most likely place for a secret.
	Params bool  `yaml:"params,omitempty" json:"params,omitempty"`
	Scope  Scope `yaml:"scope,omitempty" json:"scope,omitempty"`
}

// ClientOptions configures the HAR recorded from gavel's own http.Client calls,
// as opposed to a child process's. It is a separate recorder because those
// clients live all over the gavel process rather than inside a fixture's call
// stack.
type ClientOptions struct {
	Bodies Size     `yaml:"bodies,omitempty" json:"bodies,omitempty"`
	Redact []string `yaml:"redact,omitempty" json:"redact,omitempty"`
}

// Enabled reports whether kind will produce an artifact. Nil-safe so callers can
// write `spec.Enabled(record.KindHTTP)` without a preceding nil check.
func (s *Spec) Enabled(kind Kind) bool {
	if s == nil {
		return false
	}
	switch kind {
	case KindANSI:
		return s.ANSI != nil
	case KindHTTP:
		return s.HTTP != nil && s.HTTP.Mode != HTTPOff
	case KindSQL:
		return s.SQL != nil && s.SQL.Mode != SQLOff
	case KindClients:
		return s.Clients != nil
	default:
		return false
	}
}

// Kinds lists the enabled recorders in a stable order.
func (s *Spec) Kinds() []Kind {
	var kinds []Kind
	for _, kind := range []Kind{KindANSI, KindHTTP, KindSQL, KindClients} {
		if s.Enabled(kind) {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

// IsEmpty reports whether the spec would start nothing.
func (s *Spec) IsEmpty() bool { return len(s.Kinds()) == 0 }

// Implemented lists the recorders that actually have an implementation. The
// `record:` surface, its parsers and the artifact store land ahead of the
// recorders themselves, and each recorder registers itself here as it arrives.
// Asking for one that has not is an error rather than a fixture that quietly
// produces no artifact — a silently absent recording is exactly the failure mode
// this package exists to make visible.
var Implemented = map[Kind]bool{}

// Missing lists the enabled recorders that have no implementation yet.
func (s *Spec) Missing() []Kind {
	var missing []Kind
	for _, kind := range s.Kinds() {
		if !Implemented[kind] {
			missing = append(missing, kind)
		}
	}
	return missing
}

// ApplyDefaults fills the modes and scopes the docs and schema promise, so the
// runtime, `fixtures --help` and the JSON schema agree in exactly one place.
func (s *Spec) ApplyDefaults() {
	if s == nil {
		return
	}
	if s.HTTP != nil {
		if s.HTTP.Mode == "" {
			s.HTTP.Mode = HTTPConnect
		}
		if s.HTTP.Scope == "" {
			s.HTTP.Scope = ScopeFile
		}
	}
	if s.SQL != nil {
		if s.SQL.Mode == "" {
			s.SQL.Mode = SQLProxy
		}
		if s.SQL.Scope == "" {
			s.SQL.Scope = ScopeFile
		}
	}
}

// Parse builds a Spec from the shorthand accepted by the `Record` table column
// and the --record flag: a comma- or space-separated list of kinds. An empty
// value declares nothing and returns nil; "none", "off" and "false" declare an
// opt-out and return an empty Spec, which is what lets a single row escape both
// a file-level `record:` and a run-wide --record.
func Parse(value string) (*Spec, error) {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	if len(fields) == 0 {
		return nil, nil
	}

	spec := &Spec{}
	for _, field := range fields {
		switch Kind(field) {
		case KindANSI:
			spec.ANSI = &ANSIOptions{}
		case KindHTTP:
			spec.HTTP = &HTTPOptions{}
		case KindSQL:
			spec.SQL = &SQLOptions{}
		case KindClients:
			spec.Clients = &ClientOptions{}
		default:
			switch field {
			case "none", "off", "false":
				return &Spec{}, nil
			case "all", "true":
				spec.ANSI, spec.HTTP, spec.SQL = &ANSIOptions{}, &HTTPOptions{}, &SQLOptions{}
			default:
				return nil, fmt.Errorf("record: unknown recorder %q (want ansi, http, sql, clients, all or none)", field)
			}
		}
	}
	spec.ApplyDefaults()
	return spec, nil
}

// UnmarshalYAML accepts all three surfaces the docs advertise: a bare kind
// (`record: ansi`), a list of kinds (`record: [ansi, http]`), and the full
// per-recorder mapping. goccy dispatches to this via its BytesUnmarshaler
// interface, handing over the raw YAML of the node.
func (s *Spec) UnmarshalYAML(data []byte) error {
	var scalar string
	if err := yaml.Unmarshal(data, &scalar); err == nil && strings.TrimSpace(scalar) != "" {
		return s.fromShorthand(scalar)
	}

	var list []string
	if err := yaml.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return s.fromShorthand(strings.Join(list, ","))
	}

	// The alias sheds the UnmarshalYAML method, so this decodes the mapping
	// rather than recursing into this function.
	type plain Spec
	var full plain
	if err := yaml.Unmarshal(data, &full); err != nil {
		return fmt.Errorf("record: %w", err)
	}
	*s = Spec(full)
	s.ApplyDefaults()
	return nil
}

func (s *Spec) fromShorthand(value string) error {
	parsed, err := Parse(value)
	if err != nil {
		return err
	}
	if parsed != nil {
		*s = *parsed
	}
	return nil
}

// Size is a byte count that accepts the humanized forms used in YAML — 4MiB,
// 64KiB, 1MB — as well as a plain integer.
type Size int64

// UnmarshalYAML parses either a number or a humanized size string.
func (z *Size) UnmarshalYAML(data []byte) error {
	var number int64
	if err := yaml.Unmarshal(data, &number); err == nil {
		*z = Size(number)
		return nil
	}

	var text string
	if err := yaml.Unmarshal(data, &text); err != nil {
		return fmt.Errorf("size must be a number or a string like 4MiB: %w", err)
	}
	parsed, err := humanize.ParseBytes(strings.TrimSpace(text))
	if err != nil {
		return fmt.Errorf("size %q: %w", text, err)
	}
	*z = Size(parsed)
	return nil
}

// Or returns the size, or fallback when unset.
func (z Size) Or(fallback Size) Size {
	if z <= 0 {
		return fallback
	}
	return z
}
