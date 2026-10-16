// Tests the run API client a sandbox job calls serve with: it finds its run in the job's
// environment, sends the run token, and reports serve's refusals.
package runclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromEnvFindsTheJobsRun(t *testing.T) {
	env := map[string]string{"GAVEL_SERVER": "https://gavel.example/", "GAVEL_RUN_TOKEN": "token", "GAVEL_RUN_ID": "run-1"}
	client, err := FromEnv(func(key string) string { return env[key] })
	require.NoError(t, err)
	assert.Equal(t, "https://gavel.example", client.Server)
	assert.Equal(t, "run-1", client.Run)

	delete(env, "GAVEL_RUN_TOKEN")
	_, err = FromEnv(func(key string) string { return env[key] })
	require.ErrorIs(t, err, ErrNotInAJob)
	assert.EqualError(t, err, "not in a sandbox job: GAVEL_SERVER, GAVEL_RUN_TOKEN and GAVEL_RUN_ID must be set")
}

func TestDoSendsTheRunTokenAndReportsRefusals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			http.Error(w, `{"error":"bad token"}`, http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/api/runs/run-1/missing" {
			http.Error(w, `{"error":"no such thing"}`, http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, r.Method+" "+r.URL.Path)
	}))
	t.Cleanup(server.Close)
	client := &Client{Server: server.URL, Token: "token", Run: "run-1", HTTP: server.Client()}

	resp, err := client.Do(context.Background(), http.MethodGet, "spec", nil)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, "GET /api/runs/run-1/spec", string(body))

	_, err = client.Do(context.Background(), http.MethodGet, "missing", nil)
	require.EqualError(t, err, "HTTP 404: no such thing")

	_, err = (&Client{Server: server.URL, Token: "wrong", Run: "run-1", HTTP: server.Client()}).Do(context.Background(), http.MethodGet, "spec", nil)
	require.EqualError(t, err, "HTTP 401: bad token")
}
