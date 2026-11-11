// Tests finding the organizations of sandbox runs, as the substrate inventory names its actors'
// runs: runs of this Clerk instance's organizations only.
package platform_test

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/platform"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunOrgs(t *testing.T) {
	db := openDB(t)
	alpha := addOrg(t, db, issuer, "org_alpha", "alpha", "Alpha")
	other := addOrg(t, db, "https://other-clerk.example", "org_other", "other", "Another Clerk instance")
	run := func(org uuid.UUID, kind string) uuid.UUID {
		id := uuid.New()
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, created_at) VALUES (?, ?, ?, 'running', 'substrate', ?)`, id, org, kind, time.Now()).Error)
		return id
	}
	agent, lint, foreign, unknown := run(alpha, "agent"), run(alpha, "lint"), run(other, "test"), uuid.New()

	orgs, err := platform.Store{DB: db, Issuer: issuer}.RunOrgs(t.Context(), []uuid.UUID{agent, lint, foreign, unknown})
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]platform.RunOrg{
		agent: {Org: alpha, OrgName: "Alpha", Kind: "agent"},
		lint:  {Org: alpha, OrgName: "Alpha", Kind: "lint"},
	}, orgs)

	none, err := platform.Store{DB: db, Issuer: issuer}.RunOrgs(t.Context(), nil)
	require.NoError(t, err)
	assert.Empty(t, none)
}
