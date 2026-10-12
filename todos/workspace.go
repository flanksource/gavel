package todos

import (
	"fmt"
	"strings"

	captainapi "github.com/flanksource/captain/pkg/api"
)

// ShortSHA is the 7-character abbreviation a commit is shown by.
func ShortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// CommitSubject is the first line of a commit message.
func CommitSubject(message string) string {
	subject, _, _ := strings.Cut(strings.TrimSpace(message), "\n")
	return strings.TrimSpace(subject)
}

// WorktreeFate is how teardown left a run's worktree: kept (with where and
// why), removed, or neither when teardown recorded no decision.
func WorktreeFate(worktree *captainapi.WorktreeState) string {
	switch {
	case worktree == nil:
		return ""
	case worktree.Kept:
		fate := fmt.Sprintf("kept at `%s`", worktree.Path)
		if reason := strings.TrimSpace(worktree.KeptReason); reason != "" {
			fate += ": " + reason
		}
		if n := len(worktree.Dirty); n > 0 {
			fate += fmt.Sprintf(" (%d dirty paths)", n)
		}
		return fate
	case worktree.Removed && worktree.BranchDeleted:
		return "removed; branch deleted"
	case worktree.Removed:
		return "removed"
	default:
		return fmt.Sprintf("`%s`", worktree.Path)
	}
}

// WorkspaceMarkdown is the attempt record's account of the run's workspace:
// the branch it worked on, the agent's own Setup..Head range, what teardown did
// with the worktree, and one line per commit. Empty for a run that reported no
// workspace.
func WorkspaceMarkdown(workspace *captainapi.WorkspaceRecord) string {
	if workspace == nil {
		return ""
	}
	var body strings.Builder
	if wt := workspace.Worktree; wt != nil {
		if wt.Branch != "" {
			fmt.Fprintf(&body, "- **Branch:** `%s`\n", wt.Branch)
		}
		if wt.Setup != "" && wt.Head != "" {
			fmt.Fprintf(&body, "- **Range:** `%s..%s`\n", ShortSHA(wt.Setup), ShortSHA(wt.Head))
		}
		if fate := WorktreeFate(wt); fate != "" {
			fmt.Fprintf(&body, "- **Worktree:** %s\n", fate)
		}
	}
	for _, commit := range workspace.Commits {
		fmt.Fprintf(&body, "- **Commit:** `%s` %s\n", ShortSHA(commit.SHA), CommitSubject(commit.Message))
	}
	return body.String()
}
