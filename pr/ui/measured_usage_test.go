// Tests an organization's measured CPU and memory, read from a Mission Control View given the
// organization's ID: its admins see their own organization's, platform admins any organization's.
package ui

import (
	"net/http"
	"testing"

	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/gavel/missioncontrol/missioncontroltest"
	"github.com/flanksource/gavel/platform"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// measuredUsageView is a Mission Control serving the organizations' measured use View.
func measuredUsageView(t *testing.T) (*missioncontroltest.Server, *missioncontrol.Source) {
	t.Helper()
	mc := missioncontroltest.New(t)
	mc.SetView("substrate", "org-usage", map[string]any{
		"title":  "Measured use",
		"panels": []map[string]any{{"name": "Running environments", "type": "number", "rows": []map[string]any{{"value": 2}}}},
	})
	return mc, missioncontrol.New(mc.URL, missioncontroltest.Token, "namespace=gavel")
}

func TestOrganizationAdminsSeeTheirOrganizationsMeasuredUse(t *testing.T) {
	db, _, org := openOrgDB(t)
	mc, source := measuredUsageView(t)
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetMeasuredUsage(MeasuredUsage{MissionControl: source, View: "substrate/org-usage"})
	base := serveActingAs(t, srv, org)

	status, view := hostedCall(t, base, http.MethodGet, "/api/usage/measured", uuid.New(), "org:admin", nil)
	require.Equal(t, http.StatusOK, status, view)
	assert.Equal(t, "Measured use", view["title"])
	assert.Equal(t, []map[string]any{{"variables": map[string]any{"org": org.String()}}}, mc.ViewRequests(), "the View is read for this organization")

	status, refused := hostedCall(t, base, http.MethodGet, "/api/usage/measured", uuid.New(), "org:member", nil)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Equal(t, "only an organization admin sees its measured CPU and memory", refused["error"])
	assert.Len(t, mc.ViewRequests(), 1, "a member's request doesn't reach Mission Control")
}

func TestMeasuredUseSaysWhyItIsMissing(t *testing.T) {
	db, _, org := openOrgDB(t)
	_, source := measuredUsageView(t)
	for _, test := range []struct {
		name     string
		measured MeasuredUsage
		error    string
	}{
		{"without Mission Control", MeasuredUsage{}, errNoMissionControl.Error()},
		{"without the View", MeasuredUsage{MissionControl: source}, "no View measures an organization's CPU and memory: start gavel serve with --mission-control-org-view namespace/name"},
	} {
		t.Run(test.name, func(t *testing.T) {
			srv := &Server{}
			srv.SetOrg(org, db)
			srv.SetHostedMode(true)
			srv.SetMeasuredUsage(test.measured)
			status, failure := hostedCall(t, serveActingAs(t, srv, org), http.MethodGet, "/api/usage/measured", uuid.New(), "org:admin", nil)
			assert.Equal(t, http.StatusNotFound, status)
			assert.Equal(t, test.error, failure["error"])
		})
	}
}

func TestPlatformAdminsSeeAnyOrganizationsMeasuredUse(t *testing.T) {
	db, _, org := openOrgDB(t)
	mc, source := measuredUsageView(t)
	other := uuid.New()
	base := servePlatform(t, &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}, MissionControl: source, OrgView: "substrate/org-usage"}, org, true)

	var view missioncontrol.View
	require.Equal(t, http.StatusOK, platformGet(t, base+"/api/platform/orgs/"+other.String()+"/measured-usage", &view))
	assert.Equal(t, "Measured use", view.Title)
	assert.Equal(t, []map[string]any{{"variables": map[string]any{"org": other.String()}}}, mc.ViewRequests(), "the View is read for the organization asked about")

	var failure map[string]string
	assert.Equal(t, http.StatusBadRequest, platformGet(t, base+"/api/platform/orgs/acme/measured-usage", &failure))
	assert.Contains(t, failure["error"], `organization ID "acme"`)
}
