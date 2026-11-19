// Tests the hard stop: once one of an organization's cumulative budgets is used up, its running
// runs are stopped, saying why: every run for CPU and memory, agent runs for tokens and model cost.
package sandbox

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/flanksource/gavel/budget"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// holdingExecutor runs each job until it is cancelled.
type holdingExecutor struct {
	mu        sync.Mutex
	cancelled []string
}

func (e *holdingExecutor) Name() string { return "holding" }

func (e *holdingExecutor) Submit(_ context.Context, job Job) (Handle, error) {
	return &holdingHandle{executor: e, ref: JobName(job.RunID), stop: make(chan struct{})}, nil
}

type holdingHandle struct {
	executor *holdingExecutor
	ref      string
	once     sync.Once
	stop     chan struct{}
}

func (h *holdingHandle) Logs(context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (h *holdingHandle) Wait(ctx context.Context) (Result, error) {
	select {
	case <-h.stop:
		return Result{}, errors.New("the job was killed")
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

func (h *holdingHandle) Cancel(context.Context) error {
	h.once.Do(func() {
		h.executor.mu.Lock()
		h.executor.cancelled = append(h.executor.cancelled, h.ref)
		h.executor.mu.Unlock()
		close(h.stop)
	})
	return nil
}

func (h *holdingHandle) Ref() string { return h.ref }

// spend files a model call of org costing usd with tokens input tokens, now.
func spend(t *testing.T, db *gorm.DB, org uuid.UUID, usd float64, tokens int64) {
	t.Helper()
	session, turn := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO captain_sessions (id, source) VALUES (?, 'test')`, session).Error)
	require.NoError(t, db.Exec(`INSERT INTO session_owners (session_id, org_id) VALUES (?, ?)`, session, org).Error)
	require.NoError(t, db.Exec(`INSERT INTO captain_turns (id, session_id, turn_index, status) VALUES (?, ?, 0, 'ended')`, turn, session).Error)
	require.NoError(t, db.Exec(`INSERT INTO captain_model_calls (id, turn_id, call_index, model, currency, input_tokens, input_cost, created_at)
		VALUES (?, ?, 0, 'claude-sonnet', 'USD', ?, ?, now())`, uuid.New(), turn, tokens, usd).Error)
}

// useCores files a finished run of org that used hours core-hours: half an hour at twice as many
// cores, well within a run's lifetime.
func useCores(t *testing.T, db *gorm.DB, org uuid.UUID, hours float64) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, cpu_millicores, memory_bytes, started_at, finished_at)
		VALUES (?, ?, 'lint', 'succeeded', 'substrate', ?, 1073741824, now() - interval '30 minutes', now())`, uuid.New(), org, int64(hours*2000)).Error)
}

func TestEnforceStopsAnOrganizationsRunsOnceACPUBudgetIsUsedUp(t *testing.T) {
	executor := &holdingExecutor{}
	dispatcher, db, org := newBudgetedDispatcher(t, executor)
	ctx := context.Background()
	job, err := dispatcher.Start(ctx, tenant.Scope{OrgID: org}, RunSpec{Kind: "lint"}, nil)
	require.NoError(t, err)
	agent, err := dispatcher.RecordRun(ctx, uuid.New(), tenant.Scope{OrgID: org}, RunSpec{Kind: "agent"})
	require.NoError(t, err)
	other, otherOrg := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, otherOrg, otherOrg.String()).Error)
	untouched, err := dispatcher.RecordRun(ctx, other, tenant.Scope{OrgID: otherOrg}, RunSpec{Kind: "agent"})
	require.NoError(t, err)
	defer untouched.Finish(true)

	require.NoError(t, dispatcher.Enforce(ctx, org))
	assert.NoError(t, agent.Context().Err(), "within its budgets nothing is stopped")

	useCores(t, db, org, 3)
	_, err = dispatcher.Admission.Budgets.Set(ctx, org, nil, budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 2, Window: "now-1d"})
	require.NoError(t, err)
	require.NoError(t, dispatcher.Enforce(ctx, org))

	_, err = job.Wait(ctx)
	require.Error(t, err, "the job was stopped")
	status, message := runMessage(t, db, job.ID)
	assert.Equal(t, "cancelled", status)
	assert.Contains(t, message, "stopped: budget used up: this organization has used ")
	assert.Contains(t, message, "of its 2 core-hours cpu budget")
	require.Error(t, agent.Context().Err(), "the agent run's context ends")
	assert.Contains(t, context.Cause(agent.Context()).Error(), "of its 2 core-hours cpu budget", "with why as its cause")
	agent.Finish(false)
	status, _ = runMessage(t, db, agent.ID)
	assert.Equal(t, "cancelled", status)
	assert.NoError(t, untouched.Context().Err(), "another organization's runs go on")
}

