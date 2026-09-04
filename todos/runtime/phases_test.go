package runtime

import (
	"encoding/json"
	"testing"

	capapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// verifyReportJSON is a report as captain stores it in
// captain_prompt_run_iterations.verification_result and surfaces through
// captain_prompt_run_overview.latest_verification_result.
func verifyReportJSON(t *testing.T, passed, failed int) string {
	t.Helper()
	tests := make([]capapi.VerifyNode, 0, passed+failed)
	for i := 0; i < passed; i++ {
		tests = append(tests, capapi.VerifyNode{Name: "passing", Framework: "test", Passed: true})
	}
	for i := 0; i < failed; i++ {
		tests = append(tests, capapi.VerifyNode{Name: "failing", Framework: "test", Failed: true})
	}
	report := capapi.VerifyReport{Kind: "fixture", Ran: true, Tests: tests}
	report.Summary = capapi.SummarizeNodes(report.Tests)
	report.State = capapi.VerifyStatePassed
	if failed > 0 {
		report.State = capapi.VerifyStateFailed
	}
	report.Passed = failed == 0
	require.NoError(t, report.Validate())
	data, err := json.Marshal(report)
	require.NoError(t, err)
	return string(data)
}

// The verify phase counts the checks in the definition of done, and the only
// place those counts live is captain's own verification_result. Reading them
// back out of the run's result_json — a gavel-shaped copy — is what let the two
// disagree whenever the copy was not written.
func TestVerifyPhaseProgressComesFromCaptainVerificationResult(t *testing.T) {
	run, err := phaseRunFromNative(native.IssuePhaseRun{
		Phase: native.StepVerify, State: "failed",
		VerificationResult: verifyReportJSON(t, 3, 2),
		// The iteration counters describe agent turns, not checks: a verify phase
		// that reported them would show "1/1" for a five-check fixture.
		Iterations: 1, Succeeded: 0, Failed: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, types.PhaseProgress{Done: 3, Failed: 2, Total: 5}, run.Progress)
}

func TestVerifyPhaseWithNoCaptainReportShowsNoProgress(t *testing.T) {
	run, err := phaseRunFromNative(native.IssuePhaseRun{
		Phase: native.StepVerify, State: "running", Iterations: 1,
	})

	require.NoError(t, err)
	assert.True(t, run.Progress.Empty(), "a verify phase that produced no report counts nothing")
}

func TestVerifyPhaseFailsLoudlyOnACorruptCaptainReport(t *testing.T) {
	_, err := verificationProgress(`{"summary":`)

	require.Error(t, err, "a corrupt verification result must not read as an empty pass")
}

// Every other phase counts agent iterations, and must keep doing so.
func TestAgentPhaseProgressCountsIterations(t *testing.T) {
	run, err := phaseRunFromNative(native.IssuePhaseRun{
		Phase: native.StepRun, State: "succeeded", Iterations: 4, Succeeded: 3, Failed: 1,
		VerificationResult: verifyReportJSON(t, 9, 0),
	})

	require.NoError(t, err)
	assert.Equal(t, types.PhaseProgress{Done: 3, Failed: 1, Total: 4}, run.Progress)
}
