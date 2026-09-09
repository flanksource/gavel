package prwatch

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func job(name string, startedMinute int, conclusion string) github.Job {
	started := time.Date(2026, 8, 10, 12, startedMinute, 0, 0, time.UTC)
	return github.Job{
		Name:        name,
		Status:      "COMPLETED",
		Conclusion:  conclusion,
		StartedAt:   started,
		CompletedAt: started.Add(30 * time.Second),
	}
}

func runsFixture() map[int64]*github.WorkflowRun {
	return map[int64]*github.WorkflowRun{
		31: {DatabaseID: 31, Name: "CI", Status: "COMPLETED", Conclusion: "SUCCESS",
			Jobs: []github.Job{job("build", 30, "SUCCESS")}},
		12: {DatabaseID: 12, Name: "Storybook", Status: "COMPLETED", Conclusion: "SUCCESS",
			Jobs: []github.Job{job("Storybook Preview", 10, "SUCCESS")}},
		27: {DatabaseID: 27, Name: "Storybook", Status: "COMPLETED", Conclusion: "SUCCESS",
			Jobs: []github.Job{job("Delete Storybook Preview", 20, "SUCCESS")}},
	}
}

// Runs are held in a map, so rendering them in map order reshuffles the whole
// section between two invocations against the same PR.
func TestWorkflowRenderIsStableAcrossRuns(t *testing.T) {
	result := PRWatchResult{PR: &github.PRInfo{Number: 61}, Runs: runsFixture()}

	first := result.prettyWorkflows().String()
	for i := 0; i < 20; i++ {
		require.Equal(t, first, result.prettyWorkflows().String(), "workflow order changed between renders")
	}

	// Ordered by first job start: Storybook(10) < Storybook(20) < CI(30).
	assert.Less(t, strings.Index(first, "Storybook Preview"), strings.Index(first, "build"))
}

// Two runs of one workflow otherwise render under identical headings and read
// as a duplicated section.
func TestRepeatedWorkflowNamesAreDisambiguated(t *testing.T) {
	result := PRWatchResult{PR: &github.PRInfo{Number: 61}, Runs: runsFixture()}

	rendered := result.prettyWorkflows().String()

	assert.Contains(t, rendered, "Storybook (run 12)")
	assert.Contains(t, rendered, "Storybook (run 27)")
	assert.Contains(t, rendered, "CI", "a workflow that ran once keeps its plain name")
	assert.NotContains(t, rendered, "CI (run ")
}

// A rollup-only check has no jobs to expand, so a failure would otherwise be a
// bare name with nothing to act on — and it is what drives a non-zero exit.
func TestFailedRollupCheckShowsItsDetailsURL(t *testing.T) {
	const url = "https://github.com/acme/widgets/security/code-scanning"
	result := PRWatchResult{PR: &github.PRInfo{
		Number: 61,
		StatusCheckRollup: github.StatusChecks{
			{Name: "CodeQL", Status: "COMPLETED", Conclusion: "FAILURE", DetailsURL: url},
			{Name: "CodeRabbit", Status: "COMPLETED", Conclusion: "SUCCESS", DetailsURL: "https://example.com/ok"},
		},
	}}

	rendered := result.prettyWorkflows().String()

	assert.Contains(t, rendered, url, "a failing check must link somewhere actionable")
	assert.NotContains(t, rendered, "https://example.com/ok", "a passing check needs no link")
}

func TestSanitizeCommentBody(t *testing.T) {
	body := strings.Join([]string{
		"> [!CAUTION]",
		"> Review failed",
		"",
		"<details>",
		"<summary>Recent review info</summary>",
		"",
		"<details><summary>Commits</summary>",
		"reviewed 56 files",
		"</details>",
		"</details>",
		"",
		"Real content survives.",
	}, "\n")

	got := sanitizeCommentBody(body)

	assert.NotContains(t, got, "[!CAUTION]", "alert markers render literally in a terminal")
	assert.NotContains(t, got, "Recent review info", "collapsed sections are dropped, summary included")
	assert.NotContains(t, got, "reviewed 56 files", "nested collapsed sections are dropped too")
	assert.Contains(t, got, "Review failed", "the alert's own text is kept")
	assert.Contains(t, got, "Real content survives.")
}

// A review comment quoting this markup is showing it deliberately.
func TestSanitizeCommentBodyLeavesFencedExamplesAlone(t *testing.T) {
	body := strings.Join([]string{
		"<details><summary>real collapsed section</summary>",
		"hidden",
		"</details>",
		"",
		"Here is the syntax:",
		"```markdown",
		"> [!WARNING]",
		"<details><summary>example</summary>",
		"```",
	}, "\n")

	got := sanitizeCommentBody(body)

	assert.NotContains(t, got, "real collapsed section")
	assert.NotContains(t, got, "hidden")
	assert.Contains(t, got, "> [!WARNING]", "the fenced alert example is content, not scaffolding")
	assert.Contains(t, got, "<details><summary>example</summary>")
}

func TestClosedPullRequestBotNoticesAreNotActionable(t *testing.T) {
	coderabbit := "<!-- This is an auto-generated comment: failure by coderabbit.ai -->\n\n" +
		"> [!CAUTION]\n> ## Review failed\n>\n> The pull request is closed.\n"
	preview := "[PR Preview Action](https://github.com/rossjrw/pr-preview-action) v1.8.1\n:---:\n" +
		"Preview removed because the pull request was closed.\n2026-08-10 12:20 UTC\n"
	rateLimited := "<!-- This is an auto-generated comment: failure by coderabbit.ai -->\n\n" +
		"> [!CAUTION]\n> ## Review failed\n>\n> Rate limit exceeded.\n"

	kept := filterActionableComments([]github.PRComment{
		{Body: coderabbit},
		{Body: preview},
		{Body: rateLimited},
	})

	require.Len(t, kept, 1, "only the review that failed for a real reason is actionable")
	assert.Contains(t, kept[0].Body, "Rate limit exceeded.")
}
