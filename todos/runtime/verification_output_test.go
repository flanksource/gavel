package runtime

import (
	"encoding/json"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos"
	todotypes "github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Verification tab reads a run artifact at
// definitionOfDone.output.results[].run of a prompt run's result_json. That path
// crosses ExecutionResult → DoDOutcome → VerificationOutput → FixtureResult, so
// pin it: a JSON tag renamed anywhere along the chain silently blanks the tab.
func TestExecutionResultJSONCarriesRunArtifact(t *testing.T) {
	result := &todos.ExecutionResult{
		Success: true,
		DoD: &todos.DoDOutcome{
			Ran:    true,
			Passed: false,
			Output: &todotypes.VerificationOutput{
				Results: []fixtures.FixtureResult{{
					Name: "run the suite",
					Type: "test",
					Run: &fixtures.RunArtifact{
						RunID:  "run-2026-07-30T09-00-00Z-run-the-suite",
						Kind:   "test",
						Total:  7,
						Failed: 2,
						Failures: []fixtures.RunFailure{
							{Name: "TestFoo", Suite: "pkg", Status: "failed", Message: "boom"},
						},
					},
				}},
			},
		},
	}

	data, err := json.Marshal(executionResultJSON(result))
	require.NoError(t, err)

	var wire struct {
		DoD struct {
			Ran    bool `json:"ran"`
			Passed bool `json:"passed"`
			Output struct {
				Results []struct {
					Run *fixtures.RunArtifact `json:"run"`
				} `json:"results"`
			} `json:"output"`
		} `json:"definitionOfDone"`
	}
	require.NoError(t, json.Unmarshal(data, &wire))

	assert.True(t, wire.DoD.Ran)
	assert.False(t, wire.DoD.Passed)
	require.Len(t, wire.DoD.Output.Results, 1)

	run := wire.DoD.Output.Results[0].Run
	require.NotNil(t, run, "the run artifact must survive the result_json round trip")
	assert.Equal(t, "run-2026-07-30T09-00-00Z-run-the-suite", run.RunID)
	assert.Equal(t, 7, run.Total)
	assert.Equal(t, 2, run.Failed)
	require.Len(t, run.Failures, 1)
	assert.Equal(t, "TestFoo", run.Failures[0].Name)
}
