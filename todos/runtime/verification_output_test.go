package runtime

import (
	"encoding/json"
	"testing"

	capapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Captain stores this report on an iteration row. The run artifact must survive
// the VerifyReport JSON shape consumed by the verification UI.
func TestVerificationReportCarriesRunArtifact(t *testing.T) {
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

	data, err := json.Marshal(report)
	require.NoError(t, err)

	var wire struct {
		Ran    bool `json:"ran"`
		Passed bool `json:"passed"`
		Tests  []struct {
			Detail struct {
				Run *fixtures.RunArtifact `json:"run"`
			} `json:"detail"`
		} `json:"tests"`
	}
	require.NoError(t, json.Unmarshal(data, &wire))

	assert.True(t, wire.Ran)
	assert.False(t, wire.Passed)
	require.Len(t, wire.Tests, 1)

	run := wire.Tests[0].Detail.Run
	require.NotNil(t, run, "the run artifact must survive the verification report round trip")
	assert.Equal(t, "run-2026-07-30T09-00-00Z-run-the-suite", run.RunID)
	assert.Equal(t, 7, run.Total)
	assert.Equal(t, 2, run.Failed)
	require.Len(t, run.Failures, 1)
	assert.Equal(t, "TestFoo", run.Failures[0].Name)
}
