// Runs push checks end to end on a live Substrate cluster when GAVEL_SANDBOX_SUBSTRATE_TEST=1: each
// job, a gVisor environment, fetches the pushed source from this host's run API and reports back.
package ui

import (
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

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

// substrateDispatcher starts jobs as environments on cluster, which reach api, served where the
// cluster's Service for serve points.
func substrateDispatcher(t *testing.T, cluster substratetest.Cluster, api *RunsAPI) *sandbox.Dispatcher {
	t.Helper()
	listener, err := net.Listen("tcp", cluster.Listen)
	require.NoError(t, err)
	api.TokenKey = []byte(strings.Repeat("k", 32))
	server := &http.Server{Handler: api.Handler()}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })
	executor, err := sandbox.NewSubstrate(sandbox.SubstrateOptions{
		EnvAPI: cluster.EnvAPI, Control: cluster.Control, ControlTokenFile: cluster.TokenFile, Template: cluster.Template,
	})
	require.NoError(t, err)
	logs := blob.Local(t.TempDir())
	return &sandbox.Dispatcher{
		DB: api.DB, Image: cluster.Image, Server: cluster.ServeURL, TokenKey: api.TokenKey,
		Logs: func(uuid.UUID) blob.Store { return logs }, Limits: sandbox.Limits{Timeout: 10 * time.Minute},
		Executor: executor,
	}
}
