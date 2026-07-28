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
	api.Spec
	WorkDir       string
	Mode          types.RunMode
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
