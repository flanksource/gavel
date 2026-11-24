// Tests /api/budgets: every member of a hosted organization sees its budgets with how much of each
// is used; only its admins set and remove them.
package ui

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/flanksource/gavel/budget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminsSetAndRemoveBudgetsEveryMemberSees(t *testing.T) {
	h := newHostedRuns(t, nil)
	require.NoError(t, h.db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, cpu_millicores, memory_bytes, started_at)
		VALUES (gen_random_uuid(), ?, 'agent', 'running', 'substrate', 2000, ?, now())`, h.org, int64(4<<30)).Error)

	status, set := hostedCall(t, h.base, http.MethodPut, "/api/budgets/cpu/concurrent", h.ada, "org:admin", map[string]any{"amount": 8})
	require.Equal(t, http.StatusOK, status, set)
	status, set = hostedCall(t, h.base, http.MethodPut, "/api/budgets/usd/cumulative", h.ada, "org:admin", map[string]any{"amount": 50, "window": "now/M"})
	require.Equal(t, http.StatusOK, status, set)
	assert.Equal(t, "now/M", set["window"])

	status, body := hostedGet(t, h.base, "/api/budgets", h.grace)
	require.Equal(t, http.StatusOK, status, body)
	var got BudgetsResponse
	require.NoError(t, json.Unmarshal([]byte(body), &got))
	assert.False(t, got.CanEdit, "a member sees the budgets but can't change them")
	require.Len(t, got.Budgets, 2)
	assert.Equal(t, budget.CPU, got.Budgets[0].Resource)
	assert.Equal(t, 2.0, got.Budgets[0].Used, "the cores running now")
	assert.Equal(t, budget.USD, got.Budgets[1].Resource)
	assert.NotNil(t, got.Budgets[1].Since)

	status, removed := hostedCall(t, h.base, http.MethodDelete, "/api/budgets/cpu/concurrent", h.ada, "org:admin", nil)
	require.Equal(t, http.StatusOK, status, removed)
	status, removed = hostedCall(t, h.base, http.MethodDelete, "/api/budgets/cpu/concurrent", h.ada, "org:admin", nil)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, removed["error"], "has no concurrent cpu budget")
}

func TestOnlyAdminsChangeBudgets(t *testing.T) {
	h := newHostedRuns(t, nil)

	status, refused := hostedCall(t, h.base, http.MethodPut, "/api/budgets/cpu/concurrent", h.grace, "org:member", map[string]any{"amount": 8})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "only an organization admin can change its budgets", refused["error"])
	status, _ = hostedCall(t, h.base, http.MethodDelete, "/api/budgets/cpu/concurrent", h.grace, "org:member", nil)
	assert.Equal(t, http.StatusForbidden, status)
}

func TestBudgetsRefuseWhatCantBeABudget(t *testing.T) {
	h := newHostedRuns(t, nil)

	status, refused := hostedCall(t, h.base, http.MethodPut, "/api/budgets/tokens/concurrent", h.ada, "org:admin", map[string]any{"amount": 8})
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, refused["error"], "only cpu and memory have concurrent budgets")
	status, refused = hostedCall(t, h.base, http.MethodPut, "/api/budgets/cpu/cumulative", h.ada, "org:admin", "eight")
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, refused["error"], `a budget is JSON like {"amount": 40, "window": "now/M"}`)
}
