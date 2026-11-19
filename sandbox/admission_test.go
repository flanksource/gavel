// Tests admitting runs against their organization's budgets: refusing a run when a cumulative
// budget is used up or it can never fit a concurrent one, and runs waiting their turn otherwise.
package sandbox

import (
	"context"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/gavel/budget"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func cores(n float64) substrate.Size {
	return substrate.Size{CPUMillicores: int64(n * 1000), MemoryBytes: 1 << 30}
}

func newAdmission(t *testing.T) (*Admission, *gorm.DB, uuid.UUID) {
	t.Helper()
	_, db, _, org := newDispatcher(t, &memoryExecutor{})
	return &Admission{Budgets: budget.Store{DB: db}, RunLifetime: time.Hour, Wait: time.Minute, Recheck: time.Minute}, db, org
}

func setBudget(t *testing.T, a *Admission, org uuid.UUID, b budget.Budget) {
	t.Helper()
	_, err := a.Budgets.Set(context.Background(), org, nil, b)
	require.NoError(t, err)
}

// waitFor waits for ticket's room in the background; admitted reports whether it got it within a
// moment, leaving it in line when it didn't.
func waitFor(ticket *Ticket) <-chan error {
	done := make(chan error, 1)
	go func() { done <- ticket.Wait(context.Background(), func(string) {}) }()
	return done
}

func admitted(t *testing.T, waiting <-chan error) bool {
	t.Helper()
	select {
	case err := <-waiting:
		require.NoError(t, err)
		return true
	case <-time.After(200 * time.Millisecond):
		return false
	}
}

func TestARunWithoutBudgetsIsAdmittedAtOnce(t *testing.T) {
	a, _, org := newAdmission(t)

	ticket, err := a.Enter(context.Background(), org, cores(8))
	require.NoError(t, err)
	queued, _ := ticket.Queued()
	assert.False(t, queued)
	ticket.Release()
}

