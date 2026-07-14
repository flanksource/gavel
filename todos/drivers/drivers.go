// Package drivers selects and constructs the agent driver that executes TODOs.
//
// A driver is the mechanism that drives an AI coding agent: the cmux terminal
// automation, a headless stream-json CLI, the Claude Agent SDK bridge, or the
// direct Anthropic API. It is a user-selectable dimension alongside model and
// effort. This package is the single registry both the CLI and the dashboard
// delegate to, so the selection logic lives in one place instead of duplicated
// `switch mode` blocks.
//
// It lives below todos/ (importing todos, todos/cmux, todos/claude) rather than
// in package todos itself, because the concrete executors already import todos —
// putting the factory here keeps the dependency graph acyclic.
package drivers

import (
	"fmt"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/headless"
	"github.com/flanksource/gavel/todos/types"
)

// Kind identifies an agent driver as "<agent>-<mechanism>". The agent is the
// coding agent (claude or codex); the mechanism is how it is driven.
type Kind string

const (
	// ClaudeCmux drives claude's interactive TUI inside a cmux surface.
	ClaudeCmux Kind = "claude-cmux"
	// ClaudeHeadless drives claude via `claude -p --output-format stream-json`.
	ClaudeHeadless Kind = "claude-headless"
	// ClaudeSDK drives claude via the @anthropic-ai/claude-agent-sdk bridge.
	ClaudeSDK Kind = "claude-sdk"
	// ClaudeAPI drives claude via the direct Anthropic API with a local tool loop.
	ClaudeAPI Kind = "claude-api"
	// CodexCmux drives codex's interactive TUI inside a cmux surface.
	CodexCmux Kind = "codex-cmux"
	// CodexHeadless drives codex via the app-server provider.
	CodexHeadless Kind = "codex-headless"
)

// All returns every known driver kind in display order.
func All() []Kind {
	return []Kind{ClaudeCmux, ClaudeHeadless, ClaudeSDK, ClaudeAPI, CodexCmux, CodexHeadless}
}

// Default is the driver used when none is specified.
const Default = ClaudeCmux

// Valid reports whether k is a known driver kind.
func (k Kind) Valid() bool {
	for _, candidate := range All() {
		if k == candidate {
			return true
		}
	}
	return false
}

// Agent returns the coding agent part of the kind ("claude" or "codex").
func (k Kind) Agent() string {
	if i := strings.IndexByte(string(k), '-'); i >= 0 {
		return string(k)[:i]
	}
	return string(k)
}

// Mechanism returns the driving mechanism part ("cmux", "headless", "sdk", "api").
func (k Kind) Mechanism() string {
	if i := strings.IndexByte(string(k), '-'); i >= 0 {
		return string(k)[i+1:]
	}
	return ""
}

// Implemented reports whether New can construct an executor for this kind today.
// Unimplemented kinds are still offered in pickers so they appear as the work
// lands, but selecting one returns a clear error rather than silently falling
// back to another driver.
func (k Kind) Implemented() bool {
	switch k {
	case ClaudeCmux, ClaudeHeadless, CodexHeadless:
		return true
	default:
		return false
	}
}

// Parse validates a driver string (case-insensitive), returning the Kind.
func Parse(s string) (Kind, error) {
	k := Kind(strings.ToLower(strings.TrimSpace(s)))
	if k.Valid() {
		return k, nil
	}
	return "", fmt.Errorf("invalid driver %q (valid: %s)", s, joinKinds(All()))
}

// Config carries the per-run knobs shared by every driver. Each executor uses
// the subset relevant to it (cmux ignores MaxBudgetUsd; the sdk path ignores
// Effort, etc.).
type Config struct {
	WorkDir string
	Model   string
	Backend string
	Effort  string
	// Mode selects the built-in prompt: run (implement) or plan (read-only
	// investigation producing a reviewable plan). Verify never constructs a
	// driver — it routes through the verify engine.
	Mode types.RunMode
	// ExistingPlan is the current content of the todo's recorded plan file
	// (plan mode); empty means no prior plan.
	ExistingPlan string
	// Verifiers gate each run-mode iteration (captain agent.Runner Verify hooks);
	// a failing verdict's feedback drives another attempt in the same session.
	Verifiers     []agent.Verify
	MaxIterations int
	Resume        bool
	SessionID     string
	Timeout       time.Duration
	MaxBudgetUsd  float64
	MaxTurns      int
	Tools         []string
	Dirty         bool
	// ToolModes is the per-tool exposure (tool name → enabled/ask/disabled) and
	// PermissionMode the base permission posture (a clicky ClaudePermissionMode),
	// both honoured by the captain-backed executor (cmux and headless).
	ToolModes      map[string]string
	PermissionMode string
	// PromptOverride, when non-empty, is used verbatim as the agent prompt body
	// instead of the auto-built prompt — the dashboard's editable prompt. The
	// implement/plan scaffolding is still applied per the run mode.
	PromptOverride string
	// Approvals brokers tool permissions to the shared approval registry. Set it
	// only when a resolver (the dashboard) is present; the headless/sdk drivers
	// otherwise block on the first tool needing approval. cmux ignores it (it
	// detects approval prompts on the terminal surface itself).
	Approvals bool
}

