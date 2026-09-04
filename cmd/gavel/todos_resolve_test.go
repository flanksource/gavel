package main

import (
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// resolveCLIStep folds one step for a fresh todo in dir exactly as `todos run`
// does: the CLI's own request layer on top, dispatched from the CLI host.
func resolveCLIStep(t *testing.T, dir, step string) *lifecycle.Resolution {
	t.Helper()
	provider := testProviderFor(dir)
	todo, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Resolvable", Status: types.StatusPending})
	if err != nil {
		t.Fatalf("seed todo: %v", err)
	}
	opts := todosRunOptions()
	opts.Step = step
	prepared, err := run.Resolve(t.Context(), run.Request{Provider: provider, Todo: todo, Dir: dir, Options: opts})
	if err != nil {
		t.Fatalf("resolve %s step: %v", step, err)
	}
	return prepared.Resolution
}

// TestTodosRunResolvesModelFallbacks pins the config→effective-model wiring: a
// .gavel.yaml todos.run with a compact model and a fallbacks list resolves into
// an expanded primary whose Candidates() put it first and include the configured
// fallback. Guards the fallback-drop bug where only the model name flowed end to
// end and the fallback chain was silently discarded.
func TestTodosRunResolvesModelFallbacks(t *testing.T) {
	dir := isolatedTodosRun(t, "todos:\n  run:\n    model: \"agent:opus:high\"\n    fallbacks:\n      - sonnet\n")

	spec := resolveCLIStep(t, dir, "run").Spec
	if spec.Name != "claude-opus-5" || spec.Effort != api.EffortHigh {
		t.Fatalf("resolved model = %+v, want claude-opus-5/high", spec.Model)
	}
	cands := spec.Model.Candidates()
	if len(cands) != 2 || cands[0].Name != "claude-opus-5" || cands[1].Name != "claude-sonnet-5" {
		t.Fatalf("Candidates() = %+v, want [claude-opus-5 claude-sonnet-5]", cands)
	}
}

// TestTodosRunMergesAIBaseIntoTodosRun proves the documented `ai:` base reaches
// todo runs and that todos.run overrides it field-wise rather than replacing
// the block. The model deliberately does NOT come from `ai:` here: the run
// prompt's frontmatter declares one, and a prompt's frontmatter outranks the
// base; the base supplies what the frontmatter leaves unsaid.
func TestTodosRunMergesAIBaseIntoTodosRun(t *testing.T) {
	dir := isolatedTodosRun(t, `ai:
  budget:
    cost: 4.5
    maxTurns: 3
  permissions:
    mode: acceptEdits
todos:
  run:
    budget:
      maxTurns: 12
`)

	resolution := resolveCLIStep(t, dir, "run")
	if resolution.Spec.Budget.Cost != 4.5 {
		t.Errorf("budget.cost = %v, want the ai: base 4.5", resolution.Spec.Budget.Cost)
	}
	if resolution.Spec.Permissions.Mode != api.PermissionAcceptEdits {
		t.Errorf("permissions.mode = %q, want the ai: base acceptEdits", resolution.Spec.Permissions.Mode)
	}
	if resolution.Spec.Budget.MaxTurns != 12 {
		t.Errorf("budget.maxTurns = %d, want the todos.run override 12", resolution.Spec.Budget.MaxTurns)
	}
	// The cost survives because the ai layer is the only one that declares it —
	// the trace is the provenance, so assert against the layer stack itself.
	var costLayers []string
	for _, layer := range resolution.Trace {
		if layer.Spec.Budget.Cost != 0 {
			costLayers = append(costLayers, layer.Name)
		}
	}
	if len(costLayers) != 1 || costLayers[0] != ".gavel.yaml ai" {
		t.Errorf("budget.cost was declared by %v, want only the .gavel.yaml ai layer", costLayers)
	}
}

// TestTodosTimeoutReachesEveryStep pins .gavel.yaml todos.timeout as the one
// place a project caps a todo run, whichever step runs: the implement and plan
// steps `todos run` dispatches, and the flag on `todos check` still wins as the
// request layer.
func TestTodosTimeoutReachesEveryStep(t *testing.T) {
	dir := isolatedTodosRun(t, "ai:\n  model: claude-sonnet-5\ntodos:\n  timeout: 45m\n")

	for _, step := range []string{"run", "plan"} {
		t.Run(step, func(t *testing.T) {
			if timeout := resolveCLIStep(t, dir, step).Timeout; timeout != 45*time.Minute {
				t.Fatalf("timeout = %s, want the configured 45m", timeout)
			}
		})
	}

	provider := testProviderFor(dir)
	todo, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Checkable", Status: types.StatusPending})
	if err != nil {
		t.Fatalf("seed todo: %v", err)
	}
	// A verify step needs something to verify; the flag is the request layer,
	// so an explicitly passed --timeout still wins over the configured cap.
	todo.VerificationMarkdown = "### command: smoke\n\n```bash\necho ok\n```\n\n- contains: ok\n"
	prepared, err := run.Resolve(t.Context(), run.Request{
		Provider: provider, Todo: todo, Dir: dir,
		Options: run.Options{Step: lifecycle.StepVerify, Host: lifecycle.HostCLI, Request: api.Spec{Budget: api.Budget{Timeout: "90s"}}},
	})
	if err != nil {
		t.Fatalf("resolve verify step: %v", err)
	}
	if prepared.Resolution.Timeout != 90*time.Second {
		t.Fatalf("timeout = %s, want the explicit flag 90s", prepared.Resolution.Timeout)
	}
}
