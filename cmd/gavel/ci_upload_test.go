// Tests gavel ci upload: the credential of each CI a job runs in (GitHub Actions' OIDC token, GitLab
// CI's ID token, Azure Pipelines' access token with its build), and uploading results with it.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/commons/logger"
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

	credential, err := requestCICredential(context.Background(), http.DefaultClient, getenv, "SERVER")
	require.NoError(t, err)
	assert.Equal(t, ciCredential{token: "oidc-token"}, credential, "a GitHub Actions job's OIDC token")

	data, err := os.ReadFile(file)
	require.NoError(t, err)
	runID, err := sendCIResults(context.Background(), http.DefaultClient, gavel.URL, credential, "", data)
	require.NoError(t, err)
	assert.Equal(t, "run-2026-10-15T09-00-00Z", runID)
	assert.JSONEq(t, `{"tests":[]}`, string(uploaded))

	_, err = sendCIResults(context.Background(), http.DefaultClient, gavel.URL, ciCredential{token: "wrong"}, "", data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 403: rejected")
}

func TestTheCredentialOfEachCI(t *testing.T) {
	for name, tc := range map[string]struct {
		env  map[string]string
		want ciCredential
		err  string
	}{
		"GitLab CI's ID token": {env: map[string]string{"GITLAB_CI": "true", "GAVEL_ID_TOKEN": "gitlab-token"}, want: ciCredential{token: "gitlab-token"}},
		"GitLab CI without its ID token": {env: map[string]string{"GITLAB_CI": "true"},
			err: "GITLAB_CI is set but GAVEL_ID_TOKEN is not: declare it in the job's id_tokens with aud: https://gavel.example"},
		"Azure Pipelines' access token for its build": {env: map[string]string{"TF_BUILD": "True", "SYSTEM_ACCESSTOKEN": "job-token", "BUILD_BUILDID": "91", "BUILD_REPOSITORY_ID": "rules-guid"},
			want: ciCredential{token: "job-token", headers: map[string]string{"X-Gavel-Azure-Build": "91", "X-Gavel-Azure-Repository": "rules-guid"}}},
		"Azure Pipelines without its access token": {env: map[string]string{"TF_BUILD": "True", "BUILD_BUILDID": "91", "BUILD_REPOSITORY_ID": "rules-guid"},
			err: "TF_BUILD is set but SYSTEM_ACCESSTOKEN is not: map it into the step's env as SYSTEM_ACCESSTOKEN: $(System.AccessToken)"},
		"Azure Pipelines without its build": {env: map[string]string{"TF_BUILD": "True", "SYSTEM_ACCESSTOKEN": "job-token"},
			err: "TF_BUILD is set but BUILD_BUILDID or BUILD_REPOSITORY_ID is not: run in an Azure Pipelines build of an Azure Repos repository"},
		"no CI": {env: map[string]string{},
			err: "ACTIONS_ID_TOKEN_REQUEST_URL and ACTIONS_ID_TOKEN_REQUEST_TOKEN are not set: run in GitHub Actions with permissions: id-token: write"},
	} {
		t.Run(name, func(t *testing.T) {
			getenv := func(key string) string { return tc.env[key] }
			credential, err := requestCICredential(context.Background(), http.DefaultClient, getenv, "https://gavel.example")
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				require.EqualError(t, ciCredentialMissing(getenv, "https://gavel.example"), tc.err, "checked before any request")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, credential)
			require.NoError(t, ciCredentialMissing(getenv, "https://gavel.example"))
		})
	}
}

func TestAnAzureJobSendsItsBuildWithItsResults(t *testing.T) {
	var headers http.Header
	gavel := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"runId": "run-2026-11-22T09-00-00Z"})
	}))
	t.Cleanup(gavel.Close)
	credential := ciCredential{token: "job-token", headers: map[string]string{"X-Gavel-Azure-Build": "91", "X-Gavel-Azure-Repository": "rules-guid"}}
	_, err := sendCIResults(context.Background(), http.DefaultClient, gavel.URL, credential, "run-2026-11-22T09-00-00Z", []byte(`{"tests":[]}`))
	require.NoError(t, err)
	assert.Equal(t, "Bearer job-token", headers.Get("Authorization"))
	assert.Equal(t, "91", headers.Get("X-Gavel-Azure-Build"))
	assert.Equal(t, "rules-guid", headers.Get("X-Gavel-Azure-Repository"))
}

func TestResultsStreamFromAGitLabJobWithItsIDToken(t *testing.T) {
	env := map[string]string{"GAVEL_CI_SERVER": "https://gavel.example", "GITLAB_CI": "true", "GAVEL_ID_TOKEN": "gitlab-token"}
	stream := startCIStream(func(key string) string { return env[key] })
	require.NotNil(t, stream)
	stream.Stop()
	delete(env, "GAVEL_ID_TOKEN")
	var warned bytes.Buffer
	previous := logger.GetOutput()
	logger.SetOutput(&warned)
	t.Cleanup(func() { logger.SetOutput(previous) })
	assert.Nil(t, startCIStream(func(key string) string { return env[key] }), "a GitLab job without its ID token streams nothing")
	assert.Contains(t, warned.String(), "GAVEL_CI_SERVER is set, but the job can't stream results: GITLAB_CI is set but GAVEL_ID_TOKEN is not")
}
