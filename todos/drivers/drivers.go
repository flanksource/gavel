// Package drivers selects and constructs the agent driver that executes TODOs.
//
// A driver is the mechanism that drives an AI coding agent: the cmux terminal
// automation (cmux), a supervised CLI (cli), a local agent bridge (agent), or
// the direct provider API (api). It is a user-selectable
// dimension alongside model and effort. The coding agent itself (claude or
// codex) is NOT part of the driver — it is derived from the model
// (claude.ResolveAgent). This package is the single registry both the CLI and
// the dashboard delegate to, so the selection logic lives in one place instead
// of duplicated `switch mode` blocks.
//
// It lives below todos/ (importing todos, todos/cmux, todos/claude) rather than
// in package todos itself, because the concrete executors already import todos —
// putting the factory here keeps the dependency graph acyclic.
package drivers

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api/registry"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/headless"
	"github.com/flanksource/gavel/todos/types"
)

// Kind identifies the execution mechanism that drives the coding agent. It is
// mechanism-only: the agent (claude or codex) is derived from the model, not the
// driver.
type Kind string

const (
	// Cmux drives the agent's interactive TUI inside a cmux surface.
	Cmux Kind = "cmux"
	// Cli drives the agent via a headless `-p --output-format stream-json` CLI.
	Cli Kind = "cli"
	// Agent drives the model via Captain's agent runtime.
	Agent Kind = "agent"
	// Api drives the agent via the direct provider API with a local tool loop.
	Api Kind = "api"
)

// All returns every known driver mechanism in display order.
func All() []Kind {
	return []Kind{Api, Agent, Cli, Cmux}
}

// Default is the driver used when none is specified.
const Default = Cmux

// Valid reports whether k is a known driver mechanism.
func (k Kind) Valid() bool {
	switch k {
	case Cmux, Cli, Agent, Api:
		return true
	}
	return false
}

// Mechanism returns the mechanism string. Kind is mechanism-only, so this is the
// Kind itself — kept for call sites that compare against a literal ("cmux").
func (k Kind) Mechanism() string {
	return string(k)
}

// Implemented reports whether New can construct an executor for this mechanism
// today. Unimplemented mechanisms are still offered in pickers so they appear as
// the work lands, but selecting one returns a clear error rather than silently
// falling back to another driver.
func (k Kind) Implemented() bool {
	switch k {
	case Cmux, Cli, Agent, Api:
		return true
	}
	return false
}

// Parse validates one canonical backend. Provider names, composite adapter ids,
// and old aliases are invalid rather than translated.
func Parse(s string) (Kind, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "cmux":
		return Cmux, nil
	case "cli":
		return Cli, nil
	case "agent":
		return Agent, nil
	case "api":
		return Api, nil
	}
	return "", fmt.Errorf("invalid driver %q (valid: %s)", s, joinKinds(All()))
}

// New constructs the executor for a driver mechanism. The coding agent is
// derived from cfg.Spec.Name (empty → claude; a codex/gpt model → codex); the driver
// only selects the mechanism.
//
// The returned sessionID is the orchestrator session id to seed TODOExecutor
// with: empty for cmux (it mints and manages its own `--session-id`, passed in
// via Config.SessionID, so the orchestrator must not overwrite the todo's prior
// session), and Config.SessionID for the agent path.
func New(kind Kind, cfg todos.AgentRunConfig) (todos.Executor, string, error) {
	if !kind.Valid() {
		return nil, "", fmt.Errorf("invalid driver %q (valid: %s)", kind, joinKinds(All()))
	}
	if cfg.Mode == types.ModeVerify {
		return nil, "", fmt.Errorf("verify runs through the verify engine, not an agent driver")
	}
	// The model defines the provider and the driver is the canonical backend.
	model, err := resolveModel(cfg.Spec.Name)
	if err != nil {
		return nil, "", err
	}
	cfg.Spec.Name = model
	if cfg.Spec.Mode == "" {
		cfg.Spec.Mode = registry.RuntimeMode(kind)
	}
	resolved, err := registry.ResolveModel(cfg.Spec.Model)
	if err != nil {
		return nil, "", err
	}
	cfg.Spec.Model = resolved
	kind, err = Parse(string(resolved.Mode))
	if err != nil {
		return nil, "", fmt.Errorf("resolved model backend: %w", err)
	}

	switch kind {
	case Cmux:
		// cmux manages its own provider session id.
	case Cli, Agent, Api:
	default:
		return nil, "", fmt.Errorf("unhandled driver %q", kind)
	}
	return headless.NewExecutor(cfg), "", nil
}

// DefaultTools returns the canonical Gavel edit-capable tool set.
func DefaultTools() []string {
	return todos.DefaultAgentTools()
}

// RuntimeMode is a direct type conversion because Gavel and Captain share the
// same canonical backend values.
func (k Kind) RuntimeMode() (registry.RuntimeMode, bool) {
	if !k.Valid() {
		return "", false
	}
	return registry.RuntimeMode(k), true
}

// resolveModel validates the requested model. The agent is derived from the model
// (claude.ResolveAgent) elsewhere, so there is no agent to reconcile here — the
// only guard is against a stored executor identity leaking into the model field.
func resolveModel(model string) (string, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return "", nil
	}
	// A driver/executor identity (e.g. "cmux-claude", "headless-codex") is never a
	// model. Guard against it round-tripping through storage into the model field,
	// which would otherwise launch `--model cmux-claude` and fail at the CLI.
	if isExecutorIdentity(model) {
		return "", fmt.Errorf("%q is a driver/executor identity, not a model", model)
	}
	return model, nil
}

// isExecutorIdentity reports whether s has the "<mechanism>-<agent>" shape of an
// executor Name() (e.g. "cmux-claude") — the reverse of a driver Kind — so such a
// value can never be mistaken for an LLM model.
func isExecutorIdentity(s string) bool {
	mechanism, agentName, ok := strings.Cut(s, "-")
	if !ok {
		return false
	}
	switch mechanism {
	case "cmux", "headless", "cli", "sdk", "agent", "api":
	default:
		return false
	}
	return agentName == "claude" || agentName == "codex"
}

func joinKinds(kinds []Kind) string {
	parts := make([]string, len(kinds))
	for i, k := range kinds {
		parts[i] = string(k)
	}
	return strings.Join(parts, ", ")
}
