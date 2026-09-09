package prwatch

import (
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResultFiltersCommentsMatchIDsAndAuthors(t *testing.T) {
	result := &PRWatchResult{Comments: []github.PRComment{
		{ID: 1, Author: "alice", Body: "first"},
		{ID: 2, Author: "coderabbitai[bot]", BotType: "coderabbit", Body: "second"},
		{ID: 3, Author: "bob", Body: "third"},
		{ID: 4, Author: "carol", Body: "fourth"},
	}}

	filters := newResultFilters([]string{"[1,2,!3,*,!@coderabbit]"}, nil)
	filters.apply(result)

	require.Len(t, result.Comments, 2)
	assert.Equal(t, int64(1), result.Comments[0].ID)
	assert.Equal(t, int64(4), result.Comments[1].ID)
}

func TestResultFiltersCommentsCanMatchBotAliasOnly(t *testing.T) {
	result := &PRWatchResult{Comments: []github.PRComment{
		{ID: 1, Author: "alice", Body: "first"},
		{ID: 2, Author: "coderabbitai[bot]", BotType: "coderabbit", Body: "second"},
	}}

	filters := newResultFilters([]string{"@coderabbit"}, nil)
	filters.apply(result)

	require.Len(t, result.Comments, 1)
	assert.Equal(t, int64(2), result.Comments[0].ID)
}

func TestResultFiltersActionsMatchRunWorkflowPathAndName(t *testing.T) {
	for _, tc := range []struct {
		name    string
		filter  string
		wantIDs []int64
	}{
		{name: "run id", filter: "101", wantIDs: []int64{101}},
		{name: "workflow id", filter: "11", wantIDs: []int64{101}},
		{name: "workflow path", filter: ".github/workflows/ci.yml", wantIDs: []int64{101}},
		{name: "workflow basename", filter: "ci.yml", wantIDs: []int64{101}},
		{name: "workflow name", filter: "CI", wantIDs: []int64{101}},
		{name: "wildcard exclusion", filter: "*,!deploy", wantIDs: []int64{101, 303}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := sampleActionFilterResult()
			filters := newResultFilters(nil, []string{tc.filter})
			filters.apply(result)

			assert.ElementsMatch(t, tc.wantIDs, runIDs(result.Runs))
			assert.ElementsMatch(t, tc.wantIDs, statusCheckRunIDs(t, result.PR.StatusCheckRollup))
		})
	}
}

func TestResultFiltersActionsPruneHiddenFailureFromExitCode(t *testing.T) {
	result := sampleActionFilterResult()

	filters := newResultFilters(nil, []string{"ci.yml"})
	filters.apply(result)

	require.Len(t, result.PR.StatusCheckRollup, 1)
	assert.Equal(t, "SUCCESS", result.PR.StatusCheckRollup[0].Conclusion)
	assert.Equal(t, 0, statusExitCode(result))
}

// This replaces TestResultFiltersActionNoMatchIsDoneForFollow, which asserted
// that an --actions selector matching nothing made --follow "done". That was a
// false green twice over: a typo returned exit 0, and so did every poll issued
// before GitHub had registered the freshly-pushed commit's check runs. An empty
// filtered set is now never complete, and the reason the watch stops is the hard
// error, not completion.
func TestResultFiltersActionNoMatchIsNotCompleteForFollow(t *testing.T) {
	result := sampleActionFilterResult()
	preChecks, preRuns := len(result.PR.StatusCheckRollup), len(result.Runs)

	filters := newResultFilters(nil, []string{"missing-action"})
	filters.apply(result)

	assert.Empty(t, result.Runs)
	assert.Empty(t, result.PR.StatusCheckRollup)
	assert.False(t, filters.isComplete(result), "an empty filtered rollup must never satisfy the gate")
	assert.True(t, filters.noActionMatch(preChecks, preRuns, result))
}

