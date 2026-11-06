// Runs the agent-substrate executor against a live Substrate cluster when GAVEL_SANDBOX_SUBSTRATE_TEST=1
// (setup in docs/sandbox.md): the job runs in a checkout of its commit, asks serve over its run's
// channel, reaches allowed hosts only, and its model key is injected.
package sandbox

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/gavel/sandbox/substratetest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstrateExecutorAgainstACluster(t *testing.T) {
	cluster := substratetest.FromEnv(t)
	// A server on the host stands in for serve, where the cluster's Service for it points: the job
	// must not reach it, since it asks serve over its run's channel.
	listener, err := net.Listen("tcp", cluster.Listen)
	require.NoError(t, err)
	reached := make(chan string, 1)
	serve := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case reached <- r.URL.Path:
		default:
		}
	})}
	go func() { _ = serve.Serve(listener) }()
	t.Cleanup(func() { _ = serve.Close() })

	api := &fakeRunAPI{}
	executor, err := NewSubstrate(SubstrateOptions{Options: substrate.Options{
		EnvAPI: cluster.EnvAPI, Control: cluster.Control, ControlTokenFile: cluster.TokenFile,
		Template: cluster.Template, TrustBundle: substratetest.TrustBundle,
	}, RunAPI: api.handler()})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	run := uuid.New()
	repo, first := jobRepository(t)
	handle, err := executor.Submit(ctx, Job{
		RunID: run,
		Command: []string{"sh", "-c", `echo hi
echo "readme=$(cat README.md)"
echo "token=$GAVEL_RUN_TOKEN"
echo "key=$E2E_MODEL_KEY"
echo "spec=$(curl -sS -m 20 --unix-socket "$GAVEL_RUN_API_SOCKET" "http://run-api/api/runs/$GAVEL_RUN_ID/spec")"
curl -sS -m 10 -o /dev/null "` + cluster.ServeURL + `/ping" 2>/dev/null; echo "serve=$?"
echo "golang=$(curl -sS -m 20 -o /dev/null -w '%{http_code}' https://proxy.golang.org/)"
curl -sS -m 10 -o /dev/null https://example.com 2>/dev/null; echo "example=$?"
curl -sS -m 20 --cacert "$NODE_EXTRA_CA_CERTS" -H "x-gavel-key: $E2E_MODEL_KEY" https://httpbin.org/headers
exit 3`},
		Env:         map[string]string{"GAVEL_SERVER": cluster.ServeURL, "GAVEL_RUN_ID": run.String()},
		SecretEnv:   map[string]string{"GAVEL_RUN_TOKEN": "run-token"},
		Credentials: []substrate.Credential{{Host: "httpbin.org", Header: "x-gavel-key", Env: "E2E_MODEL_KEY", URI: cluster.CredentialURI}},
		Limits:      Limits{Timeout: 5 * time.Minute},
		Source:      &JobSource{GitDir: repo, Commit: first},
	})
	require.NoError(t, err)

	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	output, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	out := string(output)
	assert.Contains(t, out, "hi\n")
	assert.Contains(t, out, "readme=first\n", "the job runs in a checkout of its commit")
	assert.Contains(t, out, "token=\n", "the run token stays out of the environment")
	assert.Contains(t, out, "key="+substrate.Placeholder+"\n", "the job holds only a placeholder for its model key")
	assert.NotContains(t, out, "key="+cluster.Credential)
	assert.Contains(t, out, `spec={"kind":"test"}`, "the job asks serve over its run's channel")
	assert.NotContains(t, out, "serve=0\n", "the job can't reach serve's address")
	select {
	case path := <-reached:
		t.Errorf("the job reached serve's address at %s", path)
	default:
	}
	assert.Contains(t, out, "golang=200", "the job reaches the Go module proxy")
	assert.Contains(t, out, "example=6", "the job can't even resolve a host its policy doesn't allow")
	assert.Contains(t, out, `"X-Gavel-Key": "`+cluster.Credential+`"`, "the gateway injects the model key")

	result, err := handle.Wait(ctx)
	require.NoError(t, err, "the job's run deletes its environment once the job exits")
	assert.Equal(t, 3, result.ExitCode)
}
