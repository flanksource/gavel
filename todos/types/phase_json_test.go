package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPhaseRunJSONContract pins the wire shape the dashboard's TodoPhaseRun
// interface reads. The two are hand-written on opposite sides of the HTTP
// boundary, so a renamed field here would silently render every phase cell as
// "never run" rather than failing anywhere.
func TestPhaseRunJSONContract(t *testing.T) {
	started := time.Date(2026, 8, 26, 11, 58, 30, 0, time.UTC)
	finished := started.Add(90 * time.Second)

	encoded, err := json.Marshal(PhaseRuns{
		VerifyPhase: {
			Phase:      VerifyPhase,
			State:      "failed",
			Progress:   PhaseProgress{Done: 3, Failed: 1, Total: 4},
			StartedAt:  &started,
			FinishedAt: &finished,
			DurationMS: 90_000,
			CostUSD:    1.25,
			Active:     true,
		},
	})
	require.NoError(t, err)

	var decoded map[string]map[string]any
	require.NoError(t, json.Unmarshal(encoded, &decoded))

	verify := decoded["verify"]
	require.NotNil(t, verify, "phases are keyed by phase name")
	assert.Equal(t, "failed", verify["state"])
	assert.Equal(t, float64(90_000), verify["duration_ms"])
	assert.Equal(t, 1.25, verify["cost_usd"])
	assert.Equal(t, true, verify["active"])
	assert.Equal(t, "2026-08-26T11:58:30Z", verify["started_at"])
	assert.Equal(t, map[string]any{"done": float64(3), "failed": float64(1), "total": float64(4)}, verify["progress"])
}

// A phase that has never run must be ABSENT rather than zero-valued: that
// absence is what the dashboard renders as an em-dash, and a zero value would
// claim the phase ran and produced nothing.
func TestPhaseRunsOmitsPhasesThatNeverRan(t *testing.T) {
	encoded, err := json.Marshal(TODO{PhaseRuns: PhaseRuns{RunPhase: {State: "succeeded"}}})
	require.NoError(t, err)

	var decoded struct {
		PhaseRuns map[string]json.RawMessage `json:"phase_runs"`
	}
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Len(t, decoded.PhaseRuns, 1)
	assert.NotContains(t, decoded.PhaseRuns, "triage")

	// And a TODO with no runs at all omits the field entirely.
	bare, err := json.Marshal(TODO{})
	require.NoError(t, err)
	assert.NotContains(t, string(bare), "phase_runs")
}
