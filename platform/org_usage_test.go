// Tests one organization's usage in a window against a migrated database: the CPU and memory its
// runs were sized at for as long as they ran in it, and its model calls' tokens and cost.
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

// sizedRun is a run of an organization at a size, started and finished when set.
type sizedRun struct {
	status                string
	cpuMillicores, memory *int64
	started, finished     *time.Time
}

func addSizedRun(t *testing.T, db *gorm.DB, org uuid.UUID, run sizedRun) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, template, cpu_millicores, memory_bytes, started_at, finished_at)
		VALUES (?, ?, 'agent', ?, 'substrate', 'gavel-org-x', ?, ?, ?, ?)`, uuid.New(), org, run.status, run.cpuMillicores, run.memory, run.started, run.finished).Error)
}

func at(t time.Time) *time.Time { return &t }
func n(v int64) *int64          { return &v }

func TestOrgUsageCountsTheSizeOfEachRunForAsLongAsItRanInTheWindow(t *testing.T) {
	db := openDB(t)
	alpha := addOrg(t, db, issuer, "org_alpha", "alpha", "Alpha")
	beta := addOrg(t, db, issuer, "org_beta", "beta", "Beta")
	now := time.Now()
	since := now.Add(-6 * time.Hour)
	lifetime := 75 * time.Minute
	gib := int64(1 << 30)

	// An hour at 1 core and 2 GiB, inside the window.
	addSizedRun(t, db, alpha, sizedRun{status: "succeeded", cpuMillicores: n(1000), memory: n(2 * gib), started: at(now.Add(-2 * time.Hour)), finished: at(now.Add(-time.Hour))})
	// Half an hour of it in the window, at 2 cores and 4 GiB.
	addSizedRun(t, db, alpha, sizedRun{status: "failed", cpuMillicores: n(2000), memory: n(4 * gib), started: at(since.Add(-30 * time.Minute)), finished: at(since.Add(30 * time.Minute))})
	// Still running for ten minutes, at half a core and 1 GiB.
	addSizedRun(t, db, alpha, sizedRun{status: "running", cpuMillicores: n(500), memory: n(gib), started: at(now.Add(-10 * time.Minute))})
	// Left running when serve stopped three hours ago: it counts only as long as a run can live.
	addSizedRun(t, db, alpha, sizedRun{status: "running", cpuMillicores: n(1000), memory: n(gib), started: at(now.Add(-3 * time.Hour))})
	// Not counted: queued, unsized (another backend), before the window, and another organization's.
	addSizedRun(t, db, alpha, sizedRun{status: "queued", cpuMillicores: n(1000), memory: n(gib)})
	addSizedRun(t, db, alpha, sizedRun{status: "succeeded", started: at(now.Add(-time.Hour)), finished: at(now)})
	addSizedRun(t, db, alpha, sizedRun{status: "succeeded", cpuMillicores: n(1000), memory: n(gib), started: at(since.Add(-2 * time.Hour)), finished: at(since.Add(-time.Hour))})
	addSizedRun(t, db, beta, sizedRun{status: "running", cpuMillicores: n(8000), memory: n(8 * gib), started: at(now.Add(-time.Hour))})

	usage, err := platform.Store{DB: db, Issuer: issuer}.OrgUsage(t.Context(), alpha, since, lifetime)
	require.NoError(t, err)

	running := 10 * time.Minute.Seconds()
	stuck := lifetime.Seconds()
	assert.InDelta(t, 3600*1+1800*2+running*0.5+stuck*1, usage.Compute.CoreSeconds, 5)
	assert.InDelta(t, 3600*2+1800*4+running*1+stuck*1, usage.Compute.GiBSeconds, 10)
	assert.Equal(t, 4, usage.Compute.Runs)
	assert.Equal(t, platform.Running{Runs: 1, CPUMillicores: 500, MemoryBytes: gib}, usage.Compute.Running,
		"only the run still within a run's lifetime is running")
}

func TestOrgUsageCountsTheTokensAndCostOfItsModelCalls(t *testing.T) {
	db := openDB(t)
	alpha := addOrg(t, db, issuer, "org_alpha", "alpha", "Alpha")
	beta := addOrg(t, db, issuer, "org_beta", "beta", "Beta")
	now := time.Now()
	since := now.Add(-24 * time.Hour)

	owned := addSession(t, db, nil, &alpha)
	child := addSession(t, db, &owned, nil)
	other := addSession(t, db, nil, &beta)
	addCall(t, db, owned, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 100, output: 20, cacheRead: 5000, cacheW: 300, listCost: 0.5, at: now.Add(-time.Hour)})
	addCall(t, db, child, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 10, output: 5, listCost: 9, providerCostUSD: 0.25, at: now.Add(-time.Hour)})
	addCall(t, db, owned, modelCall{model: "gpt-5", provider: "openai", currency: "EUR", input: 7, reasoning: 3, listCost: 4, at: now})
	addCall(t, db, owned, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 1000, listCost: 100, at: since.Add(-time.Hour)})
	addCall(t, db, other, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 2, listCost: 2, at: now})

	usage, err := platform.Store{DB: db, Issuer: issuer}.OrgUsage(t.Context(), alpha, since, time.Hour)
	require.NoError(t, err)

	assert.Equal(t, platform.Tokens{Input: 117, Output: 25, Reasoning: 3, CacheRead: 5000, CacheWrite: 300, Counted: 117 + 25 + 3 + 300}, usage.Tokens,
		"counted tokens leave out cache reads, as providers' input limits do")
	assert.InDelta(t, 0.75, usage.CostUSD, 1e-9, "a provider-reported cost wins over the list price; the euro call adds none")
	assert.Equal(t, 3, usage.Calls)
	assert.Equal(t, 1, usage.OtherCurrencyCalls)
	assert.WithinDuration(t, since, usage.Since, time.Second)
}
