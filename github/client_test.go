package github

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const prFixture = `{
  "number": 42,
  "title": "feat: add new widget",
  "author": {"login": "moshloop", "name": "Moshe"},
  "headRefName": "feat/widget",
  "baseRefName": "main",
  "state": "OPEN",
  "isDraft": false,
  "reviewDecision": "APPROVED",
  "mergeable": "MERGEABLE",
  "url": "https://github.com/org/repo/pull/42",
  "statusCheckRollup": [
    {
      "name": "build",
      "status": "COMPLETED",
      "conclusion": "SUCCESS",
      "workflowName": "CI",
      "detailsUrl": "https://github.com/org/repo/actions/runs/12345/job/67890"
    },
    {
      "name": "lint",
      "status": "IN_PROGRESS",
      "conclusion": "",
      "workflowName": "CI",
      "detailsUrl": "https://github.com/org/repo/actions/runs/12345/job/67891"
    },
    {
      "name": "test",
      "status": "COMPLETED",
      "conclusion": "FAILURE",
      "workflowName": "Test",
      "detailsUrl": "https://github.com/org/repo/actions/runs/12346/job/67892"
    }
  ]
}`

const runFixture = `{
  "databaseId": 12345,
  "name": "CI",
  "status": "completed",
  "conclusion": "failure",
  "url": "https://github.com/org/repo/actions/runs/12345",
  "jobs": [
    {
      "name": "build",
      "status": "completed",
      "conclusion": "success",
      "startedAt": "2024-01-15T10:00:00Z",
      "completedAt": "2024-01-15T10:01:23Z",
      "steps": [
        {"name": "Set up job", "status": "completed", "conclusion": "success", "number": 1},
        {"name": "Checkout", "status": "completed", "conclusion": "success", "number": 2},
        {"name": "Build", "status": "completed", "conclusion": "success", "number": 3}
      ]
    },
    {
      "name": "test",
      "status": "completed",
      "conclusion": "failure",
      "startedAt": "2024-01-15T10:00:00Z",
      "completedAt": "2024-01-15T10:02:45Z",
      "steps": [
        {"name": "Set up job", "status": "completed", "conclusion": "success", "number": 1},
        {"name": "Run tests", "status": "completed", "conclusion": "failure", "number": 2}
      ]
    }
  ]
}`

func TestParsePRJSON(t *testing.T) {
	pr, err := ParsePRJSON([]byte(prFixture))
	require.NoError(t, err)

	assert.Equal(t, 42, pr.Number)
	assert.Equal(t, "feat: add new widget", pr.Title)
	assert.Equal(t, "moshloop", pr.Author.Login)
	assert.Equal(t, "feat/widget", pr.HeadRefName)
	assert.Equal(t, "main", pr.BaseRefName)
	assert.Equal(t, "OPEN", pr.State)
	assert.False(t, pr.IsDraft)
	assert.Equal(t, "APPROVED", pr.ReviewDecision)
	assert.Equal(t, "MERGEABLE", pr.Mergeable)
	assert.Len(t, pr.StatusCheckRollup, 3)
	assert.Equal(t, "build", pr.StatusCheckRollup[0].Name)
	assert.Equal(t, "COMPLETED", pr.StatusCheckRollup[0].Status)
	assert.Equal(t, "SUCCESS", pr.StatusCheckRollup[0].Conclusion)
}

func TestParseRunJSON(t *testing.T) {
	run, err := ParseRunJSON([]byte(runFixture))
	require.NoError(t, err)

	assert.Equal(t, int64(12345), run.DatabaseID)
	assert.Equal(t, "CI", run.Name)
	assert.Len(t, run.Jobs, 2)
	assert.Equal(t, "build", run.Jobs[0].Name)
	assert.Len(t, run.Jobs[0].Steps, 3)
	assert.Equal(t, "test", run.Jobs[1].Name)
	assert.Equal(t, "failure", run.Jobs[1].Conclusion)
	assert.Len(t, run.Jobs[1].Steps, 2)
}

func TestExtractRunID(t *testing.T) {
	tests := []struct {
		url    string
		expect int64
		hasErr bool
	}{
		{"https://github.com/org/repo/actions/runs/12345/job/67890", 12345, false},
		{"https://github.com/org/repo/actions/runs/999", 999, false},
		{"https://example.com/no-run-id", 0, true},
		{"", 0, true},
	}
	for _, tc := range tests {
		id, err := ExtractRunID(tc.url)
		if tc.hasErr {
			assert.Error(t, err, "url=%s", tc.url)
		} else {
			require.NoError(t, err, "url=%s", tc.url)
			assert.Equal(t, tc.expect, id)
		}
	}
}