func TestARunIsRefusedWhenACumulativeBudgetIsUsedUp(t *testing.T) {
	a, db, org := newAdmission(t)
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, cpu_millicores, memory_bytes, started_at, finished_at)
		VALUES (?, ?, 'agent', 'succeeded', 'substrate', 4000, 1073741824, now() - interval '40 minutes', now() - interval '10 minutes')`, uuid.New(), org).Error)
	setBudget(t, a, org, budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 1.5, Window: "now-1d"})

	_, err := a.Enter(context.Background(), org, cores(1))
	require.ErrorIs(t, err, ErrBudgetUsedUp)
	assert.Contains(t, err.Error(), "this organization has used 2 core-hours of its 1.5 core-hours cpu budget since ")
	assert.Contains(t, err.Error(), "an organization admin can raise it")
}

func TestARunThatCanNeverFitAConcurrentBudgetIsRefused(t *testing.T) {
	a, _, org := newAdmission(t)
	setBudget(t, a, org, budget.Budget{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 1})
	setBudget(t, a, org, budget.Budget{Resource: budget.Memory, Kind: budget.Concurrent, Amount: 0.5})

	_, err := a.Enter(context.Background(), org, cores(2))
	require.ErrorIs(t, err, ErrBudgetUsedUp)
	assert.Contains(t, err.Error(), "a run of 2 cores can never start under this organization's concurrent cpu budget of 1 core")
	assert.Contains(t, err.Error(), "a run of 1 GiB can never start under this organization's concurrent memory budget of 0.5 GiB")
}

func TestRunsWaitTheirTurnWhileTheOrganizationHoldsTooMuch(t *testing.T) {
	a, _, org := newAdmission(t)
	setBudget(t, a, org, budget.Budget{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 2})
	ctx := context.Background()

	first, err := a.Enter(ctx, org, cores(2))
	require.NoError(t, err)
	second, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)
	queued, reason := second.Queued()
	assert.True(t, queued)
	assert.Equal(t, "queued: this organization's runs hold 2 of its 2 concurrent cores, and no run is ahead of this one", reason)
	third, err := a.Enter(ctx, org, cores(2))
	require.NoError(t, err)
	_, reason = third.Queued()
	assert.Equal(t, "queued: this organization's runs hold 2 of its 2 concurrent cores, and 1 run is ahead of this one", reason)
	fourth, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)
	secondWaits, thirdWaits, fourthWaits := waitFor(second), waitFor(third), waitFor(fourth)

	assert.False(t, admitted(t, secondWaits), "nothing has ended yet")
	first.Release()
	assert.True(t, admitted(t, secondWaits), "the first in line starts once there is room")
	assert.False(t, admitted(t, thirdWaits), "2 more cores don't fit beside the second's 1")
	assert.False(t, admitted(t, fourthWaits), "a smaller run doesn't overtake the one ahead of it")
	second.Release()
	assert.True(t, admitted(t, thirdWaits))
	third.Release()
	assert.True(t, admitted(t, fourthWaits))
	fourth.Release()
}

func TestAWaitingRunGivesUpAfterTheLongestWait(t *testing.T) {
	a, _, org := newAdmission(t)
	a.Wait = 50 * time.Millisecond
	setBudget(t, a, org, budget.Budget{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 1})
	ctx := context.Background()
	holding, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)
	defer holding.Release()
	waiting, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)

	err = waiting.Wait(ctx, func(string) {})
	require.ErrorIs(t, err, ErrQueueTimeout)
	assert.Contains(t, err.Error(), "waited 50ms for room and gave up: this organization's runs hold 1 of its 1 concurrent core")
	next, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)
	_, reason := next.Queued()
	assert.Contains(t, reason, "no run is ahead of this one", "a run that gave up leaves the line")
}

func TestAWaitingRunStartsWhenAnAdminRaisesTheBudget(t *testing.T) {
	a, _, org := newAdmission(t)
	a.Recheck = 20 * time.Millisecond
	setBudget(t, a, org, budget.Budget{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 1})
	ctx := context.Background()
	holding, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)
	defer holding.Release()
	waiting, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)

	waits := waitFor(waiting)
	setBudget(t, a, org, budget.Budget{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 2})
	assert.True(t, admitted(t, waits))
	waiting.Release()
}

func TestAWaitingRunIsRefusedWhenACumulativeBudgetIsUsedUpMeanwhile(t *testing.T) {
	a, db, org := newAdmission(t)
	a.Recheck = 20 * time.Millisecond
	setBudget(t, a, org, budget.Budget{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 1})
	ctx := context.Background()
	holding, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)
	defer holding.Release()
	waiting, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)

	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, cpu_millicores, memory_bytes, started_at, finished_at)
		VALUES (?, ?, 'agent', 'succeeded', 'substrate', 4000, 1073741824, now() - interval '40 minutes', now() - interval '10 minutes')`, uuid.New(), org).Error)
	setBudget(t, a, org, budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 1, Window: "now-1d"})

	waitCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	assert.ErrorIs(t, waiting.Wait(waitCtx, func(string) {}), ErrBudgetUsedUp)
}

func TestAWaitingRunReportsWhatItWaitsFor(t *testing.T) {
	a, _, org := newAdmission(t)
	setBudget(t, a, org, budget.Budget{Resource: budget.Memory, Kind: budget.Concurrent, Amount: 1})
	ctx := context.Background()
	holding, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)
	waiting, err := a.Enter(ctx, org, cores(1))
	require.NoError(t, err)
	var reports []string
	done := make(chan error)
	go func() { done <- waiting.Wait(ctx, func(report string) { reports = append(reports, report) }) }()

	holding.Release()
	require.NoError(t, <-done)
	assert.Equal(t, []string{"queued: this organization's runs hold 1 of its 1 concurrent GiB, and no run is ahead of this one", "started after waiting for room"}, reports)
	waiting.Release()
}
