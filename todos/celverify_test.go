package todos

import (
	"context"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

// newTestHookContext builds a minimal HookContext good enough for celVerifier /
// evalRetry: a non-nil Request (Verify's retry path dereferences it) and a
// non-nil Response/Workspace (HookContext.Workspace() allocates lazily off
// Response, which must itself be non-nil).
func newTestHookContext() *agent.HookContext {
	return &agent.HookContext{
		Context:  context.Background(),
		Request:  &captainai.Request{},
		Response: &api.Response{Workspace: &api.Workspace{}},
	}
}

func fakeStep(res fixtures.FixtureResult) stepFixture {
	return stepFixture{
		runner: func(fixtures.FixtureTest, fixtures.RunOptions) fixtures.FixtureResult { return res },
	}
}

// The default retry predicate re-runs while the results have any errors OR
// warnings — the corrected contract (warnings were treated as pass before).
func TestEvalRetryDefault(t *testing.T) {
	hc := newTestHookContext()
	cases := []struct {
		name   string
		status task.Status
		want   bool // retry?
	}{
		{"clean passes", task.StatusPASS, false},
		{"failure retries", task.StatusFAIL, true},
		{"warning retries", task.StatusWarning, true},
		{"skip is clean", task.StatusSKIP, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			results := []fixtures.FixtureResult{{Status: tc.status}}
			retry, err := evalRetry(types.DefaultRetryExpr, results, nil, hc)
			if err != nil {
				t.Fatalf("evalRetry: %v", err)
			}
			if retry != tc.want {
				t.Errorf("retry = %t, want %t", retry, tc.want)
			}
		})
	}
}

func TestCelVerifierVerdict(t *testing.T) {
	hc := newTestHookContext()
	failing := fixtures.FixtureResult{
		Name: "checks:test", Status: task.StatusFAIL,
		Children: []*fixtures.FixtureNode{
			{Results: &fixtures.FixtureResult{Name: "TestA", Status: task.StatusFAIL, Error: "boom"}},
			{Results: &fixtures.FixtureResult{Name: "TestB", Status: task.StatusPASS}},
		},
	}
	v := &celVerifier{name: "definition-of-done", retryExpr: types.DefaultRetryExpr, steps: []stepFixture{fakeStep(failing)}}
	result, err := v.Verify(hc)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if result.Valid {
		t.Fatal("failing DoD must be invalid (retry)")
	}
	if result.Retry == nil {
		t.Fatal("an invalid verdict must propose a retry request")
	}
	feedback := result.Retry.Prompt.User
	if !strings.Contains(feedback, "TestA") || !strings.Contains(feedback, "boom") {
		t.Errorf("feedback missing the failing child: %q", feedback)
	}
	if strings.Contains(feedback, "TestB") {
		t.Errorf("feedback should omit the passing child: %q", feedback)
	}

	v.steps = []stepFixture{fakeStep(fixtures.FixtureResult{Name: "checks:test", Status: task.StatusPASS})}
	result, err = v.Verify(hc)
	if err != nil || !result.Valid {
		t.Fatalf("passing DoD must be valid, got %+v, %v", result, err)
	}
}

// A custom predicate can tolerate warnings by keying only off failures.
func TestCelCustomExprToleratesWarnings(t *testing.T) {
	hc := newTestHookContext()
	warned := []fixtures.FixtureResult{{Status: task.StatusWarning}}
	retry, err := evalRetry("results.failed > 0", warned, nil, hc)
	if err != nil {
		t.Fatalf("evalRetry: %v", err)
	}
	if retry {
		t.Error("a `results.failed > 0` predicate should tolerate warnings")
	}
}

