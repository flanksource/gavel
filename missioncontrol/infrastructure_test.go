// Tests reading gavel's infrastructure from Mission Control against a fake of its resource and
// changes search APIs: the searches sent, the resources returned, and each list's failure explained.
package missioncontrol_test

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/gavel/missioncontrol/missioncontroltest"
	"github.com/flanksource/mission-control-sdk/go/clientapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfrastructure(t *testing.T) {
	updated := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	changed := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	mc := missioncontroltest.New(t)
	mc.SetResources(clientapi.SearchResourcesResponse{
		Configs: []clientapi.SelectedResource{
			{ID: "c1", Name: "gavel", Namespace: "gavel", Type: "Kubernetes::Deployment", Health: "healthy", Status: "Running", UpdatedAt: &updated},
			{ID: "c2", Name: "gavel-postgres", Namespace: "gavel", Type: "Kubernetes::StatefulSet", Health: "healthy"},
		},
		Checks: []clientapi.SelectedResource{{ID: "k1", Name: "gavel-http", Health: "unhealthy", UpdatedAt: &updated}},
	})
	mc.SetChanges("c1", clientapi.ConfigChangeRow{ID: "ch1", ConfigID: "c1", ConfigName: "gavel", ChangeType: "diff", Summary: "status.replicas", Severity: "info", CreatedAt: &changed})
	mc.SetChanges("c9", clientapi.ConfigChangeRow{ID: "ch9", ConfigID: "c9", ConfigName: "not-gavel", ChangeType: "diff"})

	got, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
	require.NoError(t, err)
	searches := mc.Searches()
	require.Len(t, searches, 2, "config items and health checks are read one at a time")
	selector := []clientapi.ResourceSelector{{Search: "namespace=gavel"}}
	assert.Equal(t, clientapi.SearchResourcesRequest{Limit: 200, Timestamps: true, Configs: selector}, searches[0],
		"the configured search selects gavel's config items, with when each was updated")
	assert.Equal(t, clientapi.SearchResourcesRequest{Limit: 200, Timestamps: true, Checks: selector}, searches[1], "and its health checks")
	assert.Equal(t, []clientapi.CatalogChangesSearchRequest{{BaseCatalogSearch: clientapi.BaseCatalogSearch{
		CatalogID: "c1,c2", From: "now-7d", PageSize: 200, SortBy: "-created_at", Recursive: clientapi.CatalogChangeRecursiveNone,
	}}}, mc.ChangeSearches(), "changes are the newest of the config items found, not of their children, since Mission Control can't search changes by namespace")
	assert.Equal(t, mc.URL, got.MissionControl)
	require.Len(t, got.Configs, 2)
	assert.Equal(t, "Kubernetes::Deployment", got.Configs[0].Type)
	assert.Equal(t, &updated, got.Configs[0].UpdatedAt)
	require.Len(t, got.Checks, 1)
	assert.Equal(t, "unhealthy", got.Checks[0].Health)
	assert.Equal(t, &updated, got.Checks[0].UpdatedAt)
	assert.Equal(t, []clientapi.ConfigChangeRow{{ID: "ch1", ConfigID: "c1", ConfigName: "gavel", ChangeType: "diff", Summary: "status.replicas", Severity: "info", CreatedAt: &changed}}, got.Changes,
		"each change names its config item and says when it happened")
	assert.Equal(t, 7, got.ChangesDays)
	assert.Empty(t, got.ConfigsError+got.ChecksError+got.ChangesError)
}

func TestInfrastructureWithoutConfigItemsHasNoChanges(t *testing.T) {
	mc := missioncontroltest.New(t)

	got, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel").Infrastructure(t.Context())
	require.NoError(t, err)
	assert.Empty(t, mc.ChangeSearches(), "with no config items there are no changes to read")
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
		mc.SetChanges("c1", clientapi.ConfigChangeRow{ID: "ch1"})
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
		assert.Empty(t, mc.ChangeSearches(), "no changes are read without config items")
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
