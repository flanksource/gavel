// Runs the agent-substrate executor against a live Substrate cluster when GAVEL_SANDBOX_SUBSTRATE_TEST=1
// (setup in docs/sandbox.md): the job reaches serve and allowed hosts only, and its model key is injected.
package sandbox

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/flanksource/gavel/sandbox/internal/substrate/ateenvv1alpha"
	"github.com/flanksource/gavel/sandbox/substratetest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSubstrateExecutorAgainstACluster(t *testing.T) {
	cluster := substratetest.FromEnv(t)
	// A server on the host stands in for serve; a Service in the cluster names it.
	listener, err := net.Listen("tcp", cluster.Listen)
	require.NoError(t, err)
	pinged := make(chan string, 1)
	serve := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case pinged <- r.URL.Path:
		default:
		}
		_, _ = io.WriteString(w, "pong")
	})}
	go func() { _ = serve.Serve(listener) }()
	t.Cleanup(func() { _ = serve.Close() })

	executor, err := NewSubstrate(SubstrateOptions{
		EnvAPI: cluster.EnvAPI, Control: cluster.Control, ControlTokenFile: cluster.TokenFile,
		Template: cluster.Template, TrustBundle: substratetest.TrustBundle,
		Credentials: []InjectedCredential{{Host: "httpbin.org", Header: "x-gavel-key", Env: "E2E_MODEL_KEY", URI: cluster.CredentialURI}},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	handle, err := executor.Submit(ctx, Job{
		RunID: uuid.New(),
		Image: cluster.Image,
		Command: []string{"sh", "-c", `echo hi
echo "token=$GAVEL_RUN_TOKEN"
echo "key=$E2E_MODEL_KEY"
echo "serve=$(curl -sS -m 20 "$GAVEL_SERVER/ping")"
echo "golang=$(curl -sS -m 20 -o /dev/null -w '%{http_code}' https://proxy.golang.org/)"
curl -sS -m 10 -o /dev/null https://example.com 2>/dev/null; echo "example=$?"
curl -sS -m 20 --cacert "$NODE_EXTRA_CA_CERTS" -H "x-gavel-key: $E2E_MODEL_KEY" https://httpbin.org/headers
exit 3`},
		Env:       map[string]string{"GAVEL_SERVER": cluster.ServeURL},
		SecretEnv: map[string]string{"GAVEL_RUN_TOKEN": "secret-token"},
		Limits:    Limits{Timeout: 5 * time.Minute},
	})
	require.NoError(t, err)

	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	output, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	out := string(output)
	assert.Contains(t, out, "hi\n")
	assert.Contains(t, out, "token=secret-token", "the run token reaches the job")
	assert.Contains(t, out, "key="+substrateCredentialPlaceholder+"\n", "the job holds only a placeholder for its model key")
	assert.NotContains(t, out, "key="+cluster.Credential)
	assert.Contains(t, out, "serve=pong", "the job reaches serve")
	assert.Equal(t, "/ping", <-pinged)
	assert.Contains(t, out, "golang=200", "the job reaches the Go module proxy")
	assert.Contains(t, out, "example=6", "the job can't even resolve a host its policy doesn't allow")
	assert.Contains(t, out, `"X-Gavel-Key": "`+cluster.Credential+`"`, "the gateway injects the model key")

	result, err := handle.Wait(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ExitCode)

	substrate := executor.(*substrateExecutor)
	_, err = substrate.envs.GetEnvironment(ctx, &ateenvv1alpha.GetEnvironmentRequest{Id: handle.Ref(), Atespace: substrate.options.Atespace})
	assert.Equal(t, codes.NotFound, status.Code(err), "the environment is deleted once the job exits: %v", err)
}
