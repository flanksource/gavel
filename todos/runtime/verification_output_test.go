package runtime

import (
	"encoding/json"
	"testing"

	capapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Verification tab reads a run artifact at
// definitionOfDone.report.tests[].detail.run of a prompt run's result_json. That
// path crosses ExecutionResult → DoDOutcome → api.VerifyReport → VerifyNode, so
// pin it: a JSON tag renamed anywhere along the chain silently blanks the tab.
func TestExecutionResultJSONCarriesRunArtifact(t *testing.T) {
	artifact := fixtures.RunArtifact{
		RunID:  "run-2026-07-30T09-00-00Z-run-the-suite",
		Kind:   "test",
		Total:  7,
		Failed: 2,
		Failures: []fixtures.RunFailure{
			{Name: "TestFoo", Suite: "pkg", Status: "failed", Message: "boom"},
		},
	}
	detail, err := json.Marshal(map[string]any{"run": artifact})
	require.NoError(t, err)

	report := capapi.VerifyReport{
		Kind: "fixture", Ran: true, State: capapi.VerifyStateFailed,
		Tests: []capapi.VerifyNode{{Name: "run the suite", Framework: "test", Failed: true, Detail: detail}},
	}
	report.Summary = capapi.SummarizeNodes(report.Tests)
	require.NoError(t, report.Validate())

	result := &todos.ExecutionResult{
		Success: true,
		DoD:     &todos.DoDOutcome{Ran: true, Passed: false, Report: &report},
	}

	data, err := json.Marshal(executionResultJSON(result))
	require.NoError(t, err)

	var wire struct {
		DoD struct {
			Ran    bool `json:"ran"`
			Passed bool `json:"passed"`
			Report struct {
				Tests []struct {
					Detail struct {
						Run *fixtures.RunArtifact `json:"run"`
					} `json:"detail"`
				} `json:"tests"`
			} `json:"report"`
		} `json:"definitionOfDone"`
	}
	require.NoError(t, json.Unmarshal(data, &wire))

	assert.True(t, wire.DoD.Ran)
	assert.False(t, wire.DoD.Passed)
	require.Len(t, wire.DoD.Report.Tests, 1)

	run := wire.DoD.Report.Tests[0].Detail.Run
	require.NotNil(t, run, "the run artifact must survive the result_json round trip")
	assert.Equal(t, "run-2026-07-30T09-00-00Z-run-the-suite", run.RunID)
	assert.Equal(t, 7, run.Total)
	assert.Equal(t, 2, run.Failed)
	require.Len(t, run.Failures, 1)
	assert.Equal(t, "TestFoo", run.Failures[0].Name)
}
