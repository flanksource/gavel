// Package drivers selects and constructs the agent driver that executes TODOs.
//
// A driver is the mechanism that drives an AI coding agent: the cmux terminal
// automation (cmux), a headless stream-json CLI (cli), the Claude Agent SDK
// bridge (sdk), or the direct provider API (api). It is a user-selectable
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

	captainai "github.com/flanksource/captain/pkg/ai"
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
	// Sdk drives the agent via the vendor Agent SDK bridge.
	Sdk Kind = "sdk"
	// Api drives the agent via the direct provider API with a local tool loop.
	Api Kind = "api"
)

// All returns every known driver mechanism in display order.
func All() []Kind {
	return []Kind{Cmux, Cli, Sdk, Api}
}

// Default is the driver used when none is specified.
const Default = Cmux

// Valid reports whether k is a known driver mechanism.
func (k Kind) Valid() bool {
	switch k {
	case Cmux, Cli, Sdk, Api:
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
	case Cmux, Cli:
		return true
	}
	return false
}

// Parse validates a driver string (case-insensitive), returning the Kind. It
// normalizes the legacy names `agent`→sdk and `headless`→cli, and accepts a
// composite agent+mechanism value in either order (`claude-cmux`, `codex-headless`,
// `cmux-claude`, `headless-codex`, …) by keeping only the mechanism part — both
// orders are persisted in prompt-run runtimes and produced by executor names, so
// this is a pragmatic migration, not a permanent alias.
func Parse(s string) (Kind, error) {
	v := strings.ToLower(strings.TrimSpace(s))
	if i := strings.IndexByte(v, '-'); i >= 0 {
		if isAgentName(v[:i]) {
			v = v[i+1:]
		} else if isAgentName(v[i+1:]) {
			v = v[:i]
		}
	}
	switch v {
	case "cmux":
		return Cmux, nil
	case "cli", "headless":
		return Cli, nil
	case "sdk", "agent":
		return Sdk, nil
	case "api":
		return Api, nil
	}
	return "", fmt.Errorf("invalid driver %q (valid: %s)", s, joinKinds(All()))
}

func isAgentName(s string) bool {
	return s == "claude" || s == "codex"
}

// New constructs the executor for a driver mechanism. The coding agent is
// derived from cfg.Spec.Name (empty → claude; a codex/gpt model → codex); the driver
// only selects the mechanism.
//
// The returned sessionID is the orchestrator session id to seed TODOExecutor
// with: empty for cmux (it mints and manages its own `--session-id`, passed in
// via Config.SessionID, so the orchestrator must not overwrite the todo's prior
// session), and Config.SessionID for the sdk path.
func New(kind Kind, cfg todos.AgentRunConfig) (todos.Executor, string, error) {
	if !kind.Valid() {
		return nil, "", fmt.Errorf("invalid driver %q (valid: %s)", kind, joinKinds(All()))
	}
	if cfg.Mode == types.ModeVerify {
		return nil, "", fmt.Errorf("verify runs through the verify engine, not an agent driver")
	}
	// The model defines the agent; the driver defines the mechanism. resolveModel
	// only guards against a stored executor identity leaking into the model field.
	model, err := resolveModel(cfg.Spec.Name)
	if err != nil {
		return nil, "", err
	}
	cfg.Spec.Name = model

	switch kind {
	case Cmux:
		// cmux drives the captain cmux provider through the same captain-backed
		// executor as the CLI path, selected by a cmux backend. It returns "" as the
		// orchestrator session id (it manages its own --session-id via Config.SessionID).
		b, err := BackendFor(model, Cmux)
		if err != nil {
			return nil, "", err
		}
		cfg.Spec.Backend = b
	case Cli:
		// The CLI (headless) path leaves the backend as configured: an explicit
		// backend from the dashboard (claude-agent/claude-cli/codex-agent) wins, and
		// an empty backend defers to captain, which resolves the concrete claude/codex
		// backend from the model+agent. BackendFor gives the target (agent, cli)
		// backends once every combination is wired through the headless streamer.
	case Sdk:
		return nil, "", fmt.Errorf("the sdk driver is not yet wired; use the cli driver (the same agent via captain)")
	case Api:
		return nil, "", fmt.Errorf("driver %q is not yet implemented", kind)
	default:
		return nil, "", fmt.Errorf("unhandled driver %q", kind)
	}
	return headless.NewExecutor(cfg), "", nil
}

// DefaultTools returns the canonical Gavel edit-capable tool set.
func DefaultTools() []string {
	return todos.DefaultAgentTools()
}

// RuntimeMode maps a driver mechanism onto captain's runtime mode. A captain
// Backend is exactly a (provider, mode) pair — which is the same statement as
// this package's "the model defines the agent; the driver defines the mechanism"
// — so the (agent, mechanism) → backend table this used to hold is captain's
// Provider.BackendFor, and BackendFor is now the only copy.
//
// Cli maps to the agent SDK, not ModeCLI: the headless path defaults an empty
// backend to claude-agent, so `--driver cli` has always meant "headless, agent
// SDK". Mapping it to ModeCLI would silently swap the binary being driven.
func (k Kind) RuntimeMode() (registry.RuntimeMode, bool) {
	switch k {
	case Cmux:
		return registry.ModeCmux, true
	case Cli, Sdk:
		return registry.ModeAgent, true
	case Api:
		return registry.ModeAPI, true
	}
	return "", false
}

// BackendFor resolves the captain backend a (model, mechanism) pair runs on,
// deriving the provider from the model itself.
func BackendFor(model string, k Kind) (captainai.Backend, error) {
	mode, ok := k.RuntimeMode()
	if !ok {
		return "", fmt.Errorf("invalid driver %q (valid: %s)", k, joinKinds(All()))
	}
	p, _, _, claimed := registry.ProviderForToken(model)
	if !claimed {
		// An unknown or empty model is claude's by convention here (see
		// claude.ResolveAgent); captain resolves the exact model later.
		p = registry.Anthropic
	}
	return p.BackendFor(mode)
}

// ForBackend is the inverse of BackendFor on its mechanism half: a captain
// backend is a (provider, mode) pair, so a configured backend already names the
// mechanism it runs on. It exists so a stated `ai.backend` outranks an unstated
// driver — picking Default there would reject a coherent config for disagreeing
// with a value nobody supplied.
func ForBackend(b captainai.Backend) (Kind, error) {
	_, mode, ok := registry.ProviderFor(b)
	if !ok {
		return "", fmt.Errorf("no driver for backend %q", b)
	}
	for _, k := range All() {
		if m, ok := k.RuntimeMode(); ok && m == mode {
			return k, nil
		}
	}
	return "", fmt.Errorf("no driver for backend %q (mode %q)", b, mode)
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
	case "cmux", "headless", "cli", "sdk", "api":
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