func TestAllComplete(t *testing.T) {
	tests := []struct {
		name   string
		checks StatusChecks
		expect bool
	}{
		{"all completed", StatusChecks{
			{Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Status: "COMPLETED", Conclusion: "FAILURE"},
		}, true},
		{"one in progress", StatusChecks{
			{Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Status: "IN_PROGRESS"},
		}, false},
		{"empty", StatusChecks{}, false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expect, tc.checks.AllComplete(), tc.name)
	}
}

func TestHasFailure(t *testing.T) {
	tests := []struct {
		name   string
		checks StatusChecks
		expect bool
	}{
		{"has failure", StatusChecks{
			{Conclusion: "SUCCESS"},
			{Conclusion: "FAILURE"},
		}, true},
		{"has timed out", StatusChecks{
			{Conclusion: "TIMED_OUT"},
		}, true},
		{"all success", StatusChecks{
			{Conclusion: "SUCCESS"},
			{Conclusion: "NEUTRAL"},
		}, false},
		{"empty", StatusChecks{}, false},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expect, tc.checks.HasFailure(), tc.name)
	}
}

func TestStatusIcon(t *testing.T) {
	tests := []struct {
		status, conclusion string
		expectContains     string
	}{
		{"COMPLETED", "SUCCESS", "✓"},
		{"COMPLETED", "FAILURE", "✗"},
		{"COMPLETED", "CANCELLED", "■"},
		{"COMPLETED", "SKIPPED", "→"},
		{"IN_PROGRESS", "", "●"},
		{"QUEUED", "", "○"},
		{"PENDING", "", "○"},
	}
	for _, tc := range tests {
		icon := StatusIcon(tc.status, tc.conclusion)
		assert.Contains(t, icon.String(), tc.expectContains,
			"status=%s conclusion=%s", tc.status, tc.conclusion)
	}
}

func TestParseGitHubRepo(t *testing.T) {
	tests := []struct {
		url    string
		expect string
		hasErr bool
	}{
		{"https://github.com/flanksource/mission-control.git", "flanksource/mission-control", false},
		{"https://github.com/flanksource/mission-control", "flanksource/mission-control", false},
		{"git@github.com:flanksource/mission-control.git", "flanksource/mission-control", false},
		{"git@github.com:flanksource/mission-control", "flanksource/mission-control", false},
		{"ssh://git@github.com/flanksource/mission-control.git", "flanksource/mission-control", false},
		{"ssh://git@github.com/flanksource/mission-control", "flanksource/mission-control", false},
		{"https://gitlab.com/some/repo", "", true},
		{"not-a-url", "", true},
	}
	for _, tc := range tests {
		result, err := parseGitHubRepo(tc.url)
		if tc.hasErr {
			assert.Error(t, err, "url=%s", tc.url)
		} else {
			require.NoError(t, err, "url=%s", tc.url)
			assert.Equal(t, tc.expect, result, "url=%s", tc.url)
		}
	}
}

func TestTokenResolution(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		envGH     string
		envGitHub string
		expect    string
		hasErr    bool
	}{
		{"field takes priority", "field-token", "env-gh", "env-github", "field-token", false},
		{"GITHUB_TOKEN fallback", "", "", "env-github", "env-github", false},
		{"GH_TOKEN fallback", "", "env-gh", "", "env-gh", false},
		{"GITHUB_TOKEN before GH_TOKEN", "", "env-gh", "env-github", "env-github", false},
		{"no token", "", "", "", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GITHUB_TOKEN", tc.envGitHub)
			t.Setenv("GH_TOKEN", tc.envGH)
			opts := Options{Token: tc.token}
			result, err := opts.token()
			if tc.hasErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, result)
			}
		})
	}
}