// changed_files only binds when the hook is scoped to the agent's changed
// files (Scope == ScopeChanged); a ScopeAll run leaves it empty even when the
// workspace recorded changed files, since "changed" isn't a meaningful
// restriction when the hook is meant to act on the whole tree.
func TestCelChangedFilesBinding(t *testing.T) {
	clean := []fixtures.FixtureResult{{Status: task.StatusPASS}}

	t.Run("scope changed binds and matches", func(t *testing.T) {
		hc := newTestHookContext()
		hc.Scope = agent.ScopeChanged
		hc.Response.Workspace.Changed = []string{"a.go", "b.go"}
		retry, err := evalRetry(`changed_files.exists(f, f == "a.go")`, clean, nil, hc)
		if err != nil {
			t.Fatalf("evalRetry: %v", err)
		}
		if !retry {
			t.Error("changed_files should bind and match a.go")
		}
	})

	t.Run("scope all leaves changed_files empty", func(t *testing.T) {
		hc := newTestHookContext()
		hc.Scope = agent.ScopeAll
		hc.Response.Workspace.Changed = []string{"a.go", "b.go"}
		retry, err := evalRetry(`changed_files.exists(f, f == "a.go")`, clean, nil, hc)
		if err != nil {
			t.Fatalf("evalRetry: %v", err)
		}
		if retry {
			t.Error("changed_files must not bind outside ScopeChanged")
		}
	})
}

// The default predicate also fails when any acceptance-criteria checklist item
// is not passed, via results.checklist.all(i, i.passed).
func TestCelChecklist(t *testing.T) {
	hc := newTestHookContext()
	clean := []fixtures.FixtureResult{{Status: task.StatusPASS}}

	allPass := []map[string]any{
		{"item": "handles 5xx", "passed": true, "message": ""},
		{"item": "has tests", "passed": true, "message": ""},
	}
	retry, err := evalRetry(types.DefaultRetryExpr, clean, allPass, hc)
	if err != nil {
		t.Fatalf("evalRetry: %v", err)
	}
	if retry {
		t.Error("all criteria passed → should not retry")
	}

	mixed := []map[string]any{
		{"item": "handles 5xx", "passed": true, "message": ""},
		{"item": "has tests", "passed": false, "message": "no test added"},
	}
	retry, err = evalRetry(types.DefaultRetryExpr, clean, mixed, hc)
	if err != nil {
		t.Fatalf("evalRetry: %v", err)
	}
	if !retry {
		t.Error("an unmet criterion should retry")
	}

	// No checklist → vacuously all-passed, no retry from the checklist term.
	retry, err = evalRetry(types.DefaultRetryExpr, clean, nil, hc)
	if err != nil {
		t.Fatalf("evalRetry: %v", err)
	}
	if retry {
		t.Error("absent checklist should not force a retry")
	}
}

// A `## Verification` section holding a `yaml test` fence (a RunnerStep node)
// must dispatch to the test-step hook, not error in the type registry as it did
// before the dispatch fix.
func TestDispatchFixtureRoutesRunnerStep(t *testing.T) {
	orig := fixtures.TestStepRunner
	t.Cleanup(func() { fixtures.TestStepRunner = orig })
	called := false
	fixtures.TestStepRunner = func(f fixtures.FixtureTest, _ fixtures.RunOptions) fixtures.FixtureResult {
		called = true
		return fixtures.FixtureResult{Name: f.Name, Status: task.StatusPASS}
	}
	test := fixtures.FixtureTest{Name: "verify tests", RunnerStep: &fixtures.RunnerStepSpec{Kind: fixtures.RunnerKindTest}}
	res := dispatchFixture(context.Background(), test, fixtures.RunOptions{})
	if !called {
		t.Fatal("a runner-step fixture must dispatch to TestStepRunner, not the type registry")
	}
	if res.Status != task.StatusPASS {
		t.Errorf("status = %v, want pass", res.Status)
	}
}

// A malformed predicate must fail loud, never silently pass verification.
func TestCelMalformedErrors(t *testing.T) {
	hc := newTestHookContext()
	clean := []fixtures.FixtureResult{{Status: task.StatusPASS}}
	if _, err := evalRetry("results.failed >", clean, nil, hc); err == nil {
		t.Fatal("malformed CEL must error, not silently pass")
	}
}
