package github

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFailureConclusion(t *testing.T) {
	cases := []struct {
		conclusion string
		want       bool
	}{
		{"failure", true},
		{"FAILURE", true},
		{"timed_out", true},
		{"TIMED_OUT", true},
		{"startup_failure", true},
		{"STARTUP_FAILURE", true},
		{"cancelled", false},
		{"CANCELLED", false},
		{"success", false},
		{"skipped", false},
		{"neutral", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.conclusion, func(t *testing.T) {
			assert.Equal(t, tc.want, IsFailureConclusion(tc.conclusion))
		})
	}
}

// An in-progress run (empty run conclusion) whose jobs have already failed
// must report a failed job — the previous run-conclusion-only check missed
// this, which is why --logs showed nothing while a slow e2e job kept the run
// in_progress.
func TestRunHasFailedJob(t *testing.T) {
	inProgressWithFailure := &WorkflowRun{
		Status:     "in_progress",
		Conclusion: "",
		Jobs: []Job{
			{Name: "check", Conclusion: "failure"},
			{Name: "e2e", Conclusion: ""},
		},
	}
	assert.True(t, RunHasFailedJob(inProgressWithFailure))

	allPassing := &WorkflowRun{
		Status:     "completed",
		Conclusion: "success",
		Jobs:       []Job{{Name: "build", Conclusion: "success"}},
	}
	assert.False(t, RunHasFailedJob(allPassing))

	cancelledNotFailure := &WorkflowRun{
		Status:     "completed",
		Conclusion: "cancelled",
		Jobs:       []Job{{Name: "build", Conclusion: "cancelled"}},
	}
	assert.False(t, RunHasFailedJob(cancelledNotFailure))
}

// A timed_out job with attached logs must render its step logs — the previous
// failure-only guard returned before any step was emitted.
func TestJobPrettyRendersLogsForTimedOutJob(t *testing.T) {
	job := Job{
		Name:       "build",
		Status:     "completed",
		Conclusion: "timed_out",
		Steps: []Step{
			{
				Name:       "go test",
				Status:     "completed",
				Conclusion: "failure",
				Logs:       "FAIL\tgithub.com/foo/bar\t0.01s",
			},
		},
	}
	out := job.Pretty().String()
	assert.Contains(t, out, "go test")
	assert.Contains(t, out, "FAIL\tgithub.com/foo/bar")
}

// A successful job must NOT render step logs, even if a step carries some.
func TestJobPrettyHidesLogsForSuccessfulJob(t *testing.T) {
	job := Job{
		Name:       "build",
		Status:     "completed",
		Conclusion: "success",
		Steps: []Step{
			{Name: "go test", Status: "completed", Conclusion: "success", Logs: "ok"},
		},
	}
	out := job.Pretty().String()
	assert.NotContains(t, out, "ok")
	assert.False(t, strings.Contains(out, "go test"), "step names should be hidden for passing jobs")
}

// When a failed job has no per-step logs but a job-level log tail, Pretty must
// fall back to rendering that tail.
func TestJobPrettyFallsBackToJobLogTail(t *testing.T) {
	job := Job{
		Name:       "build",
		Status:     "completed",
		Conclusion: "failure",
		Logs:       "panic: runtime error: index out of range",
	}
	out := job.Pretty().String()
	assert.Contains(t, out, "Log tail")
	assert.Contains(t, out, "panic: runtime error")
}
