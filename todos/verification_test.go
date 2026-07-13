package todos

import (
	"strings"
	"testing"

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
