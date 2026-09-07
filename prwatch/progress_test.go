package prwatch

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
)

func TestFollowProgressLineCountsCheckStates(t *testing.T) {
	for _, tc := range []struct {
		name     string
		mutate   func(*PRWatchResult)
		contains []string
		absent   []string
	}{
		{
			name:     "nothing started yet",
			mutate:   func(r *PRWatchResult) { r.PR.StatusCheckRollup[0].Status = "QUEUED" },
			contains: []string{"#1: 0/1 checks complete", "1 running"},
			absent:   []string{"failing"},
		},
		{
			name:     "all green",
			mutate:   func(*PRWatchResult) {},
			contains: []string{"#1: 1/1 checks complete"},
			absent:   []string{"running", "failing"},
		},
		{
			name:     "a definitive failure is counted complete and failing",
			mutate:   func(r *PRWatchResult) { r.PR.StatusCheckRollup[0].Conclusion = "FAILURE" },
			contains: []string{"1/1 checks complete", "1 failing"},
		},
		{
			// IsFailureConclusion excludes NEUTRAL, and so must the heartbeat.
			name:     "neutral is not failing",
			mutate:   func(r *PRWatchResult) { r.PR.StatusCheckRollup[0].Conclusion = "NEUTRAL" },
			contains: []string{"1/1 checks complete"},
			absent:   []string{"failing"},
		},
		{
			name: "an unresolved thread is surfaced",
			mutate: func(r *PRWatchResult) {
				r.Comments = []github.PRComment{{ID: 1, IsReviewThread: true}}
			},
			contains: []string{"1 unresolved comment(s)"},
		},
		{
			// Lint violations alone are Warned, never Failed, so this uses a
			// signal HasFailure actually counts.
			name: "a gavel artifact failure a green rollup hides",
			mutate: func(r *PRWatchResult) {
				r.GavelResults = []*GavelResultsSummary{{StickyID: "gavel", TestsFailed: 2}}
			},
			contains: []string{"1 gavel artifact(s) failing"},
		},
		{
			name: "a failed job under a green rollup",
			mutate: func(r *PRWatchResult) {
				r.Runs = map[int64]*github.WorkflowRun{
					7: {DatabaseID: 7, Jobs: []github.Job{{Name: "e2e", Conclusion: "failure"}}},
				}
			},
			contains: []string{"failed jobs present"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := singleCheckResult()
			tc.mutate(result)

			line := followProgressLine(result, 30*time.Second)

			for _, want := range tc.contains {
				assert.Contains(t, line, want)
			}
			for _, unwanted := range tc.absent {
				assert.NotContains(t, line, unwanted)
			}
			assert.Contains(t, line, "next poll in 30s")
		})
	}
}

// The whole point of the heartbeat is that it replaces a 40-line frame, so a
// reader that cannot redraw gets one line per poll instead of the table again.
func TestFollowProgressLineIsASingleLine(t *testing.T) {
	result := singleCheckResult()
	result.PR.StatusCheckRollup = append(result.PR.StatusCheckRollup,
		github.StatusCheck{Name: "Analyze (go)", Status: "IN_PROGRESS", WorkflowName: "CodeQL"})
	result.Comments = []github.PRComment{{ID: 1, IsReviewThread: true}}
	result.Runs = map[int64]*github.WorkflowRun{
		7: {DatabaseID: 7, Jobs: []github.Job{{Name: "e2e", Conclusion: "failure"}}},
	}

	line := followProgressLine(result, time.Minute)

	assert.NotContains(t, line, "\n")
	assert.False(t, strings.HasSuffix(line, "\n"))
}

func singleCheckResult() *PRWatchResult {
	return &PRWatchResult{
		PR: &github.PRInfo{
			Number: 1,
			StatusCheckRollup: github.StatusChecks{
				{Name: "Test", Status: "COMPLETED", Conclusion: "SUCCESS", WorkflowName: "CI"},
			},
		},
	}
}