// New constructs the executor for a driver kind.
//
// The returned sessionID is the orchestrator session id to seed TODOExecutor
// with: empty for cmux (it mints and manages its own `--session-id`, passed in
// via Config.SessionID, so the orchestrator must not overwrite the todo's prior
// session), and Config.SessionID for the sdk path.
func New(kind Kind, cfg Config) (todos.Executor, string, error) {
	if !kind.Valid() {
		return nil, "", fmt.Errorf("invalid driver %q", kind)
	}
	if cfg.Mode == types.ModeVerify {
		return nil, "", fmt.Errorf("verify runs through the verify engine, not an agent driver")
	}
	model, err := resolveModel(kind, cfg.Model)
	if err != nil {
		return nil, "", err
	}

	backend := cfg.Backend
	switch kind.Mechanism() {
	case "cmux":
		// cmux drives the captain cmux provider through the same captain-backed
		// executor as headless, selected by a cmux backend. It returns "" as the
		// orchestrator session id (it manages its own --session-id via Config.SessionID).
		backend = string(cmuxBackend(kind.Agent()))
	case "headless":
	case "sdk":
		return nil, "", fmt.Errorf("the claude-sdk driver was removed; use claude-headless (the same agent via captain)")
	case "api":
		return nil, "", fmt.Errorf("driver %q is not yet implemented", kind)
	default:
		return nil, "", fmt.Errorf("unhandled driver mechanism %q", kind.Mechanism())
	}
	return headless.NewExecutor(headless.Config{
		WorkDir:        cfg.WorkDir,
		Agent:          kind.Agent(),
		Model:          model,
		Backend:        backend,
		Effort:         cfg.Effort,
		MaxTurns:       cfg.MaxTurns,
		Tools:          cfg.Tools,
		Timeout:        cfg.Timeout,
		Mode:           cfg.Mode,
		ExistingPlan:   cfg.ExistingPlan,
		Verifiers:      cfg.Verifiers,
		MaxIterations:  cfg.MaxIterations,
		PromptOverride: cfg.PromptOverride,
		Approvals:      cfg.Approvals,
		Resume:         cfg.Resume,
		SessionID:      cfg.SessionID,
		ToolModes:      cfg.ToolModes,
		PermissionMode: cfg.PermissionMode,
	}), "", nil
}

// DefaultTools is the standard tool allowlist for the sdk/api drivers.
func DefaultTools() []string {
	return []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"}
}

// cmuxBackend maps a coding agent to its cmux captain backend.
func cmuxBackend(agent string) captainai.Backend {
	if agent == "codex" {
		return captainai.BackendCodexCmux
	}
	return captainai.BackendClaudeCmux
}

// resolveModel reconciles the requested model with the driver's agent. An empty
// codex model is defaulted to "codex" (cmux's ResolveAgent maps "" to claude, so
// codex drivers must carry an explicit codex model). A model whose agent does
// not match the driver's is rejected loudly rather than silently re-agented.
func resolveModel(kind Kind, model string) (string, error) {
	model = strings.TrimSpace(model)
	agent := kind.Agent()
	if model == "" {
		if agent == "codex" {
			return "codex", nil
		}
		return "", nil
	}
	// A driver/executor identity (e.g. "cmux-claude", "headless-codex") is never a
	// model. Guard against it round-tripping through storage into the model field,
	// which would otherwise launch `--model cmux-claude` and fail at the CLI.
	if isExecutorIdentity(model) {
		return "", fmt.Errorf("driver %q: %q is a driver/executor identity, not a model", kind, model)
	}
	got, _ := claude.ResolveAgent(model)
	if got != agent {
		return "", fmt.Errorf("driver %q expects a %s model but %q resolves to %s", kind, agent, model, got)
	}
	return model, nil
}

// isExecutorIdentity reports whether s has the "<mechanism>-<agent>" shape of an
// executor Name() (e.g. "cmux-claude") — the reverse of a driver Kind — so such a
// value can never be mistaken for an LLM model.
func isExecutorIdentity(s string) bool {
	mechanism, agent, ok := strings.Cut(s, "-")
	if !ok {
		return false
	}
	switch mechanism {
	case "cmux", "headless", "sdk", "api":
	default:
		return false
	}
	return agent == "claude" || agent == "codex"
}

func joinKinds(kinds []Kind) string {
	parts := make([]string, len(kinds))
	for i, k := range kinds {
		parts[i] = string(k)
	}
	return strings.Join(parts, ", ")
}
