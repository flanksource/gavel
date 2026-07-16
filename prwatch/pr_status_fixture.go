package prwatch

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
)

const PRStatusVerificationMarker = "gavel:pr-status-verification"

// PRStatusVerification describes the PR comments/actions a TODO should verify
// by running gavel pr status from its Verification fixture.
type PRStatusVerification struct {
	PRNumber   int
	Repo       string
	CommentIDs []int64
	Actions    []string
}

// UpsertPRStatusVerification inserts or replaces the generated PR status exec
// block inside the TODO's Verification fixture section.
func UpsertPRStatusVerification(body string, verification PRStatusVerification) string {
	block := BuildPRStatusVerificationBlock(verification)
	if block == "" {
		return body
	}
	return todos.UpsertMarkedVerificationBlock(body, PRStatusVerificationMarker, block)
}

// BuildPRStatusVerificationBlock returns the fixture markdown block, excluding
// the generated-block marker comments owned by UpsertPRStatusVerification.
func BuildPRStatusVerificationBlock(verification PRStatusVerification) string {
	if verification.PRNumber <= 0 {
		return ""
	}
	comments := normalizeCommentIDs(verification.CommentIDs)
	actions := normalizeSelectors(verification.Actions)
	if len(comments) == 0 && len(actions) == 0 {
		return ""
	}

	args := []string{"gavel", "pr", "status", strconv.Itoa(verification.PRNumber)}
	if repo := strings.TrimSpace(verification.Repo); repo != "" {
		args = append(args, "--repo", repo)
	}
	if len(comments) > 0 {
		parts := make([]string, 0, len(comments))
		for _, id := range comments {
			parts = append(parts, strconv.FormatInt(id, 10))
		}
		args = append(args, "--comments", strings.Join(parts, ","))
	}
	if len(actions) > 0 {
		args = append(args, "--actions", strings.Join(actions, ","))
	}

	// A plain "###" heading (not "### command:") keeps this a standalone
	// ```exec``` fence, which the fixture parser executes. A "### command:"
	// heading opens a command block whose fences must be in the codeBlocks
	// allow-list (default ["bash"]), so an ```exec``` fence there is skipped and
	// no runnable node is produced.
	return "### Verify selected PR feedback\n\n```exec\n" + shellJoin(args) + "\n```"
}

func normalizeCommentIDs(ids []int64) []int64 {
	seen := map[int64]bool{}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func normalizeSelectors(selectors []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		for _, part := range strings.Split(selector, ",") {
			part = strings.TrimSpace(part)
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			out = append(out, part)
		}
	}
	return out
}

func prRepo(pr *github.PRInfo) string {
	if pr == nil || pr.URL == "" {
		return ""
	}
	repo, _, err := github.ParsePRURL(pr.URL)
	if err != nil {
		return ""
	}
	return repo
}

func workflowActionSelector(run *github.WorkflowRun) string {
	if run == nil {
		return ""
	}
	if run.WorkflowPath != "" {
		return run.WorkflowPath
	}
	if run.Name != "" {
		return run.Name
	}
	if run.DatabaseID != 0 {
		return strconv.FormatInt(run.DatabaseID, 10)
	}
	return ""
}

func shellJoin(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}
	return strings.Join(quoted, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	for _, r := range s {
		if isShellSafe(r) {
			continue
		}
		return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
	}
	return s
}

func isShellSafe(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_-./:@%=+,", r)
}
