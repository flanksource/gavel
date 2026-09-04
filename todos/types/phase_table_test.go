package types

import (
	"testing"
	"time"

	"github.com/flanksource/clicky"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTodoTableRendersPhaseColumns renders the table `gavel todos list` prints,
// through clicky rather than by inspecting PrettyRow's map.
//
// This is the assertion the map-level tests cannot make: clicky derives its
// headers from the FIRST row alone, so a phase column that is only emitted when
// a phase ran would silently vanish for the whole table whenever the first todo
// happened never to have run. The fixture puts the never-run todo first on
// purpose.
func TestTodoTableRendersPhaseColumns(t *testing.T) {
	started := time.Now().Add(-134 * time.Second)
	todos := []TODO{
		{TODOFrontmatter: TODOFrontmatter{Title: "Never run", Priority: PriorityLow}},
		{
			TODOFrontmatter: TODOFrontmatter{Title: "Ran everything", Priority: PriorityHigh},
			PhaseRuns: PhaseRuns{
				PlanPhase:   {State: "succeeded", StartedAt: &started, DurationMS: 134_000},
				TriagePhase: {State: "succeeded", DurationMS: 12_000},
				RunPhase:    {State: "running", StartedAt: &started},
				VerifyPhase: {State: "failed", Progress: PhaseProgress{Done: 3, Failed: 1, Total: 4}},
			},
		},
	}

	rendered, err := clicky.Format(todos)
	require.NoError(t, err)
	t.Logf("\n%s", rendered)

	for _, header := range []string{"Plan", "Triage", "Run", "Verify"} {
		assert.Contains(t, rendered, header,
			"a phase column must survive a first row that never ran")
	}
	assert.Contains(t, rendered, "3/4", "verification progress reaches the table")
	assert.Contains(t, rendered, "2m", "a settled phase reports how long it took")
}