func TestResultFiltersIsCompleteWithoutFilters(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*PRWatchResult)
		want   bool
	}{
		{"all completed", func(*PRWatchResult) {}, true},
		{"one in progress", func(r *PRWatchResult) {
			r.PR.StatusCheckRollup[1].Status = "IN_PROGRESS"
		}, false},
		{"no checks registered yet keeps polling", func(r *PRWatchResult) {
			r.PR.StatusCheckRollup = github.StatusChecks{}
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := sampleActionFilterResult()
			tc.mutate(result)
			filters := newResultFilters(nil, nil)
			filters.apply(result)
			assert.Equal(t, tc.want, filters.isComplete(result))
		})
	}
}

func TestResultFiltersIsCompleteWithActionFilterIgnoresUnselectedChecks(t *testing.T) {
	t.Run("an in-flight check outside the selector does not hold the gate", func(t *testing.T) {
		result := sampleActionFilterResult()
		result.PR.StatusCheckRollup[1].Status = "IN_PROGRESS"

		filters := newResultFilters(nil, []string{"ci.yml"})
		filters.apply(result)

		assert.True(t, filters.isComplete(result))
	})

	t.Run("an in-flight check inside the selector does", func(t *testing.T) {
		result := sampleActionFilterResult()
		result.PR.StatusCheckRollup[1].Status = "IN_PROGRESS"

		filters := newResultFilters(nil, []string{"deploy"})
		filters.apply(result)

		assert.False(t, filters.isComplete(result))
	})
}

func TestResultFiltersIsCompleteWithCommentFilterWaitsForResolution(t *testing.T) {
	cases := []struct {
		name       string
		isResolved bool
		isOutdated bool
		want       bool
	}{
		{"unresolved thread holds the gate", false, false, false},
		{"resolved releases it", true, false, true},
		{"outdated releases it", false, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := sampleCommentGateResult()
			result.Comments[0].IsResolved = tc.isResolved
			result.Comments[0].IsOutdated = tc.isOutdated

			filters := newResultFilters([]string{"@coderabbit"}, nil)
			filters.apply(result)

			assert.Equal(t, tc.want, filters.isComplete(result))
		})
	}
}

// Neither a review body nor a nitpick parsed out of one can ever be resolved on
// GitHub — their IsResolved is false by construction. Counting them would pin
// the gate open forever, and a Path != "" qualifier would not catch the nitpick
// because parseNitpickComments synthesizes a Path.
func TestResultFiltersIsCompleteIgnoresCommentsThatCannotBeResolved(t *testing.T) {
	result := sampleCommentGateResult()
	result.Comments = result.Comments[1:] // drop the real thread, keep the unresolvable pair

	filters := newResultFilters([]string{"@coderabbit"}, nil)
	filters.apply(result)

	require.Len(t, result.Comments, 2)
	assert.Equal(t, 0, result.UnresolvedComments())
	assert.True(t, filters.isComplete(result))
}

func TestResultFiltersIsCompleteAndsBothDimensions(t *testing.T) {
	cases := []struct {
		name          string
		checkStatus   string
		threadResolve bool
		want          bool
	}{
		{"checks complete, comment unresolved", "COMPLETED", false, false},
		{"comment resolved, check in flight", "IN_PROGRESS", true, false},
		{"both settled", "COMPLETED", true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := sampleCommentGateResult()
			result.PR.StatusCheckRollup[0].Status = tc.checkStatus
			result.Comments[0].IsResolved = tc.threadResolve

			filters := newResultFilters([]string{"@coderabbit"}, []string{"CI"})
			filters.apply(result)

			assert.Equal(t, tc.want, filters.isComplete(result))
		})
	}
}

// The surprising half of "AND of the ACTIVE dimensions": with only --comments,
// the check rollup is not one of them.
func TestResultFiltersIsCompleteIgnoresChecksWhenOnlyCommentsFiltered(t *testing.T) {
	result := sampleCommentGateResult()
	result.PR.StatusCheckRollup[0].Status = "IN_PROGRESS"
	result.Comments[0].IsResolved = true

	filters := newResultFilters([]string{"@coderabbit"}, nil)
	filters.apply(result)

	assert.True(t, filters.isComplete(result))
}

