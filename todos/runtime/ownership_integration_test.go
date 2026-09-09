package runtime

import (
	"os"
	"testing"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// abandonRun makes a run look like one whose dispatcher exited: the claim names
// a PID that is not running. That is what a killed `gavel serve` leaves behind,
// minus the seven hours of waiting.
func abandonRun(t *testing.T, provider *Provider, promptRunID string) {
	t.Helper()
	const deadPID = 0x7FFFFFFE
	require.NoError(t, provider.db.Exec(`
		UPDATE todo_issue_prompt_runs
		SET owner_pid = ?, owner_started_at = now() - interval '1 hour', owner_heartbeat_at = now() - interval '1 hour'
		WHERE prompt_run_id = ?`, deadPID, promptRunID).Error)
}

// TestOrphanedRunIsReclaimedOnDispatch is the regression oracle for a TODO that
// could never be run again: its dispatcher exited without finalizing the run, so
// the run stayed non-terminal and Captain's active-run index rejected every
// later dispatch with a bare "another active run exists" conflict.
func TestOrphanedRunIsReclaimedOnDispatch(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
	}
	provider := newResumeTestProvider(t, "gavel_todo_ownership")

	todo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Reclaim an abandoned run", Body: "The dispatcher died mid-run", Status: types.StatusPending,
	})
	require.NoError(t, err)

	first, err := provider.PrepareRun(t.Context(), todo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "headless-codex",
	})
	require.NoError(t, err)
	require.NotEqual(t, "", first.PromptRunID.String())

	// The dispatcher exits: its heartbeat stops and the run is left non-terminal,
	// still claimed by a process that no longer exists.
	provider.ownership.stop(first.PromptRunID)
	abandonRun(t, provider, first.PromptRunID.String())

	second, err := provider.PrepareRun(t.Context(), todo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "headless-codex",
	})
	require.NoError(t, err, "an abandoned run must not block the todo forever")
	assert.NotEqual(t, first.PromptRunID, second.PromptRunID, "the redispatch is a new run")

	reclaimed, err := provider.Captain().GetPromptRun(t.Context(), first.PromptRunID)
	require.NoError(t, err)
	assert.Equal(t, captaindb.PromptRunStateCancelled, reclaimed.State)
	assert.Contains(t, reclaimed.Error, "reclaimed:", "a reclaimed run must say it was reclaimed")
	assert.Contains(t, reclaimed.Error, "has exited", "and must say why its owner was judged gone")
}

// TestLiveRunRefusesDispatchUntilConfirmed covers the other half: a run whose
// dispatcher is still running is not reclaimable, and the second dispatch is a
// decision for the caller rather than something to do silently.
func TestLiveRunRefusesDispatchUntilConfirmed(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
	}
	provider := newResumeTestProvider(t, "gavel_todo_ownership_live")

	todo, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Refuse a silent second run", Body: "The first run is still going", Status: types.StatusPending,
	})
	require.NoError(t, err)

	first, err := provider.PrepareRun(t.Context(), todo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "headless-codex",
	})
	require.NoError(t, err)

	_, err = provider.PrepareRun(t.Context(), todo, todos.RunPreparation{
		Mode: types.ModeRun, ExecutorName: "headless-codex",
	})
	var owned *todos.ErrRunOwnedElsewhere
	require.ErrorAs(t, err, &owned, "a live run must be reported, not reclaimed")
	assert.Equal(t, first.PromptRunID, owned.PromptRunID)
	assert.Contains(t, owned.Error(), "--force")

	stillRunning, err := provider.Captain().GetPromptRun(t.Context(), first.PromptRunID)
	require.NoError(t, err)
	assert.Equal(t, captaindb.PromptRunStatePending, stillRunning.State, "the incumbent must be untouched")
}
