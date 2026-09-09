package prwatch

import (
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/linters"
	"github.com/stretchr/testify/assert"
)

// greenRollupResult mirrors the shape that made `gavel pr status
// flanksource/config-db#2347` exit 0: every rollup context passes, and the only
// failure signal lives in a gavel artifact harvested from a PR comment.
func greenRollupResult() *PRWatchResult {
	return &PRWatchResult{
		PR: &github.PRInfo{
			Number: 2347,
			StatusCheckRollup: github.StatusChecks{
				{Name: "license/cla", Status: "COMPLETED", Conclusion: "SUCCESS"},
				{Name: "Socket Security: Pull Request Alerts", Status: "COMPLETED", Conclusion: "NEUTRAL"},
			},
		},
	}
}

func TestPRWatchResultHasTerminalFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*PRWatchResult)
		want   bool
	}{
		{"green rollup", func(*PRWatchResult) {}, false},
		{"failure", func(r *PRWatchResult) {
			r.PR.StatusCheckRollup[0].Conclusion = "FAILURE"
		}, true},
		{"timed out", func(r *PRWatchResult) {
			r.PR.StatusCheckRollup[0].Conclusion = "TIMED_OUT"
		}, true},
		{"startup failure", func(r *PRWatchResult) {
			r.PR.StatusCheckRollup[0].Conclusion = "STARTUP_FAILURE"
		}, true},
		{"cancelled is not definitive", func(r *PRWatchResult) {
			r.PR.StatusCheckRollup[0].Conclusion = "CANCELLED"
		}, false},
		{"still running has no conclusion", func(r *PRWatchResult) {
			r.PR.StatusCheckRollup[0].Status, r.PR.StatusCheckRollup[0].Conclusion = "IN_PROGRESS", ""
		}, false},
		{"failed job under a green rollup", func(r *PRWatchResult) {
			r.Runs = map[int64]*github.WorkflowRun{
				7: {DatabaseID: 7, Jobs: []github.Job{{Name: "e2e", Conclusion: "failure"}}},
			}
		}, true},
		// A download that failed this poll is retried on the next one. Aborting
		// the watch on it would report a red PR that is merely unreadable — so
		// it deliberately diverges from statusExitCode, which still returns 1.
		{"unreadable gavel artifact is not terminal", func(r *PRWatchResult) {
			r.GavelResults = []*GavelResultsSummary{{StickyID: "gavel", Error: "download failed"}}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := greenRollupResult()
			tc.mutate(result)
			assert.Equal(t, tc.want, result.HasTerminalFailure())
		})
	}

	t.Run("nil PR does not panic", func(t *testing.T) {
		assert.False(t, PRWatchResult{}.HasTerminalFailure())
	})

	t.Run("an unreadable artifact still fails the exit code", func(t *testing.T) {
		result := greenRollupResult()
		result.GavelResults = []*GavelResultsSummary{{StickyID: "gavel", Error: "download failed"}}
		assert.False(t, result.HasTerminalFailure())
		assert.Equal(t, 1, statusExitCode(result), "the divergence is deliberate, not an oversight")
	})
}

