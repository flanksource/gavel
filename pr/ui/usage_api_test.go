// Tests /api/usage: what a hosted organization used in a window, as every member sees it: its
// runs' CPU and memory and its model calls' tokens and cost.
package ui

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageShowsAMemberWhatTheOrganizationUsed(t *testing.T) {
	h := newHostedRuns(t, nil)
	require.NoError(t, h.db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, template, cpu_millicores, memory_bytes, started_at, finished_at)
		VALUES (?, ?, 'agent', 'succeeded', 'substrate', 'gavel-org-x', 2000, ?, now() - interval '20 minutes', now() - interval '10 minutes')`, uuid.New(), h.org, int64(4<<30)).Error)
	session, turn := uuid.New(), uuid.New()
	require.NoError(t, h.db.Exec(`INSERT INTO captain_sessions (id, source) VALUES (?, 'test')`, session).Error)
	require.NoError(t, h.db.Exec(`INSERT INTO session_owners (session_id, org_id) VALUES (?, ?)`, session, h.org).Error)
	require.NoError(t, h.db.Exec(`INSERT INTO captain_turns (id, session_id, turn_index, status) VALUES (?, ?, 0, 'ended')`, turn, session).Error)
	require.NoError(t, h.db.Exec(`INSERT INTO captain_model_calls (id, turn_id, call_index, model, currency, input_tokens, output_tokens, cache_read_tokens, input_cost, created_at)
		VALUES (?, ?, 0, 'claude-sonnet', 'USD', 100, 20, 900, 0.5, now())`, uuid.New(), turn).Error)

	status, body := hostedGet(t, h.base, "/api/usage?window=now-1d", h.grace)
	require.Equal(t, http.StatusOK, status, body)
	var got UsageResponse
	require.NoError(t, json.Unmarshal([]byte(body), &got))
	assert.Equal(t, "now-1d", got.Window)
	assert.InDelta(t, 2*600, got.Usage.Compute.CoreSeconds, 5, "ten minutes at 2 cores")
	assert.InDelta(t, 4*600, got.Usage.Compute.GiBSeconds, 5, "ten minutes at 4 GiB")
	assert.Equal(t, int64(120), got.Usage.Tokens.Counted, "cache reads aren't counted")
	assert.Equal(t, int64(900), got.Usage.Tokens.CacheRead, "but are shown")
	assert.InDelta(t, 0.5, got.Usage.CostUSD, 1e-9)

	status, body = hostedGet(t, h.base, "/api/usage", h.grace)
	require.Equal(t, http.StatusOK, status, body)
	require.NoError(t, json.Unmarshal([]byte(body), &got))
	assert.Equal(t, "now/M", got.Window, "the window is this month unless asked otherwise")
}

func TestUsageRefusesAWindowItCantRead(t *testing.T) {
	h := newHostedRuns(t, nil)

	status, body := hostedGet(t, h.base, "/api/usage?window=2026-11-01", h.grace)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, `budget window \"2026-11-01\" must be relative to now`)
}

func TestUsageNeedsAHostedSandbox(t *testing.T) {
	db, _, org := openOrgDB(t)
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)

	status, body := hostedGet(t, serveActingAs(t, srv, org), "/api/usage", uuid.New())
	assert.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, body, "hosted dashboard with a sandbox")
}
