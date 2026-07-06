package todos

import (
	"context"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

func fakeStep(res fixtures.FixtureResult) stepFixture {
	return stepFixture{
		runner: func(fixtures.FixtureTest, fixtures.RunOptions) fixtures.FixtureResult { return res },
	}
}

// The default retry predicate re-runs while the results have any errors OR
// warnings — the corrected contract (warnings were treated as pass before).
func TestEvalRetryDefault(t *testing.T) {
	rc := &agent.RunContext{Ctx: context.Background()}
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
			retry, err := evalRetry(types.DefaultRetryExpr, results, nil, rc, nil)
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
	rc := &agent.RunContext{Ctx: context.Background()}
	failing := fixtures.FixtureResult{
		Name: "checks:test", Status: task.StatusFAIL,
		Children: []*fixtures.FixtureNode{
			{Results: &fixtures.FixtureResult{Name: "TestA", Status: task.StatusFAIL, Error: "boom"}},
			{Results: &fixtures.FixtureResult{Name: "TestB", Status: task.StatusPASS}},
		},
	}
	v := &celVerifier{name: "definition-of-done", retryExpr: types.DefaultRetryExpr, steps: []stepFixture{fakeStep(failing)}}
	verdict, err := v.Verify(rc, nil)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if verdict.OK {
		t.Fatal("failing DoD must be not-OK (retry)")
	}
	if !strings.Contains(verdict.Feedback, "TestA") || !strings.Contains(verdict.Feedback, "boom") {
		t.Errorf("feedback missing the failing child: %q", verdict.Feedback)
	}
	if strings.Contains(verdict.Feedback, "TestB") {
		t.Errorf("feedback should omit the passing child: %q", verdict.Feedback)
	}

	v.steps = []stepFixture{fakeStep(fixtures.FixtureResult{Name: "checks:test", Status: task.StatusPASS})}
	verdict, err = v.Verify(rc, nil)
	if err != nil || !verdict.OK {
		t.Fatalf("passing DoD must be OK, got %+v, %v", verdict, err)
	}
}

// A custom predicate can tolerate warnings by keying only off failures.
func TestCelCustomExprToleratesWarnings(t *testing.T) {
	rc := &agent.RunContext{Ctx: context.Background()}
	warned := []fixtures.FixtureResult{{Status: task.StatusWarning}}
	retry, err := evalRetry("results.failed > 0", warned, nil, rc, nil)
	if err != nil {
		t.Fatalf("evalRetry: %v", err)
	}
	if retry {
		t.Error("a `results.failed > 0` predicate should tolerate warnings")
	}
}

// The retry context exposes changed_files (and session_log) for richer rules.
func TestCelChangedFilesBinding(t *testing.T) {
	rc := &agent.RunContext{Ctx: context.Background(), ChangedFiles: []string{"a.go", "b.go"}}
	clean := []fixtures.FixtureResult{{Status: task.StatusPASS}}
	retry, err := evalRetry(`changed_files.exists(f, f == "a.go")`, clean, nil, rc, nil)
	if err != nil {
		t.Fatalf("evalRetry: %v", err)
	}
	if !retry {
		t.Error("changed_files should bind and match a.go")
	}
}

// The default predicate also fails when any acceptance-criteria checklist item
// is not passed, via results.checklist.all(i, i.passed).
func TestCelChecklist(t *testing.T) {
	rc := &agent.RunContext{Ctx: context.Background()}
	clean := []fixtures.FixtureResult{{Status: task.StatusPASS}}

	allPass := []map[string]any{
		{"item": "handles 5xx", "passed": true, "message": ""},
		{"item": "has tests", "passed": true, "message": ""},
	}
	retry, err := evalRetry(types.DefaultRetryExpr, clean, allPass, rc, nil)
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
	retry, err = evalRetry(types.DefaultRetryExpr, clean, mixed, rc, nil)
	if err != nil {
		t.Fatalf("evalRetry: %v", err)
	}
	if !retry {
		t.Error("an unmet criterion should retry")
	}

	// No checklist → vacuously all-passed, no retry from the checklist term.
	retry, err = evalRetry(types.DefaultRetryExpr, clean, nil, rc, nil)
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
	rc := &agent.RunContext{Ctx: context.Background()}
	clean := []fixtures.FixtureResult{{Status: task.StatusPASS}}
	if _, err := evalRetry("results.failed >", clean, nil, rc, nil); err == nil {
		t.Fatal("malformed CEL must error, not silently pass")
	}
}
