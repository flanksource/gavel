// Runs the Kubernetes executor against a real cluster (kind) when GAVEL_SANDBOX_KIND_TEST=1:
// busybox's output comes back through Logs and its exit code through Wait.
package sandbox

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesExecutorAgainstACluster(t *testing.T) {
	if os.Getenv("GAVEL_SANDBOX_KIND_TEST") != "1" {
		t.Skip("set GAVEL_SANDBOX_KIND_TEST=1 with KUBECONFIG pointing at a kind cluster to run the Kubernetes executor e2e")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	namespace := "gavel-sandbox-test-" + uuid.NewString()[:8]
	executor, err := NewKubernetes(ctx, KubernetesOptions{Namespace: namespace})
	require.NoError(t, err)
	_, err = executor.client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = executor.client.CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})
	})

	executor.options.ServePodSelector = map[string]string{"app.kubernetes.io/name": "gavel"}
	executor.options.ServePort = 9092
	require.NoError(t, executor.EnsureNetworkPolicy(ctx), "the API server accepts the NetworkPolicy")
	require.NoError(t, executor.EnsureNetworkPolicy(ctx), "and applying it again")

	handle, err := executor.Submit(ctx, Job{
		RunID:     uuid.New(),
		Image:     "busybox:1.37",
		Command:   []string{"sh", "-c", `echo hi; echo "token=$GAVEL_RUN_TOKEN"; exit 3`},
		SecretEnv: map[string]string{"GAVEL_RUN_TOKEN": "secret-token"},
		Limits:    Limits{CPU: "100m", Memory: "64Mi", Timeout: 2 * time.Minute},
	})
	require.NoError(t, err)

	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	output, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	assert.Contains(t, string(output), "hi")
	assert.Contains(t, string(output), "token=secret-token", "the Secret reaches the pod")

	result, err := handle.Wait(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ExitCode)
}
