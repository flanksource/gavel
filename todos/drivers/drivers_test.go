package drivers

import (
	"strings"
	"testing"
)

func TestKindAgentAndMechanism(t *testing.T) {
	cases := []struct {
		kind      Kind
		agent     string
		mechanism string
	}{
		{ClaudeCmux, "claude", "cmux"},
		{ClaudeHeadless, "claude", "headless"},
		{ClaudeSDK, "claude", "sdk"},
		{ClaudeAPI, "claude", "api"},
		{CodexCmux, "codex", "cmux"},
		{CodexHeadless, "codex", "headless"},
	}
	for _, tc := range cases {
		if got := tc.kind.Agent(); got != tc.agent {
			t.Errorf("%s.Agent() = %q, want %q", tc.kind, got, tc.agent)
		}
		if got := tc.kind.Mechanism(); got != tc.mechanism {
			t.Errorf("%s.Mechanism() = %q, want %q", tc.kind, got, tc.mechanism)
		}
	}
}

func TestParse(t *testing.T) {
	if k, err := Parse("  Claude-CMUX "); err != nil || k != ClaudeCmux {
		t.Errorf("Parse normalized input = %q, %v; want claude-cmux, nil", k, err)
	}
	if _, err := Parse("claude-tui"); err == nil {
		t.Fatal("Parse(claude-tui) should fail for an unknown driver")
	}
}

func TestNewCmuxDriversCarryAgentAndNoOrchestratorSession(t *testing.T) {
	// cmux mints its own --session-id, so the orchestrator session id is empty.
	exec, sessionID, err := New(ClaudeCmux, Config{WorkDir: "/repo"})
	if err != nil {
		t.Fatalf("New(claude-cmux): %v", err)
	}
	if sessionID != "" {
		t.Errorf("cmux orchestrator sessionID = %q, want empty", sessionID)
	}
	if got := exec.Name(); got != "cmux-claude" {
		t.Errorf("claude-cmux Name() = %q, want cmux-claude", got)
	}

	// codex-cmux with no explicit model must still resolve to the codex agent
	// (ResolveAgent maps "" to claude, so the driver defaults the model to codex).
	codexExec, _, err := New(CodexCmux, Config{WorkDir: "/repo"})
	if err != nil {
		t.Fatalf("New(codex-cmux): %v", err)
	}
	if got := codexExec.Name(); got != "cmux-codex" {
		t.Errorf("codex-cmux Name() = %q, want cmux-codex", got)
	}
}

func TestNewSDKReturnsConfiguredSessionID(t *testing.T) {
	exec, sessionID, err := New(ClaudeSDK, Config{WorkDir: "/repo", SessionID: "sess-123"})
	if err != nil {
		t.Fatalf("New(claude-sdk): %v", err)
	}
	if sessionID != "sess-123" {
		t.Errorf("sdk orchestrator sessionID = %q, want sess-123", sessionID)
	}
	if got := exec.Name(); got != "claude-code" {
		t.Errorf("claude-sdk Name() = %q, want claude-code", got)
	}
}

func TestNewRejectsModelAgentMismatch(t *testing.T) {
	_, _, err := New(ClaudeCmux, Config{WorkDir: "/repo", Model: "codex"})
	if err == nil {
		t.Fatal("New(claude-cmux, model=codex) should reject the agent mismatch")
	}
	if !strings.Contains(err.Error(), "resolves to codex") {
		t.Errorf("error = %v, want a model/agent mismatch message", err)
	}
}

func TestNewHeadlessDrivers(t *testing.T) {
	exec, sessionID, err := New(ClaudeHeadless, Config{WorkDir: "/repo"})
	if err != nil {
		t.Fatalf("New(claude-headless): %v", err)
	}
	if sessionID != "" {
		t.Errorf("headless orchestrator sessionID = %q, want empty", sessionID)
	}
	if got := exec.Name(); got != "headless-claude" {
		t.Errorf("claude-headless Name() = %q, want headless-claude", got)
	}
	codexExec, _, err := New(CodexHeadless, Config{WorkDir: "/repo"})
	if err != nil {
		t.Fatalf("New(codex-headless): %v", err)
	}
	if got := codexExec.Name(); got != "headless-codex" {
		t.Errorf("codex-headless Name() = %q, want headless-codex", got)
	}
}

func TestNewUnimplementedDriversFailClearly(t *testing.T) {
	if ClaudeAPI.Implemented() {
		t.Error("claude-api reported Implemented(), expected not yet")
	}
	if _, _, err := New(ClaudeAPI, Config{WorkDir: "/repo"}); err == nil {
		t.Error("New(claude-api) should return a not-implemented error")
	}
}
