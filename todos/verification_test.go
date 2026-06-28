package todos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

func sampleResult(implemented bool, score int) *verify.VerifyResult {
	return &verify.VerifyResult{
		Score:       score,
		Implemented: &implemented,
		AcceptanceCriteria: []verify.CriterionResult{
			{Criteria: "Streams NDJSON for large payloads", Pass: true},
			{Criteria: "Returns 400 on invalid input", Pass: false, Comments: "handler.go:42: no validation"},
		},
		Completeness: verify.CompletenessResult{Pass: false, Summary: "missing tests"},
	}
}

func TestRenderVerificationSection(t *testing.T) {
	out := RenderVerificationSection(sampleResult(false, 64))
	for _, want := range []string{
		"## Verification Result",
		"**Score:** 64/100",
		"**Implemented:** ❌",
		"Acceptance Criteria (1/2 met)",
		"Streams NDJSON for large payloads",
		"handler.go:42: no validation",
		"missing tests",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered section missing %q:\n%s", want, out)
		}
	}
}

func TestUpdateVerificationSectionReplacesInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todo.md")
	if err := os.WriteFile(path, []byte("Issue body.\n\n## Attempts\n\n| # |\n"), 0644); err != nil {
		t.Fatal(err)
	}
	todo := &types.TODO{FilePath: path}

	if err := UpdateVerificationSection(todo, sampleResult(true, 90)); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := UpdateVerificationSection(todo, sampleResult(false, 50)); err != nil {
		t.Fatalf("second write: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	if n := strings.Count(got, "## Verification Result"); n != 1 {
		t.Errorf("expected exactly one verification section, got %d:\n%s", n, got)
	}
	if !strings.Contains(got, "**Score:** 50/100") {
		t.Errorf("section should reflect the latest result:\n%s", got)
	}
	// Inserted above the running history, not clobbering it.
	if strings.Index(got, "## Verification Result") > strings.Index(got, "## Attempts") {
		t.Errorf("verification result should precede Attempts:\n%s", got)
	}
}
