package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func finished(state string, ms int64) PhaseRun {
	started := time.Now().Add(-time.Duration(ms) * time.Millisecond)
	ended := started.Add(time.Duration(ms) * time.Millisecond)
	return PhaseRun{State: state, StartedAt: &started, FinishedAt: &ended, DurationMS: ms}
}

func TestPrettyRowPhaseColumns(t *testing.T) {
	// clicky derives table headers from the first row only, so every phase
	// column has to be present even for a todo that has never run — otherwise
	// the columns would appear or vanish with the sort order.
	t.Run("emits every phase column even when nothing has run", func(t *testing.T) {
		row := TODO{TODOFrontmatter: TODOFrontmatter{Title: "Bare"}}.PrettyRow(nil)
		for _, phase := range Phases {
			require.Contains(t, row, phaseColumn(phase))
			assert.Equal(t, "—", row[phaseColumn(phase)].String(),
				"a phase that never ran must read as not-started, not blank")
		}
	})

	t.Run("renders each phase's own outcome", func(t *testing.T) {
		row := TODO{
			TODOFrontmatter: TODOFrontmatter{Title: "Four phases"},
			PhaseRuns: PhaseRuns{
				PlanPhase:   finished("succeeded", 134_000),
				RunPhase:    finished("failed", 22_000),
				VerifyPhase: PhaseRun{State: "succeeded", Progress: PhaseProgress{Done: 3, Total: 4, Failed: 1}},
			},
		}.PrettyRow(nil)

		assert.Contains(t, row["Plan"].String(), "✓")
		assert.Contains(t, row["Run"].String(), "✗")
		assert.Contains(t, row["Verify"].String(), "3/4", "verification counts its fixture's checks")
		assert.Equal(t, "—", row["Triage"].String(), "triage never ran")
		assert.Contains(t, row["Run"].ANSI(), "\x1b[", "a failed phase must be coloured in the terminal")
	})
}

func TestPhaseRunPretty(t *testing.T) {
	t.Run("omits progress a single iteration cannot express", func(t *testing.T) {
		run := PhaseRun{State: "succeeded", Progress: PhaseProgress{Done: 1, Total: 1}}
		assert.NotContains(t, run.Pretty().String(), "1/1")
	})

	t.Run("reports failures even without a total", func(t *testing.T) {
		run := PhaseRun{State: "succeeded", Progress: PhaseProgress{Failed: 2}}
		assert.Contains(t, run.Pretty().String(), "2 failed")
	})

	// The recorded duration is a snapshot from when the row was read; a phase
	// still running has to be measured against now or the timer would freeze.
	t.Run("measures a running phase against now", func(t *testing.T) {
		started := time.Now().Add(-90 * time.Second)
		run := PhaseRun{State: "running", StartedAt: &started, DurationMS: 1_000}
		assert.Contains(t, run.Pretty().String(), "◐")
		assert.NotContains(t, run.Pretty().String(), "1s", "must not use the stale recorded duration")
	})
}

func TestPhaseRunOutcomeHelpers(t *testing.T) {
	// Verification that produced failing checks is a failure even though the
	// run itself succeeded in executing the fixture.
	assert.True(t, PhaseRun{State: "succeeded", Progress: PhaseProgress{Failed: 1}}.Failed())
	assert.False(t, PhaseRun{State: "succeeded", Progress: PhaseProgress{Done: 2, Total: 2}}.Failed())
	assert.True(t, PhaseRun{State: "cancelled"}.Failed())
	assert.True(t, PhaseRun{State: "running"}.Running())
	assert.False(t, PhaseRun{State: "succeeded"}.Running())
}

func TestPhaseRunsOrdered(t *testing.T) {
	runs := PhaseRuns{
		VerifyPhase: {Phase: VerifyPhase},
		PlanPhase:   {Phase: PlanPhase},
		RunPhase:    {Phase: RunPhase},
	}
	ordered := make([]Phase, 0, 3)
	for _, run := range runs.Ordered() {
		ordered = append(ordered, run.Phase)
	}
	// Pipeline order, not map order: what you plan, you run; what you run, you verify.
	assert.Equal(t, []Phase{PlanPhase, RunPhase, VerifyPhase}, ordered)
}

func TestPrettyDetailedListsOnlyPhasesThatRan(t *testing.T) {
	todo := TODO{
		TODOFrontmatter: TODOFrontmatter{Title: "Partly done"},
		PhaseRuns:       PhaseRuns{PlanPhase: finished("succeeded", 5_000)},
	}
	rendered := todo.PrettyDetailed().String()
	assert.Contains(t, rendered, "Plan:")
	assert.NotContains(t, rendered, "Triage:", "a phase that never ran is not worth a line")
}
