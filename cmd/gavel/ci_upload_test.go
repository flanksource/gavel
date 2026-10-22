// Tests gavel ci upload: composing the Actions OIDC token request, and uploading results to a
// gavel server with the token it returns.
package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCTokenRequestURL(t *testing.T) {
	for requestURL, want := range map[string]string{
		"https://pipelines.actions.githubusercontent.com/abc/idtoken?api-version=2.0": "https://pipelines.actions.githubusercontent.com/abc/idtoken?api-version=2.0&audience=https%3A%2F%2Fgavel.example",
		"https://example.test/token":                "https://example.test/token?audience=https%3A%2F%2Fgavel.example",
		"https://example.test/token?audience=other": "https://example.test/token?audience=https%3A%2F%2Fgavel.example",
	} {
		got, err := oidcTokenRequestURL(requestURL, "https://gavel.example")
		require.NoError(t, err, requestURL)
		assert.Equal(t, want, got, requestURL)
	}
	_, err := oidcTokenRequestURL("://bad", "https://gavel.example")
	require.Error(t, err)
}

func TestCIUploadRequiresTheActionsTokenVariables(t *testing.T) {
	_, err := requestActionsOIDCToken(context.Background(), http.DefaultClient, func(string) string { return "" }, "https://gavel.example")
	require.EqualError(t, err, "ACTIONS_ID_TOKEN_REQUEST_URL and ACTIONS_ID_TOKEN_REQUEST_TOKEN are not set: run in GitHub Actions with permissions: id-token: write")
}

func TestCIUpload(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "bearer request-token" || r.URL.Query().Get("audience") != "SERVER" {
			http.Error(w, "bad token request", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"value": "oidc-token"})
	}))
	t.Cleanup(tokenServer.Close)
	var uploaded []byte
	gavel := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/ci/results" || r.Header.Get("Authorization") != "Bearer oidc-token" {
			http.Error(w, `{"error":"rejected"}`, http.StatusForbidden)
			return
		}
		uploaded, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"runId": "run-2026-10-15T09-00-00Z"})
	}))
	t.Cleanup(gavel.Close)
	file := filepath.Join(t.TempDir(), "results.json")
	require.NoError(t, os.WriteFile(file, []byte(`{"tests":[]}`), 0o644))
	env := map[string]string{
		"ACTIONS_ID_TOKEN_REQUEST_URL":   tokenServer.URL + "/token?api-version=2.0",
		"ACTIONS_ID_TOKEN_REQUEST_TOKEN": "request-token",
	}
	getenv := func(key string) string { return env[key] }

	// The audience is the server URL; this fake token server expects the literal SERVER.
	token, err := requestActionsOIDCToken(context.Background(), http.DefaultClient, getenv, "SERVER")
	require.NoError(t, err)
	assert.Equal(t, "oidc-token", token)

	runID, err := uploadCIResults(context.Background(), http.DefaultClient, gavel.URL, token, file)
	require.NoError(t, err)
	assert.Equal(t, "run-2026-10-15T09-00-00Z", runID)
	assert.JSONEq(t, `{"tests":[]}`, string(uploaded))

	_, err = uploadCIResults(context.Background(), http.DefaultClient, gavel.URL, "wrong", file)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 403: rejected")
}
