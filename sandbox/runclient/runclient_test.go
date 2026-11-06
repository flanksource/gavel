// Tests the run API client a sandbox job calls serve with: it finds its run in the job's
// environment, sends the run token, and reports serve's refusals.
package runclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
	assert.EqualError(t, err, "not in a sandbox job: GAVEL_RUN_ID, and GAVEL_RUN_API_SOCKET or GAVEL_SERVER and GAVEL_RUN_TOKEN, must be set")
}

// shortSocket is a socket path in a directory of its own under /tmp: a socket's path must fit in
// 108 bytes, which a test's temporary directory may not leave room for.
func shortSocket(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "run-api-")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "run-api.sock")
}

// The environment's program serves the run API on a socket to what runs under it, forwarding
// each request to the run's caller over the host call, which answers it as the run.
func TestFromEnvReachesTheRunAPIThroughTheEnvironmentsSocket(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("GET /api/runs/run-1/spec", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer run-token", r.Header.Get("Authorization"), "the caller answers as the run")
		_, _ = io.WriteString(w, `{"kind":"test"}`)
	})
	handler.HandleFunc("POST /api/runs/run-1/results", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"tests":[]}`, string(body))
		http.Error(w, `{"error":"the run already finished"}`, http.StatusConflict)
	})
	answer := AnswerHostCall(handler, "run-1", "run-token", JobReach)
	socket := shortSocket(t)
	stop, err := ServeOverHostCall(socket, func(name string, data json.RawMessage) (json.RawMessage, error) {
		assert.Equal(t, HostCallName, name)
		return answer(context.Background(), data)
	})
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, stop()) })

	env := map[string]string{"GAVEL_RUN_API_SOCKET": socket, "GAVEL_RUN_ID": "run-1"}
	client, err := FromEnv(func(key string) string { return env[key] })
	require.NoError(t, err)
	assert.Equal(t, "run-1", client.Run)
	assert.Empty(t, client.Token, "the environment holds no run token")

	resp, err := client.Do(context.Background(), http.MethodGet, "spec", nil)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.JSONEq(t, `{"kind":"test"}`, string(body))

	_, err = client.Do(context.Background(), http.MethodPost, "results", strings.NewReader(`{"tests":[]}`))
	require.EqualError(t, err, "HTTP 409: the run already finished", "a refusal reads as serve's over HTTP does")

	_, err = client.Do(context.Background(), http.MethodGet, "source", nil)
	require.EqualError(t, err, `HTTP 502: a run's program reaches only its spec and results through its caller, not "source"`)
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
