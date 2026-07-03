package types

import (
	"testing"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/verify"
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

func baseResult(f fixtures.FixtureTest) fixtures.FixtureResult {
	return fixtures.FixtureResult{Name: f.Name, Type: "verify", Test: f, Metadata: map[string]interface{}{}}
}

func boolPtr(b bool) *bool { return &b }

func TestScoreAIStepImplicitThresholdPass(t *testing.T) {
	f := aiStepFixture()
	vr := &verify.VerifyResult{Score: 100, Implemented: boolPtr(true)}
	res := scoreAIStep(f, baseResult(f), vr, time.Now())
	if res.Status != task.StatusPASS {
		t.Fatalf("status = %v, want PASS (error=%q)", res.Status, res.Error)
	}
}

func TestScoreAIStepImplicitThresholdFail(t *testing.T) {
	f := aiStepFixture()
	// Below the default threshold (80) and not implemented → fail.
	vr := &verify.VerifyResult{Score: 40, Implemented: boolPtr(false)}
	res := scoreAIStep(f, baseResult(f), vr, time.Now())
	if res.Status != task.StatusFAIL {
		t.Fatalf("status = %v, want FAIL", res.Status)
	}
}

func TestScoreAIStepCELOverridePass(t *testing.T) {
	f := aiStepFixture()
	f.Expected.CEL = "json.score >= 80"
	vr := &verify.VerifyResult{Score: 100, Implemented: boolPtr(true)}
	res := scoreAIStep(f, baseResult(f), vr, time.Now())
	if res.Status != task.StatusPASS {
		t.Fatalf("status = %v, want PASS (error=%q)", res.Status, res.Error)
	}
}

func TestScoreAIStepCELOverrideFail(t *testing.T) {
	f := aiStepFixture()
	// A score of 100 can't satisfy an impossible threshold, so CEL fails the step
	// even though the implicit rule would have passed.
	f.Expected.CEL = "json.score >= 200"
	vr := &verify.VerifyResult{Score: 100, Implemented: boolPtr(true)}
	res := scoreAIStep(f, baseResult(f), vr, time.Now())
	if res.Status != task.StatusFAIL {
		t.Fatalf("status = %v, want FAIL", res.Status)
	}
}