func TestResultFiltersNoCommentMatch(t *testing.T) {
	t.Run("a selector that pruned real comments to nothing is an error", func(t *testing.T) {
		result := sampleCommentGateResult()
		preComments := len(result.Comments)

		filters := newResultFilters([]string{"@nobody"}, nil)
		filters.apply(result)

		assert.True(t, filters.noCommentMatch(preComments, result))
	})

	t.Run("a PR whose comments have not arrived yet keeps polling", func(t *testing.T) {
		result := sampleCommentGateResult()
		result.Comments = nil

		filters := newResultFilters([]string{"@coderabbit"}, nil)
		filters.apply(result)

		assert.False(t, filters.noCommentMatch(0, result),
			"a review bot posts a minute after a push; an empty-before set is not a typo")
	})
}

// sampleCommentGateResult mirrors what MergeAndFilter produces: a real review
// thread, the review body its nitpicks were parsed out of, and one of those
// nitpicks. Only the first can ever be resolved on GitHub.
func sampleCommentGateResult() *PRWatchResult {
	return &PRWatchResult{
		PR: &github.PRInfo{
			Number: 1,
			StatusCheckRollup: github.StatusChecks{
				{Name: "unit", Status: "COMPLETED", Conclusion: "SUCCESS", WorkflowName: "CI", DetailsURL: "https://github.com/org/repo/actions/runs/101/job/1"},
			},
		},
		Comments: []github.PRComment{
			{ID: 300, Author: "coderabbitai[bot]", BotType: "coderabbit", Path: "foo.go", Line: 42, IsReviewThread: true},
			{ID: 200, Author: "coderabbitai[bot]", BotType: "coderabbit", Body: "**Actionable comments posted: 2**"},
			{ID: 200, Author: "coderabbitai[bot]", BotType: "coderabbit", Path: "bar.go", Line: 7, Severity: "nitpick"},
		},
	}
}

func TestResultFiltersActionsMatchJobName(t *testing.T) {
	t.Run("a job name keeps only that job and its check", func(t *testing.T) {
		result := sampleJobFilterResult()
		filters := newResultFilters(nil, []string{"Install Tests - windows-amd64"})
		filters.apply(result)

		// The Test workflow run is kept, pruned to just the matching job.
		require.Len(t, result.Runs, 1)
		run := result.Runs[502]
		require.NotNil(t, run)
		require.Len(t, run.Jobs, 1)
		assert.Equal(t, "Install Tests - windows-amd64", run.Jobs[0].Name)

		// The sibling "Unit Tests" check of the same run must NOT leak in.
		require.Len(t, result.PR.StatusCheckRollup, 1)
		assert.Equal(t, "Install Tests - windows-amd64", result.PR.StatusCheckRollup[0].Name)
		assert.Equal(t, 1, statusExitCode(result))
	})

	t.Run("a job name whose workflow name differs still matches", func(t *testing.T) {
		result := sampleJobFilterResult()
		filters := newResultFilters(nil, []string{"lint"})
		filters.apply(result)

		require.Len(t, result.PR.StatusCheckRollup, 1)
		assert.Equal(t, "lint", result.PR.StatusCheckRollup[0].Name)
		assert.ElementsMatch(t, []int64{501}, runIDs(result.Runs))
		assert.Equal(t, 1, statusExitCode(result))
	})
}

func TestResultFiltersNoActionMatchDetectsPrunedToEmpty(t *testing.T) {
	result := sampleJobFilterResult()
	preChecks := len(result.PR.StatusCheckRollup)
	preRuns := len(result.Runs)
	options := actionSelectorOptions(result.PR, result.Runs)

	filters := newResultFilters(nil, []string{"no-such-check"})
	filters.apply(result)

	assert.True(t, filters.noActionMatch(preChecks, preRuns, result))
	assert.Contains(t, options, "lint")
	assert.Contains(t, options, "Install Tests - windows-amd64")
	assert.Contains(t, options, "Test")
}

