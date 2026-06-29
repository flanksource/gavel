package types

import (
	"testing"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
)

func aiStepFixture() fixtures.FixtureTest {
	f := fixtures.FixtureTest{
		Name:      "verify feature",
		SourceDir: ".",
		Expected:  fixtures.Expectations{Properties: map[string]interface{}{}},
		AIStep: &fixtures.AIStepSpec{
			Criteria: []fixtures.ChecklistItem{
				{Text: "tests-added"},
				{Text: "the endpoint streams NDJSON"},
			},
		},
	}
	f.FrontMatter.AI = &fixtures.FixtureAIConfig{Model: "claude-code-sonnet"}
	return f
}

func TestRunAIStepImplicitThresholdPass(t *testing.T) {
	res := RunAIStep(aiStepFixture(), fixtures.RunOptions{WorkDir: "."})
	if res.Status != task.StatusPASS {
		t.Fatalf("status = %v, want PASS (error=%q)", res.Status, res.Error)
	}
}

func TestRunAIStepImplicitThresholdFail(t *testing.T) {
	t.Setenv("MOCK_VERIFY_JSON", `{"checks":{"x":{"pass":false}},"ratings":{},"completeness":{"pass":false},"implemented":false}`)
	res := RunAIStep(aiStepFixture(), fixtures.RunOptions{WorkDir: "."})
	if res.Status != task.StatusFAIL {
		t.Fatalf("status = %v, want FAIL", res.Status)
	}
}

func TestRunAIStepCELOverridePass(t *testing.T) {
	f := aiStepFixture()
	f.Expected.CEL = "json.score >= 80"
	res := RunAIStep(f, fixtures.RunOptions{WorkDir: "."})
	if res.Status != task.StatusPASS {
		t.Fatalf("status = %v, want PASS (error=%q)", res.Status, res.Error)
	}
}

func TestRunAIStepCELOverrideFail(t *testing.T) {
	f := aiStepFixture()
	// Default mock scores 100; an impossible threshold makes CEL govern a failure.
	f.Expected.CEL = "json.score >= 200"
	res := RunAIStep(f, fixtures.RunOptions{WorkDir: "."})
	if res.Status != task.StatusFAIL {
		t.Fatalf("status = %v, want FAIL", res.Status)
	}
}
