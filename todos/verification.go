package todos

import (
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// verificationResultSection is intentionally distinct from the fixtures
// "## Verification" section (which holds verification *tests*) so writing a
// result never clobbers a TODO's test fixtures.
const verificationResultSection = "Verification Result"

// verificationInsertBefore keeps the result section above the running history.
var verificationInsertBefore = []string{"Attempts", "Failure History"}

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

// UpdateVerificationSection rewrites the "## Verification Result" section in a
// file-backed TODO, replacing any prior result in place (atomic write).
func UpdateVerificationSection(todo *types.TODO, result *verify.VerifyResult) error {
	content, err := os.ReadFile(todo.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read TODO file: %w", err)
	}
	updated := ReplaceOrAppendSection(string(content), verificationResultSection, RenderVerificationSection(result), verificationInsertBefore...)

	tmp := todo.FilePath + ".tmp"
	if err := os.WriteFile(tmp, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmp, todo.FilePath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}
