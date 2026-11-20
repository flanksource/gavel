// Reads gavel's infrastructure and substrate's metrics View from a real Mission Control, when one is
// named: set GAVEL_MISSION_CONTROL_TEST=1 with GAVEL_MISSION_CONTROL_URL, _TOKEN, _INFRA and _SUBSTRATE_VIEW.
package missioncontrol_test

import (
	"os"
	"testing"

	"github.com/flanksource/gavel/missioncontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfrastructureFromARealMissionControl(t *testing.T) {
	if os.Getenv("GAVEL_MISSION_CONTROL_TEST") != "1" {
		t.Skip("set GAVEL_MISSION_CONTROL_TEST=1 with GAVEL_MISSION_CONTROL_URL, GAVEL_MISSION_CONTROL_TOKEN and GAVEL_MISSION_CONTROL_INFRA to read a real Mission Control")
	}
	url, token, search := os.Getenv("GAVEL_MISSION_CONTROL_URL"), os.Getenv("GAVEL_MISSION_CONTROL_TOKEN"), os.Getenv("GAVEL_MISSION_CONTROL_INFRA")
	require.NotEmpty(t, url, "GAVEL_MISSION_CONTROL_URL names the Mission Control backend")
	require.NotEmpty(t, token, "GAVEL_MISSION_CONTROL_TOKEN is an access token with the viewer role")
	require.NotEmpty(t, search, "GAVEL_MISSION_CONTROL_INFRA is the search selecting gavel's infrastructure")

	got, err := missioncontrol.New(url, token, search).Infrastructure(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, got.Configs, "the search %q selects no config items in %s", search, url)
}

func TestViewFromARealMissionControl(t *testing.T) {
	if os.Getenv("GAVEL_MISSION_CONTROL_TEST") != "1" {
		t.Skip("set GAVEL_MISSION_CONTROL_TEST=1 with GAVEL_MISSION_CONTROL_URL, GAVEL_MISSION_CONTROL_TOKEN and GAVEL_MISSION_CONTROL_SUBSTRATE_VIEW to read a real Mission Control")
	}
	url, token, view := os.Getenv("GAVEL_MISSION_CONTROL_URL"), os.Getenv("GAVEL_MISSION_CONTROL_TOKEN"), os.Getenv("GAVEL_MISSION_CONTROL_SUBSTRATE_VIEW")
	require.NotEmpty(t, url, "GAVEL_MISSION_CONTROL_URL names the Mission Control backend")
	require.NotEmpty(t, token, "GAVEL_MISSION_CONTROL_TOKEN is an access token with the viewer role")
	require.NotEmpty(t, view, "GAVEL_MISSION_CONTROL_SUBSTRATE_VIEW is the namespace/name of substrate's metrics View")

	got, err := missioncontrol.New(url, token, "").View(t.Context(), view, nil)
	require.NoError(t, err)
	assert.Empty(t, got.RefreshError, "the view's queries ran")
	var drawn int
	for _, panel := range got.Panels {
		if panel.Type == "timeseries" && panel.Unsupported == "" && len(panel.Series) > 0 {
			drawn++
		}
	}
	assert.Positive(t, drawn, "the view %s has a timeseries panel with data", view)
}