func TestResultFiltersNoActionMatchFalseWhenSomethingMatched(t *testing.T) {
	result := sampleJobFilterResult()
	preChecks := len(result.PR.StatusCheckRollup)
	preRuns := len(result.Runs)

	filters := newResultFilters(nil, []string{"lint"})
	filters.apply(result)

	assert.False(t, filters.noActionMatch(preChecks, preRuns, result))
}

func sampleJobFilterResult() *PRWatchResult {
	return &PRWatchResult{
		PR: &github.PRInfo{
			Number: 1,
			StatusCheckRollup: github.StatusChecks{
				{Name: "lint", Status: "COMPLETED", Conclusion: "FAILURE", WorkflowName: "golangci-lint", DetailsURL: "https://github.com/org/repo/actions/runs/501/job/1"},
				{Name: "Install Tests - windows-amd64", Status: "COMPLETED", Conclusion: "FAILURE", WorkflowName: "Test", DetailsURL: "https://github.com/org/repo/actions/runs/502/job/2"},
				{Name: "Unit Tests", Status: "COMPLETED", Conclusion: "SUCCESS", WorkflowName: "Test", DetailsURL: "https://github.com/org/repo/actions/runs/502/job/3"},
			},
		},
		Runs: map[int64]*github.WorkflowRun{
			501: {DatabaseID: 501, WorkflowID: 51, WorkflowPath: ".github/workflows/lint.yml", Name: "golangci-lint", Jobs: []github.Job{{Name: "lint"}}},
			502: {DatabaseID: 502, WorkflowID: 52, WorkflowPath: ".github/workflows/test.yml", Name: "Test", Jobs: []github.Job{{Name: "Install Tests - windows-amd64"}, {Name: "Unit Tests"}}},
		},
	}
}

func sampleActionFilterResult() *PRWatchResult {
	return &PRWatchResult{
		PR: &github.PRInfo{
			Number: 1,
			StatusCheckRollup: github.StatusChecks{
				{Name: "unit", Status: "COMPLETED", Conclusion: "SUCCESS", WorkflowName: "CI", DetailsURL: "https://github.com/org/repo/actions/runs/101/job/1"},
				{Name: "deploy", Status: "COMPLETED", Conclusion: "FAILURE", WorkflowName: "Deploy", DetailsURL: "https://github.com/org/repo/actions/runs/202/job/2"},
				{Name: "docs", Status: "COMPLETED", Conclusion: "SUCCESS", WorkflowName: "Docs", DetailsURL: "https://github.com/org/repo/actions/runs/303/job/3"},
			},
		},
		Runs: map[int64]*github.WorkflowRun{
			101: {DatabaseID: 101, WorkflowID: 11, WorkflowPath: ".github/workflows/ci.yml", Name: "CI", Status: "completed", Conclusion: "success"},
			202: {DatabaseID: 202, WorkflowID: 22, WorkflowPath: ".github/workflows/deploy.yml", Name: "Deploy", Status: "completed", Conclusion: "failure"},
			303: {DatabaseID: 303, WorkflowID: 33, WorkflowPath: ".github/workflows/docs.yml", Name: "Docs", Status: "completed", Conclusion: "success"},
		},
	}
}

func runIDs(runs map[int64]*github.WorkflowRun) []int64 {
	ids := make([]int64, 0, len(runs))
	for _, run := range runs {
		ids = append(ids, run.DatabaseID)
	}
	return ids
}

func statusCheckRunIDs(t *testing.T, checks github.StatusChecks) []int64 {
	t.Helper()
	ids := make([]int64, 0, len(checks))
	for _, check := range checks {
		id, err := github.ExtractRunID(check.DetailsURL)
		require.NoError(t, err)
		ids = append(ids, id)
	}
	return ids
}
