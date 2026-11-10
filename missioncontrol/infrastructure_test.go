// Tests reading gavel's infrastructure from Mission Control against a fake of its resource
// search API: the selector sent, the resources returned, and failures explained.
package missioncontrol_test

import (
	"testing"

	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/gavel/missioncontrol/missioncontroltest"
	"github.com/flanksource/incident-commander/clientapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfrastructure(t *testing.T) {
	mc := missioncontroltest.New(t)
	mc.SetResources(clientapi.SearchResourcesResponse{
		Configs: []clientapi.SelectedResource{{ID: "c1", Name: "gavel", Namespace: "gavel", Type: "Kubernetes::Deployment", Health: "healthy", Status: "Running"}},
		Checks:  []clientapi.SelectedResource{{ID: "k1", Name: "gavel-http", Health: "unhealthy"}},
	})

	got, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
	require.NoError(t, err)
	require.Len(t, mc.Searches(), 1)
	sent := mc.Searches()[0]
	selector := []clientapi.ResourceSelector{{Search: "namespace=gavel"}}
	assert.Equal(t, selector, sent.Configs, "the configured search selects gavel's configs")
	assert.Equal(t, selector, sent.Checks)
	assert.Equal(t, selector, sent.ConfigChanges)
	assert.Equal(t, mc.URL, got.MissionControl)
	require.Len(t, got.Configs, 1)
	assert.Equal(t, "Kubernetes::Deployment", got.Configs[0].Type)
	require.Len(t, got.Checks, 1)
	assert.Equal(t, "unhealthy", got.Checks[0].Health)
	assert.NotNil(t, got.Changes, "no changes is an empty list, not null")
}

func TestInfrastructureExplainsFailures(t *testing.T) {
	mc := missioncontroltest.New(t)

	_, err := missioncontrol.New(mc.URL, "wrong-token", "namespace=gavel").Infrastructure(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read gavel's infrastructure from Mission Control at "+mc.URL)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "check GAVEL_MISSION_CONTROL_TOKEN")
	assert.NotContains(t, err.Error(), "wrong-token")

	_, err = missioncontrol.New("http://127.0.0.1:1", missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Mission Control at http://127.0.0.1:1")
}