func TestMapGraphQLPRToPRInfo(t *testing.T) {
	conclusion := "SUCCESS"
	gqlPR := graphQLPR{
		Number:         42,
		Title:          "feat: widgets",
		Author:         graphQLAuthor{Login: "moshloop"},
		HeadRefName:    "feat/widgets",
		BaseRefName:    "main",
		State:          "OPEN",
		IsDraft:        false,
		ReviewDecision: "APPROVED",
		Mergeable:      "MERGEABLE",
		URL:            "https://github.com/org/repo/pull/42",
		Commits: graphQLCommits{
			Nodes: []graphQLCommitNode{{
				Commit: graphQLCommit{
					StatusCheckRollup: &graphQLStatusCheckRollup{
						Contexts: graphQLContexts{
							Nodes: []graphQLCheckNode{
								{
									Typename:   "CheckRun",
									Name:       "build",
									Status:     "COMPLETED",
									Conclusion: &conclusion,
									DetailsURL: "https://github.com/org/repo/actions/runs/123/job/456",
									CheckSuite: &graphQLCheckSuite{
										WorkflowRun: &graphQLWorkflowRun{
											Workflow: graphQLWorkflow{Name: "CI"},
										},
									},
								},
								{
									Typename:  "StatusContext",
									Context:   "ci/external",
									State:     "success",
									TargetURL: "https://external.ci/build/789",
								},
							},
						},
					},
				},
			}},
		},
	}

	pr := gqlPR.toPRInfo()

	assert.Equal(t, 42, pr.Number)
	assert.Equal(t, "feat: widgets", pr.Title)
	assert.Equal(t, "moshloop", pr.Author.Login)
	assert.Equal(t, "feat/widgets", pr.HeadRefName)
	assert.Equal(t, "main", pr.BaseRefName)
	assert.Equal(t, "OPEN", pr.State)
	assert.Equal(t, "APPROVED", pr.ReviewDecision)
	assert.Equal(t, "MERGEABLE", pr.Mergeable)
	require.Len(t, pr.StatusCheckRollup, 2)

	check := pr.StatusCheckRollup[0]
	assert.Equal(t, "build", check.Name)
	assert.Equal(t, "COMPLETED", check.Status)
	assert.Equal(t, "SUCCESS", check.Conclusion)
	assert.Equal(t, "CI", check.WorkflowName)

	status := pr.StatusCheckRollup[1]
	assert.Equal(t, "ci/external", status.Name)
	assert.Equal(t, "COMPLETED", status.Status)
	assert.Equal(t, "SUCCESS", status.Conclusion)
	assert.Equal(t, "https://external.ci/build/789", status.DetailsURL)
}

func TestMapCheckRunToStatusCheck(t *testing.T) {
	conclusion := "FAILURE"
	node := graphQLCheckNode{
		Typename:   "CheckRun",
		Name:       "test",
		Status:     "completed",
		Conclusion: &conclusion,
		DetailsURL: "https://github.com/org/repo/actions/runs/100/job/200",
		CheckSuite: &graphQLCheckSuite{
			WorkflowRun: &graphQLWorkflowRun{
				Workflow: graphQLWorkflow{Name: "Tests"},
			},
		},
	}

	sc := node.checkRunToStatusCheck()
	assert.Equal(t, "test", sc.Name)
	assert.Equal(t, "COMPLETED", sc.Status)
	assert.Equal(t, "FAILURE", sc.Conclusion)
	assert.Equal(t, "Tests", sc.WorkflowName)
}

func TestMapStatusContextToStatusCheck(t *testing.T) {
	tests := []struct {
		state            string
		expectStatus     string
		expectConclusion string
	}{
		{"SUCCESS", "COMPLETED", "SUCCESS"},
		{"FAILURE", "COMPLETED", "FAILURE"},
		{"ERROR", "COMPLETED", "FAILURE"},
		{"PENDING", "PENDING", ""},
		{"EXPECTED", "QUEUED", ""},
	}
	for _, tc := range tests {
		node := graphQLCheckNode{
			Typename:  "StatusContext",
			Context:   "ci/test",
			State:     tc.state,
			TargetURL: "https://example.com",
		}
		sc := node.statusContextToStatusCheck()
		assert.Equal(t, tc.expectStatus, sc.Status, "state=%s", tc.state)
		assert.Equal(t, tc.expectConclusion, sc.Conclusion, "state=%s", tc.state)
		assert.Equal(t, "ci/test", sc.Name)
	}
}

func TestMapRestJobToJob(t *testing.T) {
	started := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	completed := time.Date(2024, 1, 15, 10, 2, 45, 0, time.UTC)
	rj := restJob{
		ID:          67890,
		Name:        "test",
		Status:      "completed",
		Conclusion:  "failure",
		StartedAt:   started,
		CompletedAt: completed,
		HTMLURL:     "https://github.com/org/repo/actions/runs/12345/jobs/67890",
		Steps: []restStep{
			{Name: "Set up job", Status: "completed", Conclusion: "success", Number: 1},
			{Name: "Run tests", Status: "completed", Conclusion: "failure", Number: 2},
		},
	}

	job := rj.toJob()
	assert.Equal(t, int64(67890), job.DatabaseID)
	assert.Equal(t, "test", job.Name)
	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, "failure", job.Conclusion)
	assert.Equal(t, started, job.StartedAt)
	assert.Equal(t, completed, job.CompletedAt)
	require.Len(t, job.Steps, 2)
	assert.Equal(t, "Run tests", job.Steps[1].Name)
	assert.Equal(t, "failure", job.Steps[1].Conclusion)
	assert.Equal(t, 2, job.Steps[1].Number)
}