func TestEnforceStopsOnlyAgentRunsOnceAModelBudgetIsUsedUp(t *testing.T) {
	executor := &holdingExecutor{}
	dispatcher, db, org := newBudgetedDispatcher(t, executor)
	ctx := context.Background()
	job, err := dispatcher.Start(ctx, tenant.Scope{OrgID: org}, RunSpec{Kind: "lint"}, nil)
	require.NoError(t, err)
	agent, err := dispatcher.RecordRun(ctx, uuid.New(), tenant.Scope{OrgID: org}, RunSpec{Kind: "agent"})
	require.NoError(t, err)

	spend(t, db, org, 6, 100)
	_, err = dispatcher.Admission.Budgets.Set(ctx, org, nil, budget.Budget{Resource: budget.USD, Kind: budget.Cumulative, Amount: 5, Window: "now-1d"})
	require.NoError(t, err)
	require.NoError(t, dispatcher.Enforce(ctx, org))

	require.Error(t, agent.Context().Err())
	assert.Contains(t, context.Cause(agent.Context()).Error(), "this organization has used $6.00 of its $5.00 model cost budget")
	agent.Finish(false)
	status, _ := runMessage(t, db, job.ID)
	assert.Equal(t, "running", status, "a job makes no model calls")
	require.NoError(t, job.Cancel(ctx))
	_, _ = job.Wait(ctx)
}

func TestWatchBudgetsStopsRunsOnItsInterval(t *testing.T) {
	executor := &holdingExecutor{}
	dispatcher, db, org := newBudgetedDispatcher(t, executor)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	job, err := dispatcher.Start(ctx, tenant.Scope{OrgID: org}, RunSpec{Kind: "test"}, nil)
	require.NoError(t, err)
	go dispatcher.WatchBudgets(ctx, 20*time.Millisecond)

	useCores(t, db, org, 3)
	_, err = dispatcher.Admission.Budgets.Set(ctx, org, nil, budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 1, Window: "now-1d"})
	require.NoError(t, err)

	waitCtx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	_, err = job.Wait(waitCtx)
	require.Error(t, err)
	require.NotErrorIs(t, err, context.DeadlineExceeded, "the watch stopped the job")
	status, _ := runMessage(t, db, job.ID)
	assert.Equal(t, "cancelled", status)
}

func TestRemainingModelCostIsWhatTheModelCostBudgetLeaves(t *testing.T) {
	dispatcher, db, org := newBudgetedDispatcher(t, &holdingExecutor{})
	ctx := context.Background()

	_, capped, err := dispatcher.RemainingModelCost(ctx, org)
	require.NoError(t, err)
	assert.False(t, capped, "without a model cost budget nothing caps a run's spend")

	spend(t, db, org, 1.25, 10)
	_, err = dispatcher.Admission.Budgets.Set(ctx, org, nil, budget.Budget{Resource: budget.USD, Kind: budget.Cumulative, Amount: 5, Window: "now/M"})
	require.NoError(t, err)
	remaining, capped, err := dispatcher.RemainingModelCost(ctx, org)
	require.NoError(t, err)
	assert.True(t, capped)
	assert.InDelta(t, 3.75, remaining, 1e-9)
}
