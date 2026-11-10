// Tests reading gavel's infrastructure from Mission Control against a server that speaks
// its resource search API: the selector sent, the resources returned, and failures explained.
package missioncontrol_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/incident-commander/clientapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// searchServer answers POST /resources/search for the bearer token "mc-token" with
// response, recording the request it was sent.
func searchServer(t *testing.T, response clientapi.SearchResourcesResponse, got *clientapi.SearchResourcesRequest) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/resources/search" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer mc-token" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid access token"})
			return
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(got))
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func TestInfrastructure(t *testing.T) {
	var sent clientapi.SearchResourcesRequest
	url := searchServer(t, clientapi.SearchResourcesResponse{
		Configs: []clientapi.SelectedResource{{ID: "c1", Name: "gavel", Namespace: "gavel", Type: "Kubernetes::Deployment", Health: "healthy", Status: "Running"}},
		Checks:  []clientapi.SelectedResource{{ID: "k1", Name: "gavel-http", Health: "unhealthy"}},
	}, &sent)

	got, err := missioncontrol.New(url, "mc-token", "namespace=gavel").Infrastructure(t.Context())
	require.NoError(t, err)
	selector := []clientapi.ResourceSelector{{Search: "namespace=gavel"}}
	assert.Equal(t, selector, sent.Configs, "the configured search selects gavel's configs")
	assert.Equal(t, selector, sent.Checks)
	assert.Equal(t, selector, sent.ConfigChanges)
	assert.Equal(t, url, got.MissionControl)
	require.Len(t, got.Configs, 1)
	assert.Equal(t, "Kubernetes::Deployment", got.Configs[0].Type)
	require.Len(t, got.Checks, 1)
	assert.Equal(t, "unhealthy", got.Checks[0].Health)
	assert.NotNil(t, got.Changes, "no changes is an empty list, not null")
}

func TestInfrastructureExplainsFailures(t *testing.T) {
	url := searchServer(t, clientapi.SearchResourcesResponse{}, &clientapi.SearchResourcesRequest{})

	_, err := missioncontrol.New(url, "wrong-token", "namespace=gavel").Infrastructure(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read gavel's infrastructure from Mission Control at "+url)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "check GAVEL_MISSION_CONTROL_TOKEN")
	assert.NotContains(t, err.Error(), "wrong-token")

	_, err = missioncontrol.New("http://127.0.0.1:1", "mc-token", "namespace=gavel").Infrastructure(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Mission Control at http://127.0.0.1:1")
}
