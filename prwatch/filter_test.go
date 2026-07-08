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
