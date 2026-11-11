// Reports Captain's budget rules for the platform dashboard: each live rule with the completed
// USD spend of every group inside its window, as Captain derives it from the model-call ledger.
package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flanksource/captain/pkg/budgets"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

// Budget is a live budget rule. WindowStart is nil, and WindowError says why, when its window
// can't be evaluated; Groups then is empty.
type Budget struct {
	ID          uuid.UUID                 `json:"id"`
	Name        string                    `json:"name"`
	Match       captaindb.BudgetRuleMatch `json:"match"`
	GroupBy     []string                  `json:"groupBy"`
	AmountUSD   float64                   `json:"amountUSD"`
	Window      string                    `json:"window"`
	WindowStart *time.Time                `json:"windowStart,omitempty"`
	WindowError string                    `json:"windowError,omitempty"`
	Groups      []BudgetGroup             `json:"groups"`
}

// BudgetGroup is the completed spend of one group of a rule inside its window. Error says why
// SpendUSD isn't its total.
type BudgetGroup struct {
	Values   map[string]string `json:"values"`
	SpendUSD float64           `json:"spendUSD"`
	Calls    int               `json:"calls"`
	Error    string            `json:"error,omitempty"`
}

// Budgets reads the live budget rules, by name, with their spend for windows ending at now.
func (s Store) Budgets(ctx context.Context, now time.Time) ([]Budget, error) {
	db := s.DB.WithContext(ctx)
	var rules []struct {
		ID      uuid.UUID
		Name    string
		Match   string
		GroupBy string
		Amount  float64
		Window  string
	}
	if err := db.Raw(`SELECT id, name, match::text AS match, group_by::text AS group_by, amount, "window"
		FROM captain_budget_rules WHERE deleted_at IS NULL ORDER BY name`).Scan(&rules).Error; err != nil {
		return nil, fmt.Errorf("read the budget rules: %w", err)
	}
	result := make([]Budget, 0, len(rules))
	for _, rule := range rules {
		budget := Budget{ID: rule.ID, Name: rule.Name, AmountUSD: rule.Amount, Window: rule.Window, Groups: []BudgetGroup{}}
		if err := json.Unmarshal([]byte(rule.Match), &budget.Match); err != nil {
			return nil, fmt.Errorf("read budget rule %s's match: %w", rule.Name, err)
		}
		if err := json.Unmarshal([]byte(rule.GroupBy), &budget.GroupBy); err != nil {
			return nil, fmt.Errorf("read budget rule %s's groupBy: %w", rule.Name, err)
		}
		start, err := budgets.WindowStart(rule.Window, now)
		if err != nil {
			budget.WindowError = err.Error()
			result = append(result, budget)
			continue
		}
		budget.WindowStart = &start
		if budget.Groups, err = s.budgetGroups(ctx, rule.ID, start); err != nil {
			return nil, fmt.Errorf("read budget rule %s's spend: %w", rule.Name, err)
		}
		result = append(result, budget)
	}
	return result, nil
}

// budgetGroups totals the rule's completed calls since start by group, as Captain's
// BudgetSpendUSD does: a group holding a call in another currency has no USD total.
func (s Store) budgetGroups(ctx context.Context, rule uuid.UUID, start time.Time) ([]BudgetGroup, error) {
	var rows []struct {
		Values        string
		SpendUSD      float64
		Calls         int
		OtherCurrency int
	}
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT budgets.group_values::text AS values, COALESCE(sum(`+captaindb.ModelCallCostUSD+`), 0) AS spend_usd,
			count(*) AS calls, count(*) FILTER (WHERE upper(calls.currency) <> 'USD') AS other_currency
		FROM captain_model_calls calls
		JOIN captain_model_call_budgets budgets ON budgets.model_call_id = calls.id
		WHERE budgets.budget_rule_id = ? AND calls.ended_at IS NOT NULL AND calls.ended_at >= ?
		GROUP BY budgets.group_values ORDER BY 1`, rule, start).Scan(&rows).Error; err != nil {
		return nil, err
	}
	groups := make([]BudgetGroup, 0, len(rows))
	for _, row := range rows {
		group := BudgetGroup{SpendUSD: row.SpendUSD, Calls: row.Calls}
		if err := json.Unmarshal([]byte(row.Values), &group.Values); err != nil {
			return nil, fmt.Errorf("read group %s: %w", row.Values, err)
		}
		if row.OtherCurrency > 0 {
			group.SpendUSD = 0
			group.Error = fmt.Sprintf("%d completed %s priced in another currency, which can't be added to USD spend",
				row.OtherCurrency, plural(row.OtherCurrency, "call is", "calls are"))
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
