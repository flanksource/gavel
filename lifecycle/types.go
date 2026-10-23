// Package lifecycle evaluates ordered workflow definitions. Domain hosts own
// subject data, execution and the meaning of outcomes; Captain runs prompts.
package lifecycle

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

// Lifecycle is one workflow definition.
type Lifecycle struct {
	Name string `yaml:"name" json:"name"`
	// Subject declares the CEL variables a step's predicates may read from the
	// todo, name → type. The declarations are strict: a predicate naming a field
	// this map does not declare fails to compile, and a host that omits a declared
	// field fails at evaluation. See SubjectTypes for the accepted type names.
	Subject  map[string]string `yaml:"subject" json:"subject"`
	Steps    []Step            `yaml:"steps" json:"steps"`
	Order    *[]string         `yaml:"order,omitempty" json:"order,omitempty"`
	Disabled *[]string         `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

// Step is one prompt run in the lifecycle.
type Step struct {
	Name string `yaml:"name" json:"name"`
	// Prompt is the captain prompt reference the step renders: a built-in name
	// such as `todos.run`, or `file:<path>` for a project-owned template.
	Prompt string `yaml:"prompt" json:"prompt"`
	// Events is the vocabulary of the envelope's `event` field: the run's
	// schema enumerates exactly these, and the outcomes read the one it chose.
	Events []string `yaml:"events" json:"events"`
	// Tools names the todo tool sets (todos/tools.Sets) the run is given.
	Tools []string `yaml:"tools,omitempty" json:"tools,omitempty"`
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

// OutcomeKeep is the status an outcome names when the run itself decided the
// todo's status — triage assigns one through todo_update, or deliberately
// leaves it alone — so the host must not write one.
const OutcomeKeep = "keep"

// Validate checks the definition's shape. Predicates are compiled by New, which
// is where a CEL error surfaces with the step and expression it belongs to.
func (l Lifecycle) Validate() error {
	if strings.TrimSpace(l.Name) == "" {
		return fmt.Errorf("lifecycle name is required")
	}
	if len(l.Subject) == 0 {
		return fmt.Errorf("lifecycle %s declares no subject", l.Name)
	}
	if _, err := NewEnv(l.Subject); err != nil {
		return fmt.Errorf("lifecycle %s: %w", l.Name, err)
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
	if err := s.validateEvents(name); err != nil {
		return err
	}
	if len(s.Outcomes) == 0 {
		return fmt.Errorf("step %s: at least one outcome is required", name)
	}
	for i, outcome := range s.Outcomes {
		if strings.TrimSpace(outcome.Status) == "" {
			return fmt.Errorf("step %s: outcome %d has no status", name, i)
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

func (s Step) validateEvents(name string) error {
	if len(s.Events) == 0 {
		return fmt.Errorf("step %s: events is required; it is the vocabulary of the envelope's event", name)
	}
	seen := map[string]bool{}
	for _, event := range s.Events {
		if strings.TrimSpace(event) == "" || event != strings.TrimSpace(event) {
			return fmt.Errorf("step %s: event %q is blank or padded", name, event)
		}
		if seen[event] {
			return fmt.Errorf("step %s: event %q is declared twice", name, event)
		}
		seen[event] = true
	}
	return nil
}
