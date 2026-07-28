package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// todosDefaults holds the `.gavel.yaml` `todos` values resolved once per
// invocation. They sit beneath the CLI flags — and, for model/budget/turns,
// beneath per-todo frontmatter — giving the precedence:
//
//	explicit CLI flag > per-todo frontmatter > .gavel.yaml todos > built-in default
//
// The zero value contributes nothing, so flags/frontmatter/built-ins apply
// exactly as before when no todos config is present.
type todosDefaults struct {
	Driver  string
	Spec    api.Spec
	GroupBy string
	Timeout time.Duration
}

// todosDef is populated by loadTodosDefaults at the start of a run; the driver
// factory (resolveDriverKind/newAgentRunConfig) reads it as the base layer.
var todosDef todosDefaults

// loadTodosDefaults resolves the merged .gavel.yaml `todos` config for workDir
// into todosDef. A malformed todos.timeout is surfaced, not swallowed.
func loadTodosDefaults(workDir string) error {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return err
	}
	tc := cfg.Todos
	spec := tc.Run.Spec
	if todosRunMode == types.ModePlan {
		spec = tc.Plan.Spec
	}
	todosDef = todosDefaults{
		Driver:  tc.Driver,
		Spec:    spec,
		GroupBy: tc.GroupBy,
	}
	timeout := strings.TrimSpace(tc.Timeout)
	if timeout == "" {
		timeout = strings.TrimSpace(spec.Budget.Timeout)
	}
	if timeout != "" {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("invalid todos timeout %q in .gavel.yaml: %w", timeout, err)
		}
		todosDef.Timeout = d
	}
	return nil
}

// runTimeout is the resolved wall-clock deadline: the .gavel.yaml todos.timeout
// when set, else the built-in 30-minute default.
func (d todosDefaults) runTimeout() time.Duration {
	if d.Timeout > 0 {
		return d.Timeout
	}
	return 30 * time.Minute
}
