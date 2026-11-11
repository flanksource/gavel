// Tests the platform dashboard's budgets against a migrated database: each live budget rule with
// the completed spend of every group inside its window, and why a window or a group has no total.
package platform_test

import (
	"testing"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/platform"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func addBudgetRule(t *testing.T, db *gorm.DB, name, window string, amount float64, deleted bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	var deletedAt *time.Time
	if deleted {
		at := time.Now()
		deletedAt = &at
	}
	require.NoError(t, db.Exec(`INSERT INTO captain_budget_rules (id, name, match, group_by, amount, "window", deleted_at)
		VALUES (?, ?, '{"models":["anthropic/*"]}', '["org"]', ?, ?, ?)`, id, name, amount, window, deletedAt).Error)
	return id
}

func attribute(t *testing.T, db *gorm.DB, call, rule uuid.UUID, group string) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO captain_model_call_budgets (model_call_id, budget_rule_id, group_values) VALUES (?, ?, ?::jsonb)`, call, rule, group).Error)
}

func TestBudgets(t *testing.T) {
	db := openDB(t)
	session := addSession(t, db, nil, nil)
	now := time.Date(2026, 11, 11, 15, 0, 0, 0, time.UTC)
	ended := func(at time.Time) *time.Time { return &at }
	usd := func(cost float64, end *time.Time) modelCall {
		return modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", listCost: cost, at: now.Add(-25 * time.Hour), ended: end}
	}

	daily := addBudgetRule(t, db, "Daily per org", "now/d", 10, false)
	// Alpha: two completed calls today, one completed yesterday (outside the window), one unfinished.
	for _, call := range []modelCall{usd(1, ended(now.Add(-time.Hour))), usd(2.5, ended(now.Add(-30*time.Minute))), usd(100, ended(now.Add(-24*time.Hour))), usd(50, nil)} {
		attribute(t, db, addCall(t, db, session, call), daily, `{"org":"alpha"}`)
	}
	// Beta: a completed call priced in euros, which can't be added to USD spend.
	euros := modelCall{model: "claude-sonnet", provider: "anthropic", currency: "EUR", listCost: 3, at: now.Add(-time.Hour), ended: ended(now.Add(-time.Hour))}
	attribute(t, db, addCall(t, db, session, euros), daily, `{"org":"beta"}`)

	monthly := addBudgetRule(t, db, "Monthly", "now/M", 500, false)
	broken := addBudgetRule(t, db, "Broken", "2026-11-01", 1, false)
	addBudgetRule(t, db, "Deleted", "now/d", 1, true)

	budgets, err := platform.Store{DB: db, Issuer: issuer}.Budgets(t.Context(), now)
	require.NoError(t, err)

	today := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	month := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	match := captaindb.BudgetRuleMatch{Models: []string{"anthropic/*"}}
	assert.Equal(t, []platform.Budget{
		{ID: broken, Name: "Broken", Match: match, GroupBy: []string{"org"}, AmountUSD: 1, Window: "2026-11-01",
			WindowError: `budget window "2026-11-01" must be relative to now`, Groups: []platform.BudgetGroup{}},
		{ID: daily, Name: "Daily per org", Match: match, GroupBy: []string{"org"}, AmountUSD: 10, Window: "now/d", WindowStart: &today,
			Groups: []platform.BudgetGroup{
				{Values: map[string]string{"org": "alpha"}, SpendUSD: 3.5, Calls: 2},
				{Values: map[string]string{"org": "beta"}, Calls: 1,
					Error: "1 completed call is priced in another currency, which can't be added to USD spend"},
			}},
		{ID: monthly, Name: "Monthly", Match: match, GroupBy: []string{"org"}, AmountUSD: 500, Window: "now/M", WindowStart: &month,
			Groups: []platform.BudgetGroup{}},
	}, budgets, "live rules by name; only completed calls inside the window count")
}
