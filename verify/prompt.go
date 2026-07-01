package verify

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
)

//go:embed verify-prompt.prompt
var verifyPromptTemplate string

// renderPrompt builds the reviewer prompt. The checks, rating dimensions, and
// acceptance criteria are carried by the JSON output schema (see BuildSchema),
// so the template only needs the scope instruction and issue context — keeping
// it simple enough for dotprompt (no eq/index/range-over-checks logic).
func renderPrompt(scope ReviewScope, cfg VerifyConfig, issue *IssueContext) (string, error) {
	data := map[string]any{
		"scopeInstruction": scopeInstruction(scope),
		"extraPrompt":      cfg.Prompt,
	}
	if issue != nil {
		im := map[string]any{
			"title":       issue.Title,
			"description": issue.Description,
			"sessionId":   issue.SessionID,
		}
		if len(issue.Comments) > 0 {
			comments := make([]map[string]any, 0, len(issue.Comments))
			for _, c := range issue.Comments {
				comments = append(comments, map[string]any{"author": c.Author, "body": c.Body})
			}
			im["comments"] = comments
		}
		data["issue"] = im
	}

	req, _, err := prompt.Load(verifyPromptTemplate).Render(data, nil)
	if err != nil {
		return "", fmt.Errorf("render verify prompt: %w", err)
	}
	return req.Prompt, nil
}

// scopeInstruction returns the reviewer's how-to-obtain-the-diff instruction for
// a scope. Computing it in Go (rather than a per-type template branch) keeps the
// prompt template trivial.
func scopeInstruction(scope ReviewScope) string {
	switch scope.Type {
	case "range":
		return fmt.Sprintf("Review the changes in the commit range: %s\nUse `git diff %s` to see the changes.", scope.CommitRange, scope.CommitRange)
	case "commit":
		return fmt.Sprintf("Review the changes introduced by commit %s.\nUse `git show %s` to see the diff.", scope.Commit, scope.Commit)
	case "commits":
		var b strings.Builder
		b.WriteString("Review the changes introduced by these commits (run `git show <sha>` for each):\n")
		for _, c := range scope.Commits {
			fmt.Fprintf(&b, "- %s\n", c)
		}
		return strings.TrimRight(b.String(), "\n")
	case "branch":
		return fmt.Sprintf("Review the changes between branch `%s` and the current branch.\nUse `git diff %s...HEAD` to see the changes.", scope.Branch, scope.Branch)
	case "pr":
		return fmt.Sprintf("Review the changes in PR #%d.\nUse `gh pr diff %d` to get the diff.", scope.PRNumber, scope.PRNumber)
	case "date-range":
		return fmt.Sprintf("Review commits between %s and %s.\nUse `git log --after=%q --before=%q --oneline` to list them, then diff that range.", scope.Since, scope.Until, scope.Since, scope.Until)
	case "files":
		var b strings.Builder
		b.WriteString("Review the following files:\n")
		for _, f := range scope.Files {
			fmt.Fprintf(&b, "- %s\n", f)
		}
		return strings.TrimRight(b.String(), "\n")
	default: // "diff" and any unknown scope
		return "Run `git diff HEAD` to see the uncommitted changes and review them."
	}
}