func TestFollowDone(t *testing.T) {
	// A rollup with one definitive failure alongside one check still running:
	// the shape --fail-fast exists for.
	mixed := func() *PRWatchResult {
		return &PRWatchResult{
			PR: &github.PRInfo{
				Number: 187,
				StatusCheckRollup: github.StatusChecks{
					{Name: "Test", Status: "COMPLETED", Conclusion: "FAILURE", WorkflowName: "CI",
						DetailsURL: "https://github.com/org/repo/actions/runs/101/job/1"},
					{Name: "Analyze (go)", Status: "IN_PROGRESS", WorkflowName: "CodeQL"},
				},
			},
		}
	}

	t.Run("without fail-fast a red check does not release the gate", func(t *testing.T) {
		filters := newResultFilters(nil, nil)
		assert.False(t, followDone(filters, mixed(), false))
	})

	t.Run("with fail-fast it does", func(t *testing.T) {
		filters := newResultFilters(nil, nil)
		assert.True(t, followDone(filters, mixed(), true))
	})

	t.Run("fail-fast with nothing red still waits", func(t *testing.T) {
		result := mixed()
		result.PR.StatusCheckRollup[0].Conclusion = "SUCCESS"
		filters := newResultFilters(nil, nil)
		assert.False(t, followDone(filters, result, true))
	})

	t.Run("fail-fast is scoped by --actions", func(t *testing.T) {
		result := mixed()
		filters := newResultFilters(nil, []string{"CodeQL"})
		filters.apply(result)
		assert.False(t, followDone(filters, result, true),
			"the only red check was filtered out, so there is nothing to fail fast on")
	})

	t.Run("fail-fast catches a failed job under a green rollup", func(t *testing.T) {
		result := greenRollupResult()
		result.Runs = map[int64]*github.WorkflowRun{
			7: {DatabaseID: 7, Jobs: []github.Job{{Name: "e2e", Conclusion: "failure"}}},
		}
		filters := newResultFilters(nil, nil)
		assert.True(t, followDone(filters, result, true))
	})
}

func TestStatusExitCode(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*PRWatchResult)
		wantExt int
	}{
		{
			name:    "all green",
			mutate:  func(*PRWatchResult) {},
			wantExt: 0,
		},
		{
			name: "failing rollup context",
			mutate: func(r *PRWatchResult) {
				r.PR.StatusCheckRollup = append(r.PR.StatusCheckRollup,
					github.StatusChecks{{Name: "build", Status: "COMPLETED", Conclusion: "FAILURE"}}...)
			},
			wantExt: 1,
		},
		{
			name: "gavel artifact test failure",
			mutate: func(r *PRWatchResult) {
				r.GavelResults = []*GavelResultsSummary{
					{StickyID: "gavel-core", TestsPassed: 240},
					{StickyID: "gavel-scrapers", TestsPassed: 855, TestsFailed: 1},
				}
			},
			wantExt: 1,
		},
		{
			name: "gavel artifact could not be read",
			mutate: func(r *PRWatchResult) {
				r.GavelResults = []*GavelResultsSummary{{StickyID: "gavel", Error: "download failed"}}
			},
			wantExt: 1,
		},
		{
			name: "gavel artifact linter died",
			mutate: func(r *PRWatchResult) {
				r.GavelResults = []*GavelResultsSummary{{
					StickyID: "gavel-lint",
					Lint:     []*linters.LinterResult{{Success: false, Error: "golangci-lint: exec format error"}},
				}}
			},
			wantExt: 1,
		},
		{
			name: "gavel artifact bench regression",
			mutate: func(r *PRWatchResult) {
				r.GavelResults = []*GavelResultsSummary{{StickyID: "gavel-bench", HasBench: true, BenchRegressions: 1}}
			},
			wantExt: 1,
		},
		{
			name: "lint violations alone stay a warning",
			mutate: func(r *PRWatchResult) {
				r.GavelResults = []*GavelResultsSummary{{StickyID: "gavel-lint", LintViolations: 3, LintLinters: 1}}
			},
			wantExt: 0,
		},
		{
			name: "failed job under a passing rollup context",
			mutate: func(r *PRWatchResult) {
				r.Runs = map[int64]*github.WorkflowRun{
					101: {DatabaseID: 101, Name: "CI", Status: "completed", Conclusion: "success", Jobs: []github.Job{
						{Name: "unit", Status: "completed", Conclusion: "success"},
						{Name: "e2e", Status: "completed", Conclusion: "failure"},
					}},
				}
			},
			wantExt: 1,
		},
		{
			name: "nil run is not a failure",
			mutate: func(r *PRWatchResult) {
				r.Runs = map[int64]*github.WorkflowRun{101: nil}
			},
			wantExt: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := greenRollupResult()
			tc.mutate(result)

			assert.Equal(t, tc.wantExt, statusExitCode(result))
		})
	}
}

func TestStatusExitCodeNilResult(t *testing.T) {
	assert.Equal(t, 0, statusExitCode(nil))
}
