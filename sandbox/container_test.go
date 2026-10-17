// Tests the docker/podman run arguments a sandbox job becomes: isolation flags, limits, labels,
// and secrets only through the env file.
package sandbox

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainerRunArgs(t *testing.T) {
	run := uuid.MustParse("6f1c2b8e-1111-4222-8333-444455556666")
	org := uuid.MustParse("0a0b0c0d-1111-4222-8333-444455556666")
	job := Job{
		RunID:     run,
		Scope:     tenant.Scope{OrgID: org},
		Image:     "ghcr.io/flanksource/gavel-runner:1",
		Command:   []string{"gavel", "runner", "exec", "--run", run.String()},
		Env:       map[string]string{"GAVEL_SERVER": "https://gavel.example", "CI": "1"},
		SecretEnv: map[string]string{"GAVEL_RUN_TOKEN": "secret-token"},
		Limits:    Limits{CPU: "500m", Memory: "4Gi", Timeout: time.Hour},
	}

	args, err := containerRunArgs(job, "gavel-sandbox", "/tmp/env-file")
	require.NoError(t, err)
	assert.Equal(t, []string{
		"run", "-d", "--name", "gavel-run-" + run.String(),
		"--network", "gavel-sandbox",
		"--user", "1000:1000", "--read-only",
		"--tmpfs", "/tmp:exec", "--tmpfs", "/workspace:exec,size=10g,uid=1000,gid=1000", "--tmpfs", "/home/gavel:exec,uid=1000,gid=1000",
		"--cap-drop", "ALL", "--security-opt", "no-new-privileges", "--pids-limit", "1024",
		"--cpus", "0.5", "--memory", "4294967296",
		"-e", "CI=1", "-e", "GAVEL_SERVER=https://gavel.example",
		"--env-file", "/tmp/env-file",
		"--label", "gavel.dev/run=" + run.String(), "--label", "gavel.dev/org=" + org.String(),
		"--entrypoint", "gavel",
		"ghcr.io/flanksource/gavel-runner:1", "runner", "exec", "--run", run.String(),
	}, args, "the command is the whole argv, replacing the image's entrypoint as a Kubernetes container's command does")
	assert.NotContains(t, strings.Join(args, " "), "secret-token", "secrets never reach the command line")
}

func TestContainerRunArgsDefaultsAndErrors(t *testing.T) {
	job := Job{RunID: uuid.New(), Image: "busybox", Command: []string{"true"}}
	args, err := containerRunArgs(job, "net", "")
	require.NoError(t, err)
	joined := strings.Join(args, " ")
	assert.Contains(t, joined, "--cpus 2 --memory 4294967296", "default limits")
	assert.NotContains(t, joined, "--env-file", "no env file without secrets")

	for name, tc := range map[string]struct {
		job  Job
		want string
	}{
		"no image":      {Job{RunID: uuid.New(), Command: []string{"true"}}, "sandbox job has no image"},
		"no command":    {Job{RunID: uuid.New(), Image: "busybox"}, "sandbox job has no command"},
		"bad cpu":       {Job{RunID: uuid.New(), Image: "x", Command: []string{"true"}, Limits: Limits{CPU: "two"}}, `CPU limit "two"`},
		"bad memory":    {Job{RunID: uuid.New(), Image: "x", Command: []string{"true"}, Limits: Limits{Memory: "lots"}}, `memory limit "lots"`},
		"env with an =": {Job{RunID: uuid.New(), Image: "x", Command: []string{"true"}, Env: map[string]string{"A=B": "c"}}, `environment variable name "A=B"`},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := containerRunArgs(tc.job, "net", "")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}
