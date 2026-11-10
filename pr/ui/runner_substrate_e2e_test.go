// Runs push checks end to end on a live Substrate cluster when GAVEL_SANDBOX_SUBSTRATE_TEST=1: each
// job, a gVisor environment, runs in a checkout of the pushed commit and reports back to this
// host's run API over its run's channel.
package ui

import (
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/substratetest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPushChecksRunOnSubstrate(t *testing.T) {
	cluster := substratetest.FromEnv(t)
	db, store, org := openOrgDB(t)
	dataDir := t.TempDir()
	dispatcher := substrateDispatcher(t, cluster, &RunsAPI{DB: db, Cache: store, Sources: &sandbox.Sources{DataDir: dataDir}})
	assertPushChecks(t, dispatcher, db, org, dataDir)
}

// substrateDispatcher starts jobs as environments on cluster, which reach api over their runs'
// channels; it is also served where the cluster's Service for serve points, which jobs don't reach.
func substrateDispatcher(t *testing.T, cluster substratetest.Cluster, api *RunsAPI) *sandbox.Dispatcher {
	t.Helper()
	listener, err := net.Listen("tcp", cluster.Listen)
	require.NoError(t, err)
	api.TokenKey = []byte(strings.Repeat("k", 32))
	server := &http.Server{Handler: api.Handler()}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })
	executor, err := sandbox.NewSubstrate(sandbox.SubstrateOptions{Options: substrate.Options{
		EnvAPI: cluster.EnvAPI, Control: cluster.Control, ControlTokenFile: cluster.TokenFile, Template: cluster.Template,
	}, RunAPI: api.Handler()})
	require.NoError(t, err)
	logs := blob.Local(t.TempDir())
	return &sandbox.Dispatcher{
		DB: api.DB, Server: cluster.ServeURL, TokenKey: api.TokenKey,
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: 10 * time.Minute},
		Executor: executor, Sources: api.Sources,
	}
}
