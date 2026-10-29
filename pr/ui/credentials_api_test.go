// Tests /api/credentials/fetch: the credential provider gets an organization's model key only
// with its token, for an active, recent run, asked for by that run's own substrate environment.
package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialFetch(t *testing.T) {
	useModelKeyKeeper(t)
	db, _, org := openOrgDB(t)
	keyless := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Keyless')`, keyless, keyless.String()).Error)
	require.NoError(t, modelkeys.Store{DB: db}.Set(t.Context(), org, nil, "anthropic-api-key", modelKeysTestAPIKey))
	run := func(org uuid.UUID, status string, age time.Duration) uuid.UUID {
		id := uuid.New()
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, spec, created_at)
			VALUES (?, ?, 'agent', ?, 'substrate', '{"kind":"agent"}', now() - make_interval(secs => ?))`, id, org, status, age.Seconds()).Error)
		return id
	}
	running, queued := run(org, "running", time.Minute), run(org, "queued", 0)
	finished, abandoned := run(org, "succeeded", time.Minute), run(org, "running", 76*time.Minute)
	noKey := run(keyless, "running", time.Minute)

	logged := captureLogger(t)
	api := &CredentialsAPI{DB: db, Keys: modelkeys.Store{DB: db}, Atespace: "ate-env", RunLifetime: 75 * time.Minute, Token: "provider-token"}
	handler := api.Handler()
	fetch := func(token string, run uuid.UUID, provider, actor string) (int, string) {
		t.Helper()
		body := `{"run":"` + run.String() + `","provider":"` + provider + `","actor":"` + actor + `"}`
		req := httptest.NewRequest(http.MethodPost, "/api/credentials/fetch", strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}
	actor := func(run uuid.UUID) string { return sandbox.SubstrateActor("ate-env", run) }

	status, body := fetch("provider-token", running, "anthropic", actor(running))
	require.Equal(t, http.StatusOK, status, body)
	assert.JSONEq(t, `{"value":"`+modelKeysTestAPIKey+`"}`, body)
	status, body = fetch("provider-token", queued, "anthropic", actor(queued))
	assert.Equal(t, http.StatusOK, status, "a job can call the model before its run is recorded as running: %s", body)

	for name, test := range map[string]struct {
		token, provider, actor string
		run                    uuid.UUID
		status                 int
		want                   string
	}{
		"no token":        {"", "anthropic", actor(running), running, http.StatusUnauthorized, "send the credential provider's token as Authorization: Bearer"},
		"wrong token":     {"other-token", "anthropic", actor(running), running, http.StatusUnauthorized, "not the credential provider's token"},
		"unknown run":     {"provider-token", "anthropic", actor(uuid.Nil), uuid.Nil, http.StatusNotFound, "sandbox run not found"},
		"finished run":    {"provider-token", "anthropic", actor(finished), finished, http.StatusForbidden, "run " + finished.String() + " is succeeded, not active"},
		"abandoned run":   {"provider-token", "anthropic", actor(abandoned), abandoned, http.StatusForbidden, "run " + abandoned.String() + " started over 1h15m0s ago"},
		"another actor":   {"provider-token", "anthropic", actor(queued), running, http.StatusForbidden, "isn't run " + running.String() + "'s environment"},
		"other atespace":  {"provider-token", "anthropic", sandbox.SubstrateActor("other", running), running, http.StatusForbidden, "isn't run " + running.String() + "'s environment"},
		"no org key":      {"provider-token", "anthropic", actor(noKey), noKey, http.StatusNotFound, "has no anthropic key"},
		"no provider key": {"provider-token", "openai", actor(running), running, http.StatusNotFound, "has no openai key"},
	} {
		status, body := fetch(test.token, test.run, test.provider, test.actor)
		assert.Equal(t, test.status, status, name)
		assert.Contains(t, body, test.want, name)
		assert.NotContains(t, body, "route-test-key", name)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/credentials/fetch", strings.NewReader(`{`))
	req.Header.Set("Authorization", "Bearer provider-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid JSON")

	assert.Contains(t, logged.String(), "credential fetch for run "+running.String()+": organization "+org.String()+"'s anthropic key")
	assert.Contains(t, logged.String(), "credential fetch for run "+finished.String()+"'s anthropic key refused: run "+finished.String()+" is succeeded, not active")
	assert.NotContains(t, logged.String(), "route-test-key", "the key is never logged")
}

// captureLogger sends the global logger to a buffer for the rest of t, so the lines a test
// expects are asserted instead of printed.
func captureLogger(t *testing.T) *strings.Builder {
	t.Helper()
	var buf strings.Builder
	previous := logger.StandardLogger()
	logger.Use(&buf)
	t.Cleanup(func() { logger.SetLogger(previous) })
	return &buf
}
