// Tests the platform dashboard's model usage against a migrated database: every organization's
// model calls per day, model and provider, attributed through the session or its root session.
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

// addSession files a Captain session, under root when root isn't nil, owned by org when org isn't nil.
func addSession(t *testing.T, db *gorm.DB, root *uuid.UUID, org *uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO captain_sessions (id, source, root_session_id, parent_session_id) VALUES (?, 'test', ?, ?)`, id, root, root).Error)
	if org != nil {
		require.NoError(t, db.Exec(`INSERT INTO session_owners (session_id, org_id) VALUES (?, ?)`, id, *org).Error)
	}
	return id
}

// modelCall is one row of Captain's model-call ledger.
type modelCall struct {
	model, provider, currency                   string
	input, output, reasoning, cacheRead, cacheW int64
	listCost, providerCostUSD                   float64
	at                                          time.Time
	ended                                       *time.Time
}

// addCall files call in a new turn of session and returns the call's ID.
func addCall(t *testing.T, db *gorm.DB, session uuid.UUID, call modelCall) uuid.UUID {
	t.Helper()
	turn, id := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO captain_turns (id, session_id, turn_index, status)
		VALUES (?, ?, (SELECT count(*) FROM captain_turns WHERE session_id = ?), 'ended')`, turn, session, session).Error)
	require.NoError(t, db.Exec(`INSERT INTO captain_model_calls (id, turn_id, call_index, model, provider, currency,
			input_tokens, output_tokens, reasoning_tokens, cache_read_tokens, cache_write_tokens,
			input_cost, provider_cost_usd, created_at, started_at, ended_at)
		VALUES (?, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, turn, call.model, call.provider, call.currency,
		call.input, call.output, call.reasoning, call.cacheRead, call.cacheW,
		call.listCost, call.providerCostUSD, call.at, call.at, call.ended).Error)
	return id
}

func TestModelUsage(t *testing.T) {
	db := openDB(t)
	alpha := addOrg(t, db, issuer, "org_alpha", "alpha", "Alpha")
	other := addOrg(t, db, "https://other-clerk.example", "org_other", "other", "Another Clerk instance")
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	owned := addSession(t, db, nil, &alpha)
	child := addSession(t, db, &owned, nil)
	unowned := addSession(t, db, nil, nil)
	foreign := addSession(t, db, nil, &other)

	// Alpha: a list-priced call and a provider-priced one yesterday, the second in a child session.
	addCall(t, db, owned, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 100, output: 20, listCost: 0.5, at: yesterday.Add(time.Hour)})
	addCall(t, db, child, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 10, output: 5, cacheRead: 50, listCost: 9, providerCostUSD: 0.25, at: yesterday.Add(2 * time.Hour)})
	// Alpha today: a call priced in euros, whose tokens count but whose cost can't be added.
	addCall(t, db, owned, modelCall{model: "gpt-5", provider: "openai", currency: "EUR", input: 7, reasoning: 3, listCost: 4, at: today})
	// No owner, and another Clerk instance's organization: both unattributed.
	addCall(t, db, unowned, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 1, listCost: 1, at: today})
	addCall(t, db, foreign, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 2, listCost: 2, at: today})
	// Before the window.
	addCall(t, db, owned, modelCall{model: "claude-sonnet", provider: "anthropic", currency: "USD", input: 1000, listCost: 100, at: today.Add(-20 * 24 * time.Hour)})

	usage, err := platform.Store{DB: db, Issuer: issuer}.ModelUsage(t.Context(), today.Add(-7*24*time.Hour))
	require.NoError(t, err)

	assert.Equal(t, []platform.ModelUsageDay{
		{Day: yesterday.Format(time.DateOnly), Org: &alpha, OrgName: "Alpha", Model: "claude-sonnet", Provider: "anthropic",
			Calls: 2, InputTokens: 110, OutputTokens: 25, CacheReadTokens: 50, CostUSD: 0.75},
		{Day: today.Format(time.DateOnly), Org: &alpha, OrgName: "Alpha", Model: "gpt-5", Provider: "openai",
			Calls: 1, InputTokens: 7, ReasoningTokens: 3, OtherCurrencyCalls: 1},
		{Day: today.Format(time.DateOnly), Model: "claude-sonnet", Provider: "anthropic",
			Calls: 2, InputTokens: 3, CostUSD: 3},
	}, usage, "days in order, unattributed calls last; a provider-reported cost wins over list prices")
}
