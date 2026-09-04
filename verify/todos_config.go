package verify

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

// TodosConfig configures `gavel todos run`.
type TodosConfig struct {
	// Run is the AI spec for the todo run prompt; Plan is the plan-mode spec.
	// Each overrides the base ai: spec field-wise. See prompts.TodosRun/TodosPlan.
	Run  PromptSpec `yaml:"run,omitempty" json:"run,omitempty"`
	Plan PromptSpec `yaml:"plan,omitempty" json:"plan,omitempty"`
	// Triage is the AI spec for the triage prompt: a read-only pass that compacts
	// a TODO's description and reviews its verification fixture, reporting the
	// edits for gavel to apply. See prompts.TodosTriage.
	Triage PromptSpec `yaml:"triage,omitempty" json:"triage,omitempty"`
	// CheckConcurrency bounds how many definition-of-done checks run at once
	// (`gavel todos check`, and the verification phase after a bulk triage).
	// Zero uses the built-in default; running one test suite per TODO unbounded
	// thrashes the machine.
	CheckConcurrency int `yaml:"checkConcurrency,omitempty" json:"checkConcurrency,omitempty"`
	// Verify is the spec a verification run executes as: `gavel todos check`, the
	// dashboard's verify action, and the acceptance-criteria grader inside a run's
	// definition-of-done loop. It overrides the base ai: spec field-wise and is
	// itself overridden by the request.
	//
	// It is a plain spec rather than a PromptSpec because verification has no
	// prompt template — the checklist is generated from the todo's acceptance
	// criteria — so offering a `file:` override here would be a silent no-op.
	//
	// The implementer's own run spec is deliberately NOT a layer in this chain: a
	// grader built from it inherited the session it was grading, along with the
	// coding agent's model, backend and budget.
	Verify api.Spec `yaml:"verify,omitempty" json:"verify,omitempty"`
	// Steps is the spec layer for lifecycle steps that are not built in, keyed by
	// step name: a `handoff` step a project's todos.lifecycle adds reads its
	// project configuration from `todos.steps.handoff`, exactly where `todos.run`
	// sits for the built-in run step. The four built-in steps keep their own
	// blocks above; naming one here is an error rather than a second place to
	// configure it.
	Steps map[string]api.Spec `yaml:"steps,omitempty" json:"steps,omitempty"`
	// Timeout caps a run's wall-clock duration (e.g. "30m"); empty = default.
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	// Lifecycle overrides the built-in todo lifecycle (todos/lifecycle/todos.yaml):
	// which prompt runs when, and which status its result lands the todo in.
	Lifecycle LifecycleConfig `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
	// BaseURL is the absolute origin this gavel dashboard is reachable at, e.g.
	// https://gavel.example.com. Todo bodies store attachments as server-relative
	// links, so pushing a todo to an external tracker rewrites them against this
	// origin. A loopback origin only renders for viewers on the same machine.
	BaseURL string `yaml:"baseUrl,omitempty" json:"baseUrl,omitempty"`
}

// builtinStepBlocks are the lifecycle steps configured through a typed block of
// their own rather than through todos.steps.
var builtinStepBlocks = map[string]bool{"run": true, "plan": true, "triage": true, "verify": true}

// Validate rejects a todos section that could configure one step from two
// places. A built-in step named under todos.steps would have its typed block
// and its map entry disagree with no rule saying which wins; a lifecycle that
// points at a file and also declares steps inline has the same problem.
func (c TodosConfig) Validate() error {
	for name := range c.Steps {
		if builtinStepBlocks[name] {
			return fmt.Errorf("todos.steps.%s: the built-in %s step is configured under todos.%s, not todos.steps", name, name, name)
		}
		if strings.TrimSpace(name) == "" || name != strings.ToLower(strings.TrimSpace(name)) {
			return fmt.Errorf("todos.steps: %q is not a step name; names are lower-case identifiers", name)
		}
	}
	return c.Lifecycle.Validate()
}

// LifecycleConfig is a partial todo lifecycle merged over the built-in one:
// either a File (a lifecycle YAML document relative to the project root) or an
// inline Name/Subject/Steps. Steps are kept as generic YAML here — the
// lifecycle package decodes them strictly, so verify does not depend on it.
type LifecycleConfig struct {
	File    string            `yaml:"file,omitempty" json:"file,omitempty"`
	Name    string            `yaml:"name,omitempty" json:"name,omitempty"`
	Subject map[string]string `yaml:"subject,omitempty" json:"subject,omitempty"`
	Steps   []map[string]any  `yaml:"steps,omitempty" json:"steps,omitempty"`
}

// IsZero reports whether no override is configured.
func (c LifecycleConfig) IsZero() bool {
	return c.File == "" && c.Name == "" && len(c.Subject) == 0 && len(c.Steps) == 0
}

// Validate rejects an override that is both a file and an inline definition.
// The file form used to win silently, so steps declared next to it were a
// definition nobody ran.
func (c LifecycleConfig) Validate() error {
	if c.File != "" && (c.Name != "" || len(c.Subject) > 0 || len(c.Steps) > 0) {
		return fmt.Errorf("todos.lifecycle: file (%s) and the inline name/subject/steps are mutually exclusive; move the inline definition into the file or drop file", c.File)
	}
	return nil
}
