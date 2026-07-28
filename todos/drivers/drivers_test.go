package drivers

import (
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Kind
	}{
		// Canonical mechanisms, case-insensitive with surrounding space.
		{"  CMUX ", Cmux},
		{"cli", Cli},
		{"sdk", Sdk},
		{"api", Api},
		// Legacy mechanism names normalize onto the current enum.
		{"headless", Cli},
		{"agent", Sdk},
		// Legacy composite "<agent>-<mechanism>" keeps only the mechanism part so a
		// TODO persisted with an old driver value still resolves.
		{"claude-cmux", Cmux},
		{"codex-headless", Cli},
		{"claude-sdk", Sdk},
		{"claude-api", Api},
		{"Claude-CMUX", Cmux},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("Parse(%q) = %q, %v; want %q, nil", tc.in, got, err, tc.want)
		}
	}
	if _, err := Parse("claude-tui"); err == nil {
		t.Fatal("Parse(claude-tui) should fail for an unknown mechanism")
	}
	if _, err := Parse("tui"); err == nil {
		t.Fatal("Parse(tui) should fail for an unknown mechanism")
	}
}

func TestNewCmuxDerivesAgentFromModel(t *testing.T) {
	// cmux mints its own --session-id, so the orchestrator session id is empty. An
	// empty model resolves to the claude agent.
	exec, sessionID, err := New(Cmux, todos.AgentRunConfig{WorkDir: "/repo"})
	if err != nil {
		t.Fatalf("New(cmux): %v", err)
	}
	if sessionID != "" {
		t.Errorf("cmux orchestrator sessionID = %q, want empty", sessionID)
	}
	if got := exec.Name(); got != "cmux-claude" {
		t.Errorf("cmux Name() = %q, want cmux-claude", got)
	}

	// The coding agent is derived from the model, not the driver: a codex model
	// selects the codex agent even though the driver is mechanism-only.
	codexExec, _, err := New(Cmux, todos.AgentRunConfig{WorkDir: "/repo", Spec: api.Spec{Model: api.Model{Name: "codex"}}})
	if err != nil {
		t.Fatalf("New(cmux, model=codex): %v", err)
	}
	if got := codexExec.Name(); got != "cmux-codex" {
		t.Errorf("cmux+codex Name() = %q, want cmux-codex", got)
	}
	gptExec, _, err := New(Cmux, todos.AgentRunConfig{WorkDir: "/repo", Spec: api.Spec{Model: api.Model{Name: "gpt-5.5"}}})
	if err != nil {
		t.Fatalf("New(cmux, model=gpt-5.5): %v", err)
	}
	if got := gptExec.Name(); got != "cmux-codex" {
		t.Errorf("cmux+gpt Name() = %q, want cmux-codex", got)
	}
}

func TestNewCliDerivesAgentFromModel(t *testing.T) {
	exec, sessionID, err := New(Cli, todos.AgentRunConfig{WorkDir: "/repo"})
	if err != nil {
		t.Fatalf("New(cli): %v", err)
	}
	if sessionID != "" {
		t.Errorf("cli orchestrator sessionID = %q, want empty", sessionID)
	}
	if got := exec.Name(); got != "cli-claude" {
		t.Errorf("cli Name() = %q, want cli-claude", got)
	}
	codexExec, _, err := New(Cli, todos.AgentRunConfig{WorkDir: "/repo", Spec: api.Spec{Model: api.Model{Name: "codex"}}})
	if err != nil {
		t.Fatalf("New(cli, model=codex): %v", err)
	}
	if got := codexExec.Name(); got != "cli-codex" {
		t.Errorf("cli+codex Name() = %q, want cli-codex", got)
	}
}

func TestNewRejectsVerifyMode(t *testing.T) {
	if _, _, err := New(Cli, todos.AgentRunConfig{WorkDir: "/repo", Mode: types.ModeVerify}); err == nil {
		t.Fatal("New(mode=verify) must error — verify runs through the verify engine")
	}
}

func TestNewRejectsExecutorIdentityAsModel(t *testing.T) {
	// The executor Name() ("cmux-claude") is a driver/executor identity, not a
	// model. If it round-trips through storage into Config.Model it must fail
	// loudly, never launch `--model cmux-claude`.
	for _, model := range []string{"cmux-claude", "headless-claude", "cmux-codex", "headless-codex"} {
		if _, _, err := New(Cmux, todos.AgentRunConfig{WorkDir: "/repo", Spec: api.Spec{Model: api.Model{Name: model}}}); err == nil {
			t.Errorf("New(cmux, model=%q) should reject the executor identity", model)
		}
	}
	// A real hyphenated model must still be accepted.
	if _, _, err := New(Cmux, todos.AgentRunConfig{WorkDir: "/repo", Spec: api.Spec{Model: api.Model{Name: "claude-opus-4-8"}}}); err != nil {
		t.Errorf("New(cmux, model=claude-opus-4-8): unexpected error %v", err)
	}
}

func TestNewUnimplementedDriversFailClearly(t *testing.T) {
	if Sdk.Implemented() {
		t.Error("sdk reported Implemented(), expected not yet")
	}
	if _, _, err := New(Sdk, todos.AgentRunConfig{WorkDir: "/repo"}); err == nil {
		t.Error("New(sdk) should return a not-implemented error")
	} else if !strings.Contains(err.Error(), "cli") {
		t.Errorf("New(sdk) error = %v, want it to name the cli replacement", err)
	}
	if Api.Implemented() {
		t.Error("api reported Implemented(), expected not yet")
	}
	if _, _, err := New(Api, todos.AgentRunConfig{WorkDir: "/repo"}); err == nil {
		t.Error("New(api) should return a not-implemented error")
	}
}
