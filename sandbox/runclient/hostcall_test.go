// Tests the run API over a host call: a run's program in an environment calls its caller, which
// answers by serving the request through a run API handler, as serve's answers it over HTTP.
package runclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runAPI serves one run's PR status, and refuses a request without the run's token.
func runAPI(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/runs/run-1/pr/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer run-token" {
			http.Error(w, `{"error":"bad token"}`, http.StatusUnauthorized)
			return
		}
		assert.Equal(t, "7", r.URL.Query().Get("number"))
		_, _ = w.Write([]byte(`{"state":"OPEN"}`))
	})
	mux.HandleFunc("POST /api/runs/run-1/pr/push", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "the bundle", string(body))
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"the PR's source branch is at abc, not def"}`))
	})
	return mux
}

func TestTheRunAPIAnswersOverAHostCall(t *testing.T) {
	answer := AnswerHostCall(runAPI(t), "run-1", "run-token")
	client := OverHostCall(func(name string, data json.RawMessage) (json.RawMessage, error) {
		assert.Equal(t, HostCallName, name)
		return answer(context.Background(), data)
	})

	resp, err := client.Do(context.Background(), http.MethodGet, "pr/status?number=7", nil)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"state":"OPEN"}`, string(body))

	_, err = client.Do(context.Background(), http.MethodPost, "pr/push?base=def", strings.NewReader("the bundle"))
	require.EqualError(t, err, "HTTP 409: the PR's source branch is at abc, not def", "a refusal reads as serve's over HTTP does")
}

func TestAHostCallReachesOnlyTheRunsPREndpoints(t *testing.T) {
	answer := AnswerHostCall(runAPI(t), "run-1", "run-token")
	for _, path := range []string{"source", "pr/../source", "../run-2/pr/status", "/api/credentials/fetch"} {
		data, err := json.Marshal(hostRequest{Method: http.MethodGet, Path: path})
		require.NoError(t, err)
		_, err = answer(context.Background(), data)
		assert.EqualError(t, err, `a run's program reaches only its PR's endpoints through its caller, not "`+path+`"`, path)
	}
}
