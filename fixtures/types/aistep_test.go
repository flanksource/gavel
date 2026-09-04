package types

import (
	"testing"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
)

func aiStepFixture() fixtures.FixtureTest {
	return fixtures.FixtureTest{
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
}

func baseResult(f fixtures.FixtureTest) fixtures.FixtureResult {
	return fixtures.FixtureResult{Name: f.Name, Type: "verify", Test: f, Metadata: map[string]interface{}{}}
}

func allPassed() []fixtures.ChecklistResult {
	return []fixtures.ChecklistResult{
		{Item: "tests-added", Passed: true, Message: "added a table test"},
		{Item: "the endpoint streams NDJSON", Passed: true, Message: "writes ndjson chunks"},
	}
}

func TestScoreChecklistPassesWhenEveryCriterionMet(t *testing.T) {
	f := aiStepFixture()
	res := scoreChecklist(f, baseResult(f), f.AIStep.Criteria, allPassed(), time.Now())
	if res.Status != task.StatusPASS {
		t.Fatalf("status = %v, want PASS (error=%q)", res.Status, res.Error)
	}
	checklist, ok := res.Metadata["checklist"].([]fixtures.ChecklistResult)
	if !ok || len(checklist) != 2 {
		t.Fatalf("checklist metadata = %#v, want 2 entries", res.Metadata["checklist"])
	}
	if len(res.Children) != 2 {
		t.Fatalf("children = %d, want one Test per criterion", len(res.Children))
	}
}

func TestScoreChecklistFailsWhenAnyCriterionUnmet(t *testing.T) {
	f := aiStepFixture()
	verdicts := []fixtures.ChecklistResult{
		{Item: "tests-added", Passed: true},
		{Item: "the endpoint streams NDJSON", Passed: false, Message: "still returns a JSON array"},
	}
	res := scoreChecklist(f, baseResult(f), f.AIStep.Criteria, verdicts, time.Now())
	if res.Status != task.StatusFAIL {
		t.Fatalf("status = %v, want FAIL", res.Status)
	}
}

func TestScoreChecklistCELOverride(t *testing.T) {
	pass := aiStepFixture()
	pass.Expected.CEL = "json.items.all(i, i.passed)"
	if res := scoreChecklist(pass, baseResult(pass), pass.AIStep.Criteria, allPassed(), time.Now()); res.Status != task.StatusPASS {
		t.Fatalf("CEL over all-passed checklist = %v, want PASS (error=%q)", res.Status, res.Error)
	}

	fail := aiStepFixture()
	fail.Expected.CEL = "json.items.all(i, i.passed)"
	mixed := []fixtures.ChecklistResult{
		{Item: "tests-added", Passed: true},
		{Item: "the endpoint streams NDJSON", Passed: false},
	}
	if res := scoreChecklist(fail, baseResult(fail), fail.AIStep.Criteria, mixed, time.Now()); res.Status != task.StatusFAIL {
		t.Fatalf("CEL over mixed checklist = %v, want FAIL", res.Status)
	}
}

func TestAlignChecklistMarksMissingVerdictsFailed(t *testing.T) {
	items := []fixtures.ChecklistItem{{Text: "a"}, {Text: "b"}, {Text: "c"}}
	// The model answered only one criterion, matched back by its text.
	got := alignChecklist(items, []fixtures.ChecklistResult{{Item: "b", Passed: true, Message: "ok"}})
	if len(got) != len(items) {
		t.Fatalf("aligned %d, want one entry per item", len(got))
	}
	if got[0].Item != "a" || got[0].Passed {
		t.Errorf("unanswered criterion a = %#v, want failed", got[0])
	}
	if got[1].Item != "b" || !got[1].Passed {
		t.Errorf("answered criterion b = %#v, want passed", got[1])
	}
	if got[2].Item != "c" || got[2].Passed {
		t.Errorf("unanswered criterion c = %#v, want failed", got[2])
	}
}
