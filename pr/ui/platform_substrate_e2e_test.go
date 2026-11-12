// Lists a live Substrate cluster's inventory through the platform API when
// GAVEL_SANDBOX_SUBSTRATE_TEST=1: its workers, and a run's environment under the run's organization.
package ui

import (
	"context"
	"net/http"
	"testing"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/gavel/platform"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/substratetest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformListsALiveSubstratesInventory(t *testing.T) {
	cluster := substratetest.FromEnv(t)
	db, _, org := openOrgDB(t)
	run := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, created_at) VALUES (?, ?, 'agent', 'running', 'substrate', now())`, run, org).Error)
	client, err := substrate.Dial(substrate.Options{
		EnvAPI: cluster.EnvAPI, Control: cluster.Control, ControlTokenFile: cluster.TokenFile, Template: cluster.Template,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	env, err := client.Create(t.Context(), substrate.Environment{ID: sandbox.JobName(run)})
	require.NoError(t, err)
	// t.Context() is canceled before cleanups run, so the environment is deleted with its own.
	t.Cleanup(func() { require.NoError(t, env.Delete(context.Background())) })
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}, Substrate: client}

	var got PlatformSubstrateInventory
	require.Equal(t, http.StatusOK, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/substrate/inventory", &got))
	require.Empty(t, got.WorkersError, "gavel's token needs substrate's global viewer role to list workers")
	require.Empty(t, got.ActorsError, "gavel's token needs substrate's global viewer role to list every atespace's actors")
	assert.NotEmpty(t, got.Workers)
	var listed *PlatformActor
	for i, actor := range got.Actors {
		if actor.Name == sandbox.JobName(run) {
			listed = &got.Actors[i]
		}
	}
	require.NotNil(t, listed, "the run's environment is among %d actors", len(got.Actors))
	assert.Equal(t, &platform.RunOrg{Org: org, OrgName: "Other", Kind: "agent"}, listed.Org)
}
