package prwatch

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
)

func TestPRWatchResultPretty(t *testing.T) {
	result := PRWatchResult{
		PR: &github.PRInfo{
			Number:      99,
			Title:       "feat: new feature",
			Author:      github.PRAuthor{Login: "alice"},
			HeadRefName: "feat/new",
			BaseRefName: "main",
			State:       "OPEN",
			StatusCheckRollup: github.StatusChecks{
				{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS", DetailsURL: "https://github.com/org/repo/actions/runs/789/job/1"},
			},
		},
		Runs: map[int64]*github.WorkflowRun{
			789: {
				DatabaseID: 789, Name: "CI", Status: "completed", Conclusion: "success",
				Jobs: []github.Job{{
					Name: "lint", Status: "completed", Conclusion: "success",
					StartedAt:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					CompletedAt: time.Date(2024, 1, 1, 10, 0, 30, 0, time.UTC),
				}},
			},
		},
	}

	text := result.Pretty().String()
	assert.Contains(t, text, "PR #99")
	assert.Contains(t, text, "feat: new feature")
	assert.Contains(t, text, "alice")
	assert.Contains(t, text, "CI")
	assert.Contains(t, text, "lint")
	assert.Contains(t, text, "Workflows:")
}

func TestPRWatchResultPrettyWithStepLogs(t *testing.T) {
	result := PRWatchResult{
		PR: &github.PRInfo{
			Number:      50,
			Title:       "fix: bug",
			Author:      github.PRAuthor{Login: "bob"},
			HeadRefName: "fix/bug",
			BaseRefName: "main",
			StatusCheckRollup: github.StatusChecks{
				{Name: "test", Status: "COMPLETED", Conclusion: "FAILURE", DetailsURL: "https://github.com/org/repo/actions/runs/100/job/1"},
			},
		},
		Runs: map[int64]*github.WorkflowRun{
			100: {
				DatabaseID: 100, Name: "Tests", Status: "completed", Conclusion: "failure",
				Jobs: []github.Job{{
					Name: "test", Status: "completed", Conclusion: "failure",
					StartedAt:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					CompletedAt: time.Date(2024, 1, 1, 10, 0, 45, 0, time.UTC),
					Steps: []github.Step{
						{Name: "Run tests", Status: "completed", Conclusion: "failure", Number: 1, Logs: "FAIL: TestFoo\nexit status 1"},
					},
				}},
			},
		},
	}

	text := result.Pretty().String()
	assert.Contains(t, text, "fix: bug")
	assert.Contains(t, text, "Run tests")
	assert.Contains(t, text, "FAIL: TestFoo")
	assert.Contains(t, text, "exit status 1")
}

func TestPRWatchResultNoChecks(t *testing.T) {
	result := PRWatchResult{
		PR: &github.PRInfo{
			Number: 1, Title: "empty", Author: github.PRAuthor{Login: "dev"},
			HeadRefName: "branch", BaseRefName: "main",
		},
	}
	text := result.Pretty().String()
	assert.Contains(t, text, "No checks found")
}
