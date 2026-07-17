package ui

import (
	"testing"

	"github.com/flanksource/gavel/todos/drivers"
)

// TestValidateModelForBackend pins that the deduplicated family check accepts a
// model only for a backend in the same provider family, matching the behaviour of
// the strings.HasPrefix heuristic it replaced.
func TestValidateModelForBackend(t *testing.T) {
	cases := []struct {
		backend string
		model   string
		ok      bool
	}{
		{"claude-agent", "claude", true},          // bare sentinel
		{"claude-agent", "claude-sonnet-5", true}, // exact id
		{"claude-agent", "opus", true},            // family alias
		{"claude-cli", "claude-haiku-4-5", true},  // cli backend, same family
		{"claude-agent", "gpt-5.6", false},        // wrong family
		{"codex-agent", "codex", true},            // bare sentinel
		{"codex-agent", "gpt-5.6", true},          // openai family
		{"codex-agent", "", true},                 // empty defers to the sentinel
		{"codex-agent", "claude-sonnet-5", false}, // wrong family
	}
	for _, tc := range cases {
		err := validateModelForBackend(tc.backend, tc.model)
		if (err == nil) != tc.ok {
			t.Errorf("validateModelForBackend(%q, %q) = %v, want ok=%v", tc.backend, tc.model, err, tc.ok)
		}
	}
}

// TestDefaultBackendForDriver pins that the default backend for a (mechanism,
// agent) pair still comes out as the catalog's — the cli mechanism resolves to
// the agent-SDK backend.
func TestDefaultBackendForDriver(t *testing.T) {
	cases := []struct {
		driverName string
		agent      string
		want       string
	}{
		{"cmux", "claude", "claude-cmux"},
		{"cli", "claude", "claude-agent"},
		{"cmux", "codex", "codex-cmux"},
		{"cli", "codex", "codex-agent"},
	}
	for _, tc := range cases {
		kind, err := drivers.Parse(tc.driverName)
		if err != nil {
			t.Fatalf("drivers.Parse(%q): %v", tc.driverName, err)
		}
		if got := defaultBackendForDriver(kind, tc.agent); got != tc.want {
			t.Errorf("defaultBackendForDriver(%s, %s) = %q, want %q", tc.driverName, tc.agent, got, tc.want)
		}
	}
}
