// Runs the container executor against real docker with busybox when GAVEL_SANDBOX_DOCKER_TEST=1:
// logs, exit codes, secrets through the env file, cancellation and the timeout.
package sandbox

import (
	"context"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dockerExecutor(t *testing.T) Executor {
	t.Helper()
	if os.Getenv("GAVEL_SANDBOX_DOCKER_TEST") != "1" {
		t.Skip("set GAVEL_SANDBOX_DOCKER_TEST=1 to run the container executor against docker")
	}
	network := "gavel-sandbox-test-" + uuid.NewString()[:8]
	require.NoError(t, exec.Command("docker", "network", "create", network).Run())
	t.Cleanup(func() { _ = exec.Command("docker", "network", "rm", network).Run() })
	return NewContainer("docker", ContainerOptions{Network: network})
}

func containerExists(name string) bool {
	return exec.Command("docker", "inspect", name).Run() == nil
}

func TestContainerExecutorRunsAJob(t *testing.T) {
	executor := dockerExecutor(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	job := Job{
		RunID:     uuid.New(),
		Image:     "busybox:1.37",
		Command:   []string{"sh", "-c", `echo hi; echo "token=$GAVEL_RUN_TOKEN"; echo "id=$(id -u)"; exit 3`},
		Env:       map[string]string{"CI": "1"},
		SecretEnv: map[string]string{"GAVEL_RUN_TOKEN": "secret-token"},
	}
	handle, err := executor.Submit(ctx, job)
	require.NoError(t, err)
	assert.Equal(t, "gavel-run-"+job.RunID.String(), handle.Ref())

	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	output, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	assert.Contains(t, string(output), "hi")
	assert.Contains(t, string(output), "token=secret-token", "the secret reaches the job")
	assert.Contains(t, string(output), "id=1000", "the job runs unprivileged")

	result, err := handle.Wait(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ExitCode)
	assert.False(t, containerExists(handle.Ref()), "Wait removes the container")
}

func TestContainerExecutorCancels(t *testing.T) {
	executor := dockerExecutor(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	handle, err := executor.Submit(ctx, Job{RunID: uuid.New(), Image: "busybox:1.37", Command: []string{"sleep", "60"}})
	require.NoError(t, err)

	require.NoError(t, handle.Cancel(ctx))
	assert.False(t, containerExists(handle.Ref()))
	_, err = handle.Wait(ctx)
	require.Error(t, err, "a cancelled job has no result")
}

func TestContainerExecutorEnforcesTheTimeout(t *testing.T) {
	executor := dockerExecutor(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	handle, err := executor.Submit(ctx, Job{RunID: uuid.New(), Image: "busybox:1.37", Command: []string{"sleep", "60"}, Limits: Limits{Timeout: 2 * time.Second}})
	require.NoError(t, err)

	started := time.Now()
	_, err = handle.Wait(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out after 2s")
	assert.Less(t, time.Since(started), 30*time.Second)
	assert.False(t, containerExists(handle.Ref()), "a timed-out job is removed")
}
