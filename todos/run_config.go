package todos

import (
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
)

// AgentRunConfig carries Captain's canonical run Spec alongside the
// Gavel-only orchestration state that cannot be serialized into an agent
// request.
type AgentRunConfig struct {
	// Spec is a named field, not embedded: api.Spec's value-receiver
	// MarshalJSON/MarshalYAML would otherwise be promoted onto this type and
	// emit only the spec, dropping every field below it.
	Spec    api.Spec
	WorkDir string
	Mode    types.RunMode
	// Template is the .gavel.yaml prompt override source resolved by todos/spec
	// alongside Spec; empty renders the mode's embedded default. It travels with
	// the spec so the executor does not re-read .gavel.yaml and risk resolving a
	// different override than the one the spec was built from.
	Template      string
	ExistingPlan  string
	Verifiers     []agent.Verify
	MaxIterations int
	Resume        bool
	Approvals     bool
}

// DefaultAgentTools is the standard edit-capable tool set used when a run does
// not declare an explicit per-tool policy.
func DefaultAgentTools() []string {
	return []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"}
}
