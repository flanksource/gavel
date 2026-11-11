// Tests the platform dashboard's Captain history against a migrated database: every organization's
// prompt runs per day by origin and state, the latest failures, and the sessions live now.
package platform_test

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/platform"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// promptRun is one row of Captain's prompt runs.
type promptRun struct {
	session, root     uuid.UUID
	origin, state     string
	err               string
	queued            time.Time
	started, finished *time.Time
}

func addPromptRun(t *testing.T, db *gorm.DB, run promptRun) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var errText *string
	if run.err != "" {
		errText = &run.err
	}
	require.NoError(t, db.Exec(`INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin, state, error, queued_at, started_at, finished_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, run.session, run.root, run.origin, run.state, errText, run.queued, run.started, run.finished, run.queued).Error)
	return id
}

func setSessionState(t *testing.T, db *gorm.DB, session uuid.UUID, lifecycle, health string) {
	t.Helper()
	require.NoError(t, db.Exec(`UPDATE captain_sessions SET lifecycle_status = ?, health_state = ? WHERE id = ?`, lifecycle, health, session).Error)
}

func TestCaptainHistory(t *testing.T) {
	db := openDB(t)
	alpha := addOrg(t, db, issuer, "org_alpha", "alpha", "Alpha")
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	at := func(day time.Time, offset time.Duration) *time.Time { t := day.Add(offset); return &t }

	owned := addSession(t, db, nil, &alpha)
	child := addSession(t, db, &owned, nil)
	unowned := addSession(t, db, nil, nil)
	busy := addSession(t, db, nil, &alpha)

	// Alpha yesterday: a todo run that ran 60s, and one in a child session that failed after 30s.
	addPromptRun(t, db, promptRun{session: owned, root: owned, origin: "todo", state: "succeeded",
		queued: yesterday.Add(time.Hour), started: at(yesterday, time.Hour), finished: at(yesterday, time.Hour+time.Minute)})
	failed := addPromptRun(t, db, promptRun{session: child, root: owned, origin: "todo", state: "failed", err: "the model refused",
		queued: yesterday.Add(2 * time.Hour), started: at(yesterday, 2*time.Hour), finished: at(yesterday, 2*time.Hour+30*time.Second)})
	// Alpha today: a run still going in a session of its own.
	addPromptRun(t, db, promptRun{session: busy, root: busy, origin: "agent", state: "running", queued: today, started: at(today, time.Second)})
	// Unattributed today: a failure with no session owner.
	unownedFailure := addPromptRun(t, db, promptRun{session: unowned, root: unowned, origin: "cli", state: "failed", err: "no model key",
		queued: today.Add(time.Minute), started: at(today, time.Minute), finished: at(today, 2*time.Minute)})
	// Before the window.
	addPromptRun(t, db, promptRun{session: unowned, root: unowned, origin: "cli", state: "failed", err: "old",
		queued: today.Add(-20 * 24 * time.Hour), started: at(today, -20*24*time.Hour), finished: at(today, -20*24*time.Hour+time.Second)})

	// Live top-level sessions: busy is running and healthy, owned is running but stalled; child
	// sessions aren't counted, and finished ones aren't live.
	setSessionState(t, db, busy, "running", "healthy")
	setSessionState(t, db, owned, "running", "stalled")
	setSessionState(t, db, child, "running", "healthy")
	setSessionState(t, db, unowned, "failed", "healthy")

	history, err := platform.Store{DB: db, Issuer: issuer}.CaptainHistory(t.Context(), today.Add(-7*24*time.Hour))
	require.NoError(t, err)

	day, now := yesterday.Format(time.DateOnly), today.Format(time.DateOnly)
	assert.Equal(t, []platform.CaptainRunDay{
		{Day: day, Org: &alpha, OrgName: "Alpha", Origin: "todo", State: "succeeded", Runs: 1, RunSeconds: 60},
		{Day: day, Org: &alpha, OrgName: "Alpha", Origin: "todo", State: "failed", Runs: 1, RunSeconds: 30},
		{Day: now, Org: &alpha, OrgName: "Alpha", Origin: "agent", State: "running", Runs: 1},
		{Day: now, Origin: "cli", State: "failed", Runs: 1, RunSeconds: 60},
	}, history.Days, "days in order, unattributed runs last, states in Captain's order; a run still going adds no run time")

	require.Len(t, history.Failures, 2)
	assert.Equal(t, unownedFailure, history.Failures[0].ID, "newest failure first")
	assert.Nil(t, history.Failures[0].Org)
	assert.Equal(t, "no model key", history.Failures[0].Error)
	assert.Equal(t, failed, history.Failures[1].ID)
	assert.Equal(t, &alpha, history.Failures[1].Org)
	assert.Equal(t, "the model refused", history.Failures[1].Error)

	assert.ElementsMatch(t, []platform.CaptainLiveSessions{
		{Org: &alpha, OrgName: "Alpha", Health: "healthy", Sessions: 1},
		{Org: &alpha, OrgName: "Alpha", Health: "stalled", Sessions: 1},
	}, history.Live)
}
