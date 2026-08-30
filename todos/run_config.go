package todos

import (
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
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
	// Mode is the behaviour class: whether this run commits and verifies.
	Mode types.RunMode
	// Prompt is the name of the template being run. Several prompts share one
	// Mode, so this — not Mode — identifies which instructions the agent got.
	Prompt string
	// Envelope is the structured result the prompt returns. It decides which type
	// the run's final output is parsed into.
	Envelope todoprompt.EnvelopeKind
	// Template is the .gavel.yaml prompt override source resolved by todos/spec
	// alongside Spec; empty renders the prompt's embedded default. It travels with
	// the spec so the executor does not re-read .gavel.yaml and risk resolving a
	// different override than the one the spec was built from.
	Template     string
	ExistingPlan string
	// Backlog is the compact index of other open TODOs a triage run uses to spot
	// duplicates; empty omits the section.
	Backlog       string
	Verifiers     []agent.Verify
	MaxIterations int
	Resume        bool
	Approvals     bool
}

// PromptOptions projects the run configuration onto the prompt renderer's
// options. Both the executor and `--dry-run` render through it, so a preview
// cannot silently disagree with what the run will actually send — which it did
// when each built the options itself and only one learned about a new field.
//
// workDir is passed in because the executor renders against the group's resolved
// working directory, not the configured root.
func (c AgentRunConfig) PromptOptions(workDir string) todoprompt.Options {
	return todoprompt.Options{
		WorkDir:      workDir,
		Prompt:       c.Prompt,
		Envelope:     c.Envelope,
		Mode:         c.Mode,
		Spec:         c.Spec,
		Template:     c.Template,
		ExistingPlan: c.ExistingPlan,
		Backlog:      c.Backlog,
	}
}

// DefaultAgentTools is the standard edit-capable tool set used when a run does
// not declare an explicit per-tool policy.
func DefaultAgentTools() []string {
	return []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"}
}
