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
