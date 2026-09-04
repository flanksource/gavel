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

func TestResultFiltersActionNoMatchIsDoneForFollow(t *testing.T) {
	result := sampleActionFilterResult()

	filters := newResultFilters(nil, []string{"missing-action"})
	filters.apply(result)

	assert.Empty(t, result.Runs)
	assert.Empty(t, result.PR.StatusCheckRollup)
	assert.True(t, filters.actionFilteredNoChecks(result))
	assert.Equal(t, 0, statusExitCode(result))
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
