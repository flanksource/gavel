// Package lifecycle is the todo lifecycle: an ordered list of steps, each a
// captain prompt reference plus CEL predicates saying when the step applies and
// which status its result lands the todo in. The definition is data (todos.yaml,
// overridable from .gavel.yaml), the engine evaluates it, and the host adapter
// (host.go) is the only thing that knows what a todo is.
//
// Captain runs prompts; gavel decides which prompt runs next and what its
// result means. Both halves of that decision live here so there is exactly one
// place that answers "what happens to this todo now".
package lifecycle

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
)

// Lifecycle is one workflow definition.
type Lifecycle struct {
	Name string `yaml:"name" json:"name"`
	// Subject declares the CEL variables a step's predicates may read from the
	// todo, name → type. The declarations are strict: a predicate naming a field
	// this map does not declare fails to compile, and a host that omits a declared
	// field fails at evaluation. See SubjectTypes for the accepted type names.
	Subject map[string]string `yaml:"subject" json:"subject"`
	Steps   []Step            `yaml:"steps" json:"steps"`
}

// Step is one prompt run in the lifecycle.
type Step struct {
	Name string `yaml:"name" json:"name"`
	// Prompt is the captain prompt reference the step renders: a built-in name
	// such as `todos.run`, or `file:<path>` for a project-owned template.
	Prompt string `yaml:"prompt" json:"prompt"`
	// Envelope is the structured result the prompt returns — result, plan or
	// triage — which decides the schema the run is asked for and how the host
	// applies the result. Empty means result.
	Envelope EnvelopeKind `yaml:"envelope,omitempty" json:"envelope,omitempty"`
	// When is the applicability predicate over {subject, runs, last}. Empty means
	// the step always applies.
	When string `yaml:"when,omitempty" json:"when,omitempty"`
	// Inputs map template variables to CEL expressions over the same variables
	// When reads, so a prompt's inputs are declared next to the prompt.
	Inputs map[string]string `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	// Spec is the step's own spec layer — permissions, setup, workflow — folded
	// between the prompt's frontmatter and the project's `todos.<step>` block.
	Spec *api.Spec `yaml:"spec,omitempty" json:"spec,omitempty"`
	// Outcomes map a finished run onto a status: ordered, first true wins, none
	// true is an error. The host writes the status; nothing else does.
	Outcomes []Outcome `yaml:"outcomes" json:"outcomes"`
	// Auxiliary steps are never chosen by Next: they run only when asked for by
	// name (triage), so a backlog pass does not sit in front of implementation.
	Auxiliary bool `yaml:"auxiliary,omitempty" json:"auxiliary,omitempty"`
}

// Outcome is one status transition: Status is written when When holds.
type Outcome struct {
	Status string `yaml:"status" json:"status"`
	When   string `yaml:"when" json:"when"`
}

// EnvelopeKind names the structured result a prompt returns.
type EnvelopeKind string

const (
	EnvelopeResult EnvelopeKind = "result"
	EnvelopePlan   EnvelopeKind = "plan"
	EnvelopeTriage EnvelopeKind = "triage"
)

// OutcomeKeep is the status an outcome names when the run itself decided the
// todo's status — triage assigns one through its verdict, or deliberately
// leaves it alone — so the host must not write one.
const OutcomeKeep = "keep"

// StepVerify is the step name every lifecycle must define: the verify-only
// pass `gavel todos check` and the dashboard's verify action run.
const StepVerify = "verify"

// SubjectTypes are the type names a subject declaration may use.
var SubjectTypes = []string{"string", "bool", "int", "double", "dyn", "list<string>", "list<dyn>", "map<string,dyn>"}

var validStatuses = map[string]bool{
	string(types.StatusDraft): true, string(types.StatusPending): true, string(types.StatusInProgress): true,
	string(types.StatusReview): true, string(types.StatusAsk): true, string(types.StatusVerified): true,
	string(types.StatusUnverified): true, string(types.StatusCompleted): true, string(types.StatusFailed): true,
	OutcomeKeep: true,
}

// Validate checks the definition's shape. Predicates are compiled by New, which
// is where a CEL error surfaces with the step and expression it belongs to.
func (l Lifecycle) Validate() error {
	if strings.TrimSpace(l.Name) == "" {
		return fmt.Errorf("lifecycle name is required")
	}
	if len(l.Subject) == 0 {
		return fmt.Errorf("lifecycle %s declares no subject", l.Name)
	}
	for name, typ := range l.Subject {
		if _, err := celType(typ); err != nil {
			return fmt.Errorf("lifecycle %s subject %s: %w", l.Name, name, err)
		}
	}
	if len(l.Steps) == 0 {
		return fmt.Errorf("lifecycle %s declares no steps", l.Name)
	}
	seen := map[string]bool{}
	for _, step := range l.Steps {
		if err := step.validate(); err != nil {
			return fmt.Errorf("lifecycle %s: %w", l.Name, err)
		}
		if seen[step.Name] {
			return fmt.Errorf("lifecycle %s: step %q is declared twice", l.Name, step.Name)
		}
		seen[step.Name] = true
	}
	if !seen[StepVerify] {
		return fmt.Errorf("lifecycle %s has no %q step; verification must be a step of every lifecycle", l.Name, StepVerify)
	}
	return nil
}

func (s Step) validate() error {
	name := strings.TrimSpace(s.Name)
	if name == "" {
		return fmt.Errorf("step with no name")
	}
	if name != strings.ToLower(name) || strings.ContainsAny(name, " \t/") {
		return fmt.Errorf("step %q: names are lower-case identifiers", s.Name)
	}
	if strings.TrimSpace(s.Prompt) == "" {
		return fmt.Errorf("step %s: prompt is required", name)
	}
	// The verify step runs the definition of done, not an agent prompt: a
	// template named here would be resolved and then silently ignored, so the
	// only reference it accepts is the built-in one.
	if name == StepVerify && strings.TrimSpace(s.Prompt) != promptRefPrefix+StepVerify {
		return fmt.Errorf("step %s: prompt %q is not supported; the verify step runs the definition of done and must reference %s",
			name, s.Prompt, promptRefPrefix+StepVerify)
	}
	switch s.Envelope {
	case "", EnvelopeResult, EnvelopePlan, EnvelopeTriage:
	default:
		return fmt.Errorf("step %s: envelope %q is not one of result, plan, triage", name, s.Envelope)
	}
	if len(s.Outcomes) == 0 {
		return fmt.Errorf("step %s: at least one outcome is required", name)
	}
	for i, outcome := range s.Outcomes {
		if !validStatuses[outcome.Status] {
			return fmt.Errorf("step %s: outcome %d status %q is not a todo status", name, i, outcome.Status)
		}
		if strings.TrimSpace(outcome.When) == "" {
			return fmt.Errorf("step %s: outcome %d (%s) has no predicate", name, i, outcome.Status)
		}
	}
	return nil
}

// StepNames are the declared step names in definition order — the vocabulary a
// caller naming a step chooses from.
func (l Lifecycle) StepNames() []string {
	names := make([]string, len(l.Steps))
	for i, step := range l.Steps {
		names[i] = step.Name
	}
	return names
}

// Step returns the named step.
func (l Lifecycle) Step(name string) (Step, bool) {
	for _, step := range l.Steps {
		if step.Name == name {
			return step, true
		}
	}
	return Step{}, false
}

// EnvelopeOrDefault is the step's envelope with the default applied.
func (s Step) EnvelopeOrDefault() EnvelopeKind {
	if s.Envelope == "" {
		return EnvelopeResult
	}
	return s.Envelope
}
