// Reads gavel's infrastructure from a real Mission Control, when one is named: set
// GAVEL_MISSION_CONTROL_TEST=1 with GAVEL_MISSION_CONTROL_URL, _TOKEN and _INFRA.
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
