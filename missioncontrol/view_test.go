// Tests reading a Mission Control View's panels against a fake of its view API: timeseries rows
// grouped into series, single values, panels gavel can't draw, and failures explained.
package missioncontrol_test

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/gavel/missioncontrol/missioncontroltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewGroupsTimeseriesRowsIntoSeries(t *testing.T) {
	mc := missioncontroltest.New(t)
	mc.SetView("substrate", "capacity", map[string]any{
		"title": "Substrate", "lastRefreshedAt": "2026-11-10T10:05:00Z",
		"panels": []map[string]any{
			{
				"name": "Ready workers", "description": "Last 12h", "type": "timeseries",
				"timeseries": map[string]any{"timeKey": "timestamp", "valueKey": "value"},
				"rows": []map[string]any{
					{"timestamp": "2026-11-10T10:01:00Z", "workerpool": "pool-b", "value": 2},
					{"timestamp": "2026-11-10T10:01:00Z", "workerpool": "pool-a", "value": 4},
					{"timestamp": "2026-11-10T10:00:00Z", "workerpool": "pool-a", "value": 3},
				},
			},
			{"name": "Crashes today", "type": "number", "number": map[string]any{"unit": "crashes"}, "rows": []map[string]any{{"value": "7"}}},
			{"name": "Memory used", "type": "gauge", "gauge": map[string]any{"unit": " %"}, "rows": []map[string]any{{"value": 41.5}}},
			{"name": "No data yet", "type": "number", "rows": []map[string]any{}},
			{"name": "By pool", "type": "piechart", "rows": []map[string]any{{"value": 1}}},
		},
	})

	view, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "").View(t.Context(), "substrate/capacity", nil)
	require.NoError(t, err)
	assert.Equal(t, mc.URL, view.MissionControl)
	assert.Equal(t, "Substrate", view.Title)
	assert.Equal(t, time.Date(2026, 11, 10, 10, 5, 0, 0, time.UTC), view.LastRefreshedAt)
	assert.Equal(t, []map[string]any{{"variables": map[string]any{}}}, mc.ViewRequests())

	require.Len(t, view.Panels, 5)
	at := func(minute int) time.Time { return time.Date(2026, 11, 10, 10, minute, 0, 0, time.UTC) }
	assert.Equal(t, missioncontrol.Panel{
		Name: "Ready workers", Description: "Last 12h", Type: "timeseries",
		Series: []missioncontrol.Series{
			{Labels: map[string]string{"workerpool": "pool-a"}, Points: []missioncontrol.Point{{Time: at(0), Value: 3}, {Time: at(1), Value: 4}}},
			{Labels: map[string]string{"workerpool": "pool-b"}, Points: []missioncontrol.Point{{Time: at(1), Value: 2}}},
		},
	}, view.Panels[0], "series are ordered by their labels and points by time")

	seven, fortyOne := 7.0, 41.5
	assert.Equal(t, missioncontrol.Panel{Name: "Crashes today", Type: "number", Unit: "crashes", Value: &seven}, view.Panels[1])
	assert.Equal(t, missioncontrol.Panel{Name: "Memory used", Type: "gauge", Unit: "%", Value: &fortyOne}, view.Panels[2])
	assert.Equal(t, missioncontrol.Panel{Name: "No data yet", Type: "number"}, view.Panels[3], "a panel without rows has no value")
	assert.Equal(t, missioncontrol.Panel{Name: "By pool", Type: "piechart", Unsupported: "gavel doesn't draw piechart panels"}, view.Panels[4])
}

// Mission Control runs a panel's query on SQLite, which keeps a Prometheus sample's time as Go's
// default text for a time.
func TestViewReadsTimestampsAsMissionControlsSQLiteWritesThem(t *testing.T) {
	mc := missioncontroltest.New(t)
	mc.SetView("mission-control", "substrate-org-usage", map[string]any{
		"panels": []map[string]any{
			{
				"name": "CPU (cores)", "type": "timeseries",
				"timeseries": map[string]any{"timeKey": "timestamp", "valueKey": "value"},
				"rows": []map[string]any{
					{"timestamp": "2026-11-23 23:20:53.317 +0000 UTC", "value": 0.5},
					{"timestamp": "2026-11-24 04:45:53.317 +0530 IST", "value": 0.25},
				},
			},
		},
	})

	view, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "").View(t.Context(), "mission-control/substrate-org-usage", nil)
	require.NoError(t, err)
	require.Len(t, view.Panels, 1)
	assert.Empty(t, view.Panels[0].Unsupported)
	require.Len(t, view.Panels[0].Series, 1)
	points := view.Panels[0].Series[0].Points
	require.Len(t, points, 2)
	assert.True(t, points[0].Time.Equal(time.Date(2026, 11, 23, 23, 15, 53, 317_000_000, time.UTC)), "got %s", points[0].Time)
	assert.Equal(t, 0.25, points[0].Value)
	assert.True(t, points[1].Time.Equal(time.Date(2026, 11, 23, 23, 20, 53, 317_000_000, time.UTC)), "got %s", points[1].Time)
	assert.Equal(t, 0.5, points[1].Value)
}

func TestViewSaysWhyAPanelCantBeRead(t *testing.T) {
	mc := missioncontroltest.New(t)
	mc.SetView("substrate", "capacity", map[string]any{
		"refreshError": "prometheus: connection refused",
		"panels": []map[string]any{
			{"name": "Ready workers", "type": "timeseries", "rows": []map[string]any{{"timestamp": "yesterday", "value": 1}}},
			{"name": "Crashes", "type": "number", "rows": []map[string]any{{"value": "many"}}},
		},
	})

	view, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "").View(t.Context(), "substrate/capacity", nil)
	require.NoError(t, err)
	assert.Equal(t, "prometheus: connection refused", view.RefreshError)
	require.Len(t, view.Panels, 2)
	assert.Equal(t, `row 1's timestamp "yesterday" isn't an RFC 3339 time or Go's text for one`, view.Panels[0].Unsupported)
	assert.Empty(t, view.Panels[0].Series)
	assert.Equal(t, `row 1's value "many" isn't a number`, view.Panels[1].Unsupported)
}

func TestViewExplainsFailures(t *testing.T) {
	mc := missioncontroltest.New(t)
	source := missioncontrol.New(mc.URL, missioncontroltest.Token, "")

	_, err := source.View(t.Context(), "substrate/missing", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "view substrate/missing not found in Mission Control at "+mc.URL)

	_, err = source.View(t.Context(), "capacity", nil)
	require.EqualError(t, err, `view "capacity" must be namespace/name`)

	mc.SetView("substrate", "capacity", map[string]any{})
	_, err = missioncontrol.New(mc.URL, "wrong-token", "").View(t.Context(), "substrate/capacity", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read view substrate/capacity from Mission Control at "+mc.URL)
	assert.Contains(t, err.Error(), "check GAVEL_MISSION_CONTROL_TOKEN")
}

func TestViewPassesItsVariables(t *testing.T) {
	mc := missioncontroltest.New(t)
	mc.SetView("substrate", "org-usage", map[string]any{"title": "Measured use"})

	view, err := missioncontrol.New(mc.URL, missioncontroltest.Token, "").View(t.Context(), "substrate/org-usage", map[string]string{"org": "f1af4645"})
	require.NoError(t, err)
	assert.Equal(t, "Measured use", view.Title)
	assert.Equal(t, []map[string]any{{"variables": map[string]any{"org": "f1af4645"}}}, mc.ViewRequests())
}
