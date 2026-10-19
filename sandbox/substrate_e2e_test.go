// Runs the agent-substrate executor against a live Substrate cluster when GAVEL_SANDBOX_SUBSTRATE_TEST=1
// (setup in docs/sandbox.md): the job reaches serve and allowed hosts only, and its model key is injected.
package sandbox

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/flanksource/gavel/sandbox/internal/substrate/ateenvv1alpha"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSubstrateExecutorAgainstACluster(t *testing.T) {
	if os.Getenv("GAVEL_SANDBOX_SUBSTRATE_TEST") != "1" {
		t.Skip("set GAVEL_SANDBOX_SUBSTRATE_TEST=1 with a Substrate cluster set up as docs/sandbox.md describes to run the substrate executor e2e")
	}
	setting := func(name, fallback string) string {
		if value := os.Getenv(name); value != "" {
			return value
		}
		require.NotEmpty(t, fallback, "set %s", name)
		return fallback
	}
	image := setting("GAVEL_SUBSTRATE_IMAGE", "")
	tokenFile := setting("GAVEL_SUBSTRATE_TOKEN_FILE", "")
	// The run API stands in for serve on the host; a Service in the cluster names it.
	listen := setting("GAVEL_SUBSTRATE_SERVE_LISTEN", "172.18.0.1:18080")
	server := setting("GAVEL_SUBSTRATE_SERVE_URL", "http://gavel-host.default.svc.cluster.local:18080")
	credential := setting("GAVEL_SUBSTRATE_CREDENTIAL", "gavel-e2e-injected-credential")

	listener, err := net.Listen("tcp", listen)
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
		EnvAPI:           setting("GAVEL_SUBSTRATE_ENV_API", "localhost:7777"),
		Control:          setting("GAVEL_SUBSTRATE_CONTROL", "localhost:8443"),
		ControlTokenFile: tokenFile,
		Template:         setting("GAVEL_SUBSTRATE_TEMPLATE", "gavel-runner"),
		TrustBundle:      "/run/ate/egress-trust-bundle/trust-bundle.pem",
		Credentials: []InjectedCredential{{
			Host: "httpbin.org", Header: "x-gavel-key", Env: "E2E_MODEL_KEY",
			URI: setting("GAVEL_SUBSTRATE_CREDENTIAL_URI", "ate-secret://k8s.io/default/gavel-e2e/model/key"),
		}},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	handle, err := executor.Submit(ctx, Job{
		RunID: uuid.New(),
		Image: image,
		Command: []string{"sh", "-c", `echo hi
echo "token=$GAVEL_RUN_TOKEN"
echo "key=$E2E_MODEL_KEY"
echo "serve=$(curl -sS -m 20 "$GAVEL_SERVER/ping")"
echo "golang=$(curl -sS -m 20 -o /dev/null -w '%{http_code}' https://proxy.golang.org/)"
curl -sS -m 10 -o /dev/null https://example.com 2>/dev/null; echo "example=$?"
curl -sS -m 20 --cacert "$NODE_EXTRA_CA_CERTS" -H "x-gavel-key: $E2E_MODEL_KEY" https://httpbin.org/headers
exit 3`},
		Env:       map[string]string{"GAVEL_SERVER": server},
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
	assert.NotContains(t, out, "key="+credential)
	assert.Contains(t, out, "serve=pong", "the job reaches serve")
	assert.Equal(t, "/ping", <-pinged)
	assert.Contains(t, out, "golang=200", "the job reaches the Go module proxy")
	assert.Contains(t, out, "example=6", "the job can't even resolve a host its policy doesn't allow")
	assert.Contains(t, out, `"X-Gavel-Key": "`+credential+`"`, "the gateway injects the model key")

	result, err := handle.Wait(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ExitCode)

	substrate := executor.(*substrateExecutor)
	_, err = substrate.envs.GetEnvironment(ctx, &ateenvv1alpha.GetEnvironmentRequest{Id: handle.Ref(), Atespace: substrate.options.Atespace})
	assert.Equal(t, codes.NotFound, status.Code(err), "the environment is deleted once the job exits: %v", err)
}
