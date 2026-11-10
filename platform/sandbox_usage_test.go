// Tests the platform dashboard's sandbox usage against a migrated database: runs per day,
// organization, kind and outcome, how long they ran and waited, and those still active.
package platform_test

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxUsage(t *testing.T) {
	db := openDB(t)
	alpha := addOrg(t, db, issuer, "org_alpha", "alpha", "Alpha")
	beta := addOrg(t, db, issuer, "org_beta", "beta", "Beta")
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	at := func(day time.Time, offset time.Duration) *time.Time { t := day.Add(offset); return &t }

	// Two finished test runs yesterday: 60s and 120s of running after 5s and 15s queued.
	addRun(t, db, alpha, "test", "succeeded", yesterday.Add(time.Hour), at(yesterday, time.Hour+5*time.Second), at(yesterday, time.Hour+65*time.Second))
	addRun(t, db, alpha, "test", "succeeded", yesterday.Add(2*time.Hour), at(yesterday, 2*time.Hour+15*time.Second), at(yesterday, 2*time.Hour+135*time.Second))
	// An agent run still running today, and a lint run still queued.
	addRun(t, db, alpha, "agent", "running", today, at(today, 10*time.Second), nil)
	addRun(t, db, beta, "lint", "queued", today, nil, nil)
	// A failed run before the window, counted only as history.
	addRun(t, db, beta, "lint", "failed", today.Add(-20*24*time.Hour), nil, nil)

	usage, err := platform.Store{DB: db, Issuer: issuer}.SandboxUsage(t.Context(), today.Add(-7*24*time.Hour))
	require.NoError(t, err)

	day := yesterday.Format(time.DateOnly)
	assert.Equal(t, []platform.SandboxDay{
		{Day: day, Org: alpha, OrgName: "Alpha", Kind: "test", Backend: "substrate", Status: "succeeded", Runs: 2, RunSeconds: 180, QueueSeconds: 20},
		{Day: today.Format(time.DateOnly), Org: alpha, OrgName: "Alpha", Kind: "agent", Backend: "substrate", Status: "running", Runs: 1, QueueSeconds: 10},
		{Day: today.Format(time.DateOnly), Org: beta, OrgName: "Beta", Kind: "lint", Backend: "substrate", Status: "queued", Runs: 1},
	}, usage.Days, "days in order; an unfinished run adds no run time, an unstarted one no queue time")
	assert.ElementsMatch(t, []platform.SandboxActive{
		{Org: alpha, OrgName: "Alpha", Kind: "agent", Status: "running", Runs: 1},
		{Org: beta, OrgName: "Beta", Kind: "lint", Status: "queued", Runs: 1},
	}, usage.Active)
}
