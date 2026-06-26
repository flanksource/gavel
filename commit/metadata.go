package commit

import (
	"os"
	"strings"

	"github.com/flanksource/gavel/git"
)

const (
	// EnvIssueID carries the gavel todo issue id into a `gavel commit`
	// subprocess. `gavel todos run` exports it so an agent that runs
	// `gavel commit` itself records which issue the commit belongs to.
	EnvIssueID = "GAVEL_ISSUE_ID"
	// EnvSessionID carries the agent (claude) session id the same way.
	EnvSessionID = "GAVEL_SESSION_ID"
	// EnvClaudeSessionID is the fallback session-id source used when
	// GAVEL_SESSION_ID is unset but Claude Code exported its own session id.
	EnvClaudeSessionID = "CLAUDE_SESSION_ID"

	// TrailerIssueID / trailerSessionID re-export the canonical git trailer keys
	// so commit-package consumers keep their existing reference.
	TrailerIssueID   = git.TrailerIssueID
	trailerSessionID = git.TrailerSessionID
)

// applyCommitMetadata appends git trailers identifying the gavel todo issue and
// the agent session that produced a commit. Values come from Options (set when
// gavel drives the commit in-process after a todo run) or, failing that, the
// GAVEL_ISSUE_ID / GAVEL_SESSION_ID env vars `gavel todos run` exports for an
// agent that runs `gavel commit` itself. Returns msg unchanged when metadata is
// disabled or no values are available, and never duplicates a trailer already
// present in msg.
func applyCommitMetadata(opts Options, msg string) string {
	if !opts.AddMetadata {
		return msg
	}

	issueID := firstNonEmpty(opts.IssueID, os.Getenv(EnvIssueID))
	sessionID := firstNonEmpty(opts.SessionID, os.Getenv(EnvSessionID), os.Getenv(EnvClaudeSessionID))

	var trailers []string
	if issueID != "" && !hasTrailer(msg, TrailerIssueID) {
		trailers = append(trailers, TrailerIssueID+": "+issueID)
	}
	if sessionID != "" && !hasTrailer(msg, trailerSessionID) {
		trailers = append(trailers, trailerSessionID+": "+sessionID)
	}
	if len(trailers) == 0 {
		return msg
	}

	return strings.TrimRight(msg, "\n") + "\n\n" + strings.Join(trailers, "\n")
}

// hasTrailer reports whether msg already contains a trailer line with the given
// key, so re-running metadata over an already-stamped message is idempotent.
func hasTrailer(msg, key string) bool {
	prefix := key + ":"
	for _, line := range strings.Split(msg, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
