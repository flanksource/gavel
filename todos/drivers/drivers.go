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

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/cmux"
	"github.com/flanksource/gavel/todos/headless"
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
	// CodexHeadless drives codex via `codex exec --json`.
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
	case ClaudeCmux, ClaudeSDK, ClaudeHeadless, CodexHeadless:
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
	WorkDir      string
	Model        string
	Effort       string
	Plan         bool
	Resume       bool
	SessionID    string
	Timeout      time.Duration
	MaxBudgetUsd float64
	MaxTurns     int
	Tools        []string
	Dirty        bool
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
	model, err := resolveModel(kind, cfg.Model)
	if err != nil {
		return nil, "", err
	}

	switch kind.Mechanism() {
	case "cmux":
		return cmux.NewCmuxExecutor(cmux.CmuxExecutorConfig{
			WorkDir:        cfg.WorkDir,
			Model:          model,
			Effort:         cfg.Effort,
			Plan:           cfg.Plan,
			Resume:         cfg.Resume,
			SessionID:      cfg.SessionID,
			Timeout:        cfg.Timeout,
			PromptOverride: cfg.PromptOverride,
		}), "", nil
	case "sdk":
		tools := cfg.Tools
		if len(tools) == 0 {
			tools = DefaultTools()
		}
		return claude.NewClaudeExecutor(claude.ClaudeExecutorConfig{
			WorkDir:        cfg.WorkDir,
			SessionID:      cfg.SessionID,
			MaxBudgetUsd:   cfg.MaxBudgetUsd,
			MaxTurns:       cfg.MaxTurns,
			Model:          model,
			Timeout:        cfg.Timeout,
			Tools:          tools,
			Dirty:          cfg.Dirty,
			PromptOverride: cfg.PromptOverride,
		}), cfg.SessionID, nil
	case "headless":
		return headless.NewExecutor(headless.Config{
			WorkDir:        cfg.WorkDir,
			Agent:          kind.Agent(),
			Model:          model,
			Effort:         cfg.Effort,
			MaxTurns:       cfg.MaxTurns,
			Tools:          cfg.Tools,
			Timeout:        cfg.Timeout,
			PromptOverride: cfg.PromptOverride,
			Approvals:      cfg.Approvals,
		}), "", nil
	case "api":
		return nil, "", fmt.Errorf("driver %q is not yet implemented", kind)
	default:
		return nil, "", fmt.Errorf("unhandled driver mechanism %q", kind.Mechanism())
	}
}

// DefaultTools is the standard tool allowlist for the sdk/api drivers.
func DefaultTools() []string {
	return []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"}
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
	got, _ := cmux.ResolveAgent(model)
	if got != agent {
		return "", fmt.Errorf("driver %q expects a %s model but %q resolves to %s", kind, agent, model, got)
	}
	return model, nil
}

func joinKinds(kinds []Kind) string {
	parts := make([]string, len(kinds))
	for i, k := range kinds {
		parts[i] = string(k)
	}
	return strings.Join(parts, ", ")
}
