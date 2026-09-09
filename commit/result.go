package commit

import (
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/git"
)

func printDryRunPreview(result *Result) {
	if result == nil || len(result.Commits) == 0 {
		return
	}
	fmt.Fprintln(dryRunOutput, result.Pretty().ANSI())
}

// Pretty renders the commit result in a `git log`-style colorized form:
// one header line per commit (short hash or dry-run ref + conventional
// subject) followed by indented body lines. The default reflection-based
// struct printer is intentionally bypassed.
//
// Live runs (non-dry-run) return empty text: per-commit "Committed <hash>"
// lines are already logged by runSingleCommit/runCommitAll, and `--push`
// prints PR title/body separately. The trailing block this used to emit
// only restated information the user just saw.
func (r *Result) Pretty() api.Text {
	if r == nil || len(r.Commits) == 0 || !r.DryRun {
		return clicky.Text("")
	}

	t := clicky.Text("")
	summary := fmt.Sprintf("would create %d commit(s)", len(r.Commits))
	if r.PushOnly {
		summary = fmt.Sprintf("would push %d existing commit(s)", len(r.Commits))
	}
	t = t.Append("DRY RUN", "font-bold text-yellow-600").
		Append(" ", "").
		Append(summary, "text-muted").
		NewLine()

	total := len(r.Commits)
	for i, commit := range r.Commits {
		if i > 0 {
			t = t.NewLine()
		}
		t = t.Add(commit.prettyAt(i, total, r.DryRun))
	}
	return t
}

func (c CommitResult) Pretty() api.Text {
	return c.prettyAt(0, 1, c.Hash == "")
}

func (c CommitResult) prettyAt(index, total int, dryRun bool) api.Text {
	parsed := git.NewCommit(c.Message)

	ref := shortHash(c.Hash)
	if ref == "" {
		ref = fmt.Sprintf("dry-run/%d", index+1)
		if dryRun && total > 1 {
			ref = fmt.Sprintf("%s of %d", ref, total)
		}
	}

	t := clicky.Text(ref, "text-yellow-600").Space().Add(parsed.PrettySubject()).NewLine()

	if parsed.Body != "" {
		for _, line := range strings.Split(parsed.Body, "\n") {
			t = t.Append("    ", "").Append(line).NewLine()
		}
	}
	return t
}

func shortHash(h string) string {
	if len(h) > 8 {
		return h[:8]
	}
	return h
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
