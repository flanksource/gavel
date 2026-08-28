package drivers

import (
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
		{"agent", Agent},
		{"api", Api},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("Parse(%q) = %q, %v; want %q, nil", tc.in, got, err, tc.want)
		}
	}
	for _, invalid := range []string{"sdk", "headless", "anthropic", "claude-agent", "codex-cli", "tui"} {
		if _, err := Parse(invalid); err == nil {
			t.Errorf("Parse(%q) should fail for a non-canonical backend", invalid)
		}
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

func TestNewAgentDerivesAgentFromModel(t *testing.T) {
	exec, _, err := New(Agent, todos.AgentRunConfig{
		WorkDir: "/repo",
		Spec:    api.Spec{Model: api.Model{Name: "codex"}},
	})
	if err != nil {
		t.Fatalf("New(agent): %v", err)
	}
	if got := exec.Name(); got != "agent-codex" {
		t.Errorf("agent Name() = %q, want agent-codex", got)
	}
}

func TestNewUsesTheModelBackendAsTheCanonicalDriver(t *testing.T) {
	cases := []struct {
		name  string
		kind  Kind
		model api.Model
		want  string
	}{
		{name: "structured backend", kind: Agent, model: api.Model{Name: "claude", Mode: api.ModeCLI}, want: "cli-claude"},
		{name: "compact prefix", kind: Cmux, model: api.Model{Name: "agent:opus", Mode: api.ModeAPI}, want: "agent-claude"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec, _, err := New(tc.kind, todos.AgentRunConfig{WorkDir: "/repo", Spec: api.Spec{Model: tc.model}})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.kind, err)
			}
			if got := exec.Name(); got != tc.want {
				t.Errorf("Name() = %q, want %q", got, tc.want)
			}
		})
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

func TestNewAPIDriverUsesCaptainRuntime(t *testing.T) {
	if !Api.Implemented() {
		t.Fatal("api must be a first-class implemented backend")
	}
	exec, _, err := New(Api, todos.AgentRunConfig{
		WorkDir: "/repo",
		Spec:    api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAPI}},
	})
	if err != nil {
		t.Fatalf("New(api): %v", err)
	}
	if got := exec.Name(); got != "api-claude" {
		t.Errorf("api Name() = %q, want api-claude", got)
	}
}
