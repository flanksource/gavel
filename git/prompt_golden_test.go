package git

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/require"
)

// updateGolden regenerates the committed golden files instead of asserting
// against them: `go test ./git -run TestPrompt -update-golden`.
var updateGolden = flag.Bool("update-golden", false, "update prompt golden files")

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateGolden {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(path, []byte(got), 0o644))
		return
	}
	want, err := os.ReadFile(path)
	require.NoErrorf(t, err, "missing golden %s; regenerate with: go test ./git -run TestPrompt -update-golden", path)
	require.Equal(t, string(want), got)
}

func sampleCommit() models.CommitAnalysis {
	return models.CommitAnalysis{
		Commit: models.Commit{
			Hash:       "abc1234",
			Subject:    "add login endpoint",
			Body:       "Implements POST /login",
			CommitType: models.CommitTypeFeat,
			Scope:      models.ScopeType("api"),
			// Intentionally includes <, >, & to lock in that the diff is emitted
			// raw (dotprompt renders with NoEscape) rather than HTML-escaped.
			Patch: "diff --git a/login.go b/login.go\n+func Login() bool { return a < b && c > d }\n",
		},
	}
}

func TestPromptCommitMessage_WithMaxBodyLines(t *testing.T) {
	got, err := renderCommitPrompt(sampleCommit(), commitMessagePrompt, map[string]any{"maxBodyLines": 3})
	require.NoError(t, err)
	require.Contains(t, got, "- body: at most 3 line(s)", "maxBodyLines must select the if-branch")
	require.NotContains(t, got, "&lt;", "diff must not be HTML-escaped")
	require.Contains(t, got, "a < b && c > d", "raw diff content must be preserved")
	assertGolden(t, "commit-message-with-maxbody.golden", got)
}

func TestPromptCommitMessage_WithoutMaxBodyLines(t *testing.T) {
	got, err := renderCommitPrompt(sampleCommit(), commitMessagePrompt, map[string]any{"maxBodyLines": 0})
	require.NoError(t, err)
	require.Contains(t, got, "- body: omit unless the change is non-trivial", "zero maxBodyLines must select the else-branch")
	require.NotContains(t, got, "at most", "else-branch must not mention a line cap")
	assertGolden(t, "commit-message-without-maxbody.golden", got)
}

func TestPromptFunctionalityRemoved(t *testing.T) {
	got, err := renderCommitPrompt(sampleCommit(), functionalityRemovedPrompt, nil)
	require.NoError(t, err)
	require.Contains(t, got, "a < b && c > d", "raw diff content must be preserved")
	assertGolden(t, "functionality-removed.golden", got)
}

func TestPromptCompatibilityIssues(t *testing.T) {
	got, err := renderCommitPrompt(sampleCommit(), compatibilityIssuesPrompt, nil)
	require.NoError(t, err)
	require.Contains(t, got, "a < b && c > d", "raw diff content must be preserved")
	assertGolden(t, "compatibility-issues.golden", got)
}

func TestPromptSummaryGroup(t *testing.T) {
	commits := models.CommitAnalyses{
		{
			Commit: models.Commit{
				Hash:       "abc1234",
				Subject:    "add login endpoint",
				Body:       "Implements POST /login",
				CommitType: models.CommitTypeFeat,
				Scope:      models.ScopeType("api"),
			},
			Changes: models.Changes{{File: "api/login.go"}, {File: "api/auth.go"}},
		},
		{
			Commit: models.Commit{
				Hash:       "def5678",
				Subject:    "fix nil panic on logout",
				CommitType: models.CommitTypeFix,
				Scope:      models.ScopeTypeUnknown, // empty scope exercises {{#if this.scope}} else
			},
			Changes: models.Changes{{File: "api/login.go"}}, // duplicate file exercises dedup
		},
	}

	got, err := renderSummaryPrompt(models.ScopeType("api"), "last 7 days", commits)
	require.NoError(t, err)

	require.Contains(t, got, "abc1234: feat(api): add login endpoint", "scoped commit line must render scope in parens")
	require.Contains(t, got, "def5678: fix: fix nil panic on logout", "unscoped commit must omit the parens")
	require.Contains(t, got, "- api/auth.go", "deduped+sorted files must be listed")
	require.Contains(t, got, "- api/login.go", "deduped+sorted files must be listed")
	require.Equal(t, 1, strings.Count(got, "- api/login.go"), "duplicate file must appear once")
	assertGolden(t, "summary-group.golden", got)
}
