// Tests an organization's budgets against a migrated database: setting, listing and removing them,
// refusing what can't be a budget, and each one's use in its window.
package budget_test

import (
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/budget"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openDB(t *testing.T) (*gorm.DB, uuid.UUID, uuid.UUID) {
	t.Helper()
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_budget"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org, user := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, user).Error)
	return db, org, user
}

func TestSetListAndRemoveAnOrganizationsBudgets(t *testing.T) {
	db, org, user := openDB(t)
	store := budget.Store{DB: db}
	ctx := t.Context()

	monthly, err := store.Set(ctx, org, &user, budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 40, Window: "now/M"})
	require.NoError(t, err)
	assert.Equal(t, &user, monthly.UpdatedBy)
	assert.WithinDuration(t, time.Now(), monthly.UpdatedAt, time.Minute)
	_, err = store.Set(ctx, org, &user, budget.Budget{Resource: budget.Memory, Kind: budget.Concurrent, Amount: 16})
	require.NoError(t, err)
	_, err = store.Set(ctx, org, &user, budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 60, Window: "now-30d"})
	require.NoError(t, err, "setting a budget again replaces it")

	budgets, err := store.List(ctx, org)
	require.NoError(t, err)
	require.Len(t, budgets, 2)
	assert.Equal(t, budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 60, Window: "now-30d"}, strip(budgets[0]))
	assert.Equal(t, budget.Budget{Resource: budget.Memory, Kind: budget.Concurrent, Amount: 16}, strip(budgets[1]))

	require.NoError(t, store.Remove(ctx, org, budget.CPU, budget.Cumulative))
	budgets, err = store.List(ctx, org)
	require.NoError(t, err)
	assert.Len(t, budgets, 1)
	err = store.Remove(ctx, org, budget.CPU, budget.Cumulative)
	require.ErrorIs(t, err, budget.ErrNotFound)
	assert.Contains(t, err.Error(), "organization "+org.String()+" has no cumulative cpu budget")
}

func strip(b budget.Budget) budget.Budget {
	b.UpdatedAt, b.UpdatedBy = time.Time{}, nil
	return b
}

func TestSetRefusesWhatCantBeABudget(t *testing.T) {
	db, org, _ := openDB(t)
	store := budget.Store{DB: db}
	for _, test := range []struct {
		budget budget.Budget
		want   string
	}{
		{budget.Budget{Resource: "gpu", Kind: budget.Cumulative, Amount: 1, Window: "now/M"}, `resource "gpu" is not one of cpu, memory, tokens, usd`},
		{budget.Budget{Resource: budget.CPU, Kind: "daily", Amount: 1, Window: "now/M"}, `kind "daily" is not one of cumulative, concurrent`},
		{budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 0, Window: "now/M"}, "amount must be more than 0, got 0"},
		{budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 1}, "a cumulative budget needs a window, such as now/M"},
		{budget.Budget{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 1, Window: "2026-11-01"}, `budget window "2026-11-01" must be relative to now`},
		{budget.Budget{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 1, Window: "now/M"}, "a concurrent budget caps what runs hold at once and takes no window"},
		{budget.Budget{Resource: budget.Tokens, Kind: budget.Concurrent, Amount: 1}, "only cpu and memory have concurrent budgets: tokens are spent, not held"},
	} {
		_, err := store.Set(t.Context(), org, nil, test.budget)
		require.ErrorIs(t, err, budget.ErrInvalid, test.want)
		assert.Contains(t, err.Error(), test.want)
	}
	budgets, err := store.List(t.Context(), org)
	require.NoError(t, err)
	assert.Empty(t, budgets)
}

func TestStatusIsEachBudgetsUseInItsWindow(t *testing.T) {
	db, org, _ := openDB(t)
	store := budget.Store{DB: db}
	ctx := t.Context()
	gib := int64(1 << 30)
	// Twenty minutes at 3 cores and 6 GiB, finished, and still running at 1 core and 2 GiB.
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, cpu_millicores, memory_bytes, started_at, finished_at)
		VALUES (?, ?, 'agent', 'succeeded', 'substrate', 3000, ?, now() - interval '30 minutes', now() - interval '10 minutes')`, uuid.New(), org, 6*gib).Error)
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, cpu_millicores, memory_bytes, started_at)
		VALUES (?, ?, 'agent', 'running', 'substrate', 1000, ?, now() - interval '6 minutes')`, uuid.New(), org, 2*gib).Error)
	session, turn := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO captain_sessions (id, source) VALUES (?, 'test')`, session).Error)
	require.NoError(t, db.Exec(`INSERT INTO session_owners (session_id, org_id) VALUES (?, ?)`, session, org).Error)
	require.NoError(t, db.Exec(`INSERT INTO captain_turns (id, session_id, turn_index, status) VALUES (?, ?, 0, 'ended')`, turn, session).Error)
	require.NoError(t, db.Exec(`INSERT INTO captain_model_calls (id, turn_id, call_index, model, currency, input_tokens, output_tokens, cache_read_tokens, input_cost, created_at)
		VALUES (?, ?, 0, 'claude-sonnet', 'USD', 900, 100, 50000, 2.5, now())`, uuid.New(), turn).Error)
	for _, b := range []budget.Budget{
		{Resource: budget.CPU, Kind: budget.Cumulative, Amount: 2, Window: "now-1d"},
		{Resource: budget.Memory, Kind: budget.Cumulative, Amount: 10, Window: "now-1d"},
		{Resource: budget.CPU, Kind: budget.Concurrent, Amount: 4},
		{Resource: budget.Memory, Kind: budget.Concurrent, Amount: 2},
		{Resource: budget.Tokens, Kind: budget.Cumulative, Amount: 5000, Window: "now-1d"},
		{Resource: budget.USD, Kind: budget.Cumulative, Amount: 2, Window: "now-1d"},
	} {
		_, err := store.Set(ctx, org, nil, b)
		require.NoError(t, err)
	}

	statuses, err := store.Status(ctx, org, time.Hour)
	require.NoError(t, err)
	used := map[string]float64{}
	for _, s := range statuses {
		used[string(s.Kind)+" "+string(s.Resource)] = s.Used
		if s.Kind == budget.Cumulative {
			require.NotNil(t, s.Since)
			assert.WithinDuration(t, time.Now().Add(-24*time.Hour), *s.Since, time.Minute)
		} else {
			assert.Nil(t, s.Since)
		}
	}
	assert.InDelta(t, 1+0.1, used["cumulative cpu"], 0.01, "20 minutes at 3 cores and 6 minutes at 1 core, in core-hours")
	assert.InDelta(t, 2+0.2, used["cumulative memory"], 0.02, "in GiB-hours")
	assert.Equal(t, 1.0, used["concurrent cpu"], "the cores running now")
	assert.Equal(t, 2.0, used["concurrent memory"], "the GiB running now")
	assert.Equal(t, 1000.0, used["cumulative tokens"], "cache reads don't count")
	assert.Equal(t, 2.5, used["cumulative usd"])
	for _, s := range statuses {
		assert.Equal(t, s.Used >= s.Amount, s.Reached(), "%s %s", s.Kind, s.Resource)
	}
}
