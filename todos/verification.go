package todos

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/verify"
)

// verificationResultSection is intentionally distinct from the fixtures
// "## Verification" section (which holds verification *tests*) so writing a
// result never clobbers a TODO's test fixtures.
const verificationResultSection = "Verification Result"

// RenderVerificationSection renders an issue verification verdict as a compact
// "## Verification Result" markdown section suitable for a TODO body or an issue
// comment.
func RenderVerificationSection(result *verify.VerifyResult) string {
	var b strings.Builder
	b.WriteString("## " + verificationResultSection + "\n\n")
	fmt.Fprintf(&b, "- **Score:** %d/100\n", result.Score)
	if result.Implemented != nil {
		fmt.Fprintf(&b, "- **Implemented:** %s\n", checkbox(*result.Implemented))
	}

	if len(result.AcceptanceCriteria) > 0 {
		met := 0
		for _, c := range result.AcceptanceCriteria {
			if c.Pass {
				met++
			}
		}
		fmt.Fprintf(&b, "\n### Acceptance Criteria (%d/%d met)\n\n", met, len(result.AcceptanceCriteria))
		for _, c := range result.AcceptanceCriteria {
			suffix := ""
			if c.Comments != "" {
				suffix = " — " + c.Comments
			}
			fmt.Fprintf(&b, "- %s %s%s\n", checkbox(c.Pass), c.Criteria, suffix)
		}
	}

	var failed []verify.CheckResult
	failedIDs := make([]string, 0)
	for id, cr := range result.Checks {
		if !cr.Pass {
			failed = append(failed, cr)
			failedIDs = append(failedIDs, id)
		}
	}
	if len(failedIDs) > 0 {
		b.WriteString("\n### Failed Checks\n\n")
		for i, id := range failedIDs {
			fmt.Fprintf(&b, "- ❌ %s%s\n", id, evidenceSuffix(failed[i].Evidence))
		}
	}

	if s := strings.TrimSpace(result.Completeness.Summary); s != "" {
		fmt.Fprintf(&b, "\n### Completeness\n\n%s %s\n", checkbox(result.Completeness.Pass), s)
	}
	return b.String()
}

func checkbox(ok bool) string {
	if ok {
		return "✅"
	}
	return "❌"
}

func evidenceSuffix(evidence []verify.Evidence) string {
	if len(evidence) == 0 {
		return ""
	}
	e := evidence[0]
	loc := e.File
	if e.Line > 0 {
		loc = fmt.Sprintf("%s:%d", e.File, e.Line)
	}
	if loc != "" {
		return fmt.Sprintf(" — %s: %s", loc, e.Message)
	}
	return " — " + e.Message
}
