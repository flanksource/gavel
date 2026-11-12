// Tests reading gavel's infrastructure from Mission Control against a fake of its resource
// search API: the searches sent, the resources returned, and each list's failure explained.
package missioncontrol_test

import (
	"strings"
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
		Configs: []clientapi.SelectedResource{
			{ID: "c1", Name: "gavel", Namespace: "gavel", Type: "Kubernetes::Deployment", Health: "healthy", Status: "Running"},
			{ID: "c2", Name: "gavel-postgres", Namespace: "gavel", Type: "Kubernetes::StatefulSet", Health: "healthy"},
		},
		Checks: []clientapi.SelectedResource{{ID: "k1", Name: "gavel-http", Health: "unhealthy"}},
	})
	mc.SetChanges("c1", clientapi.SelectedResource{ID: "ch1", Name: "gavel", Type: "diff"})
	mc.SetChanges("c9", clientapi.SelectedResource{ID: "ch9", Name: "not-gavel", Type: "diff"})

	got, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
	require.NoError(t, err)
	searches := mc.Searches()
	require.Len(t, searches, 3, "config items, health checks and changes are read one at a time")
	selector := []clientapi.ResourceSelector{{Search: "namespace=gavel"}}
	assert.Equal(t, clientapi.SearchResourcesRequest{Limit: 200, Configs: selector}, searches[0], "the configured search selects gavel's config items")
	assert.Equal(t, clientapi.SearchResourcesRequest{Limit: 200, Checks: selector}, searches[1], "and its health checks")
	assert.Equal(t, clientapi.SearchResourcesRequest{Limit: 200, ConfigChanges: []clientapi.ResourceSelector{{Search: "config_id=c1,c2"}}}, searches[2],
		"changes are those of the config items found, since Mission Control can't search changes by namespace")
	assert.Equal(t, mc.URL, got.MissionControl)
	require.Len(t, got.Configs, 2)
	assert.Equal(t, "Kubernetes::Deployment", got.Configs[0].Type)
	require.Len(t, got.Checks, 1)
	assert.Equal(t, "unhealthy", got.Checks[0].Health)
	require.Len(t, got.Changes, 1)
	assert.Equal(t, "ch1", got.Changes[0].ID)
	assert.Empty(t, got.ConfigsError+got.ChecksError+got.ChangesError)
}

func TestInfrastructureWithoutConfigItemsHasNoChanges(t *testing.T) {
	mc := missioncontroltest.New(t)

	got, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
	require.NoError(t, err)
	assert.Len(t, mc.Searches(), 2, "with no config items there are no changes to read")
	assert.NotNil(t, got.Configs, "no config items is an empty list, not null")
	assert.NotNil(t, got.Checks)
	assert.NotNil(t, got.Changes)
}

func TestInfrastructureSaysWhyOneListIsMissing(t *testing.T) {
	configs := []clientapi.SelectedResource{{ID: "c1", Name: "gavel"}}
	checks := []clientapi.SelectedResource{{ID: "k1", Name: "gavel-http"}}

	t.Run("health checks", func(t *testing.T) {
		mc := missioncontroltest.New(t)
		mc.SetResources(clientapi.SearchResourcesResponse{Configs: configs, Checks: checks})
		mc.SetChanges("c1", clientapi.SelectedResource{ID: "ch1"})
		mc.Refuse("checks", "canaries are down")

		got, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
		require.NoError(t, err)
		assert.Contains(t, got.ChecksError, "read gavel's health checks from Mission Control at "+mc.URL)
		assert.Contains(t, got.ChecksError, "canaries are down")
		assert.NotNil(t, got.Checks)
		assert.Len(t, got.Configs, 1, "config items still load")
		assert.Len(t, got.Changes, 1, "and their changes")
	})

	t.Run("changes", func(t *testing.T) {
		mc := missioncontroltest.New(t)
		mc.SetResources(clientapi.SearchResourcesResponse{Configs: configs, Checks: checks})
		mc.Refuse("changes", "changes are down")

		got, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
		require.NoError(t, err)
		assert.Contains(t, got.ChangesError, "read the changes of gavel's config items from Mission Control at "+mc.URL)
		assert.Contains(t, got.ChangesError, "changes are down")
		assert.Len(t, got.Configs, 1)
		assert.Len(t, got.Checks, 1)
	})

	t.Run("config items", func(t *testing.T) {
		mc := missioncontroltest.New(t)
		mc.SetResources(clientapi.SearchResourcesResponse{Configs: configs, Checks: checks})
		mc.Refuse("configs", "the catalog is down")

		got, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
		require.NoError(t, err)
		assert.Contains(t, got.ConfigsError, "read gavel's config items from Mission Control at "+mc.URL)
		assert.Contains(t, got.ConfigsError, "the catalog is down")
		assert.Equal(t, "gavel's changes are those of its config items, which couldn't be read", got.ChangesError)
		assert.Len(t, mc.Searches(), 2, "no changes are read without config items")
		assert.Len(t, got.Checks, 1, "health checks still load")
	})
}

func TestInfrastructureExplainsFailures(t *testing.T) {
	mc := missioncontroltest.New(t)

	_, err := missioncontrol.New(mc.URL, "wrong-token", "namespace=gavel").Infrastructure(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read gavel's infrastructure from Mission Control at "+mc.URL)
	assert.Contains(t, err.Error(), "401")
	assert.Equal(t, 1, strings.Count(err.Error(), "check GAVEL_MISSION_CONTROL_TOKEN"), "every list failing for one reason says it once: %v", err)
	assert.NotContains(t, err.Error(), "wrong-token")

	_, err = missioncontrol.New("http://127.0.0.1:1", missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Mission Control at http://127.0.0.1:1")
}
