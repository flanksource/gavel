// Tests what the agent-substrate executor asks of Substrate for a job: the environment's egress
// allowlist, with model keys injected outside it, and the job's process, run on a real in-process Substrate.
package sandbox

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/captain/pkg/substrate/substratetest"
	"github.com/flanksource/gavel/internal/logtest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCredential = substrate.Credential{
	Host: "api.anthropic.com", Header: "x-api-key", URI: "ate-secret://gavel/beta/6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f/anthropic", Env: "ANTHROPIC_API_KEY",
}

func TestSubstrateEgressAllowsOnlyWhatJobsNeed(t *testing.T) {
	egress, err := substrateEgress([]string{"goproxy.example.com"}, "http://gavel.gavel.svc.cluster.local:8080", []substrate.Credential{testCredential})
	require.NoError(t, err)
	assert.Equal(t, substrate.Egress{
		HTTP: []substrate.HostPort{{Host: "gavel.gavel.svc.cluster.local", Port: 8080}},
		Passthrough: []substrate.HostPort{
			{Host: "proxy.golang.org", Port: 443}, {Host: "sum.golang.org", Port: 443},
			{Host: "registry.npmjs.org", Port: 443},
			{Host: "pypi.org", Port: 443}, {Host: "files.pythonhosted.org", Port: 443},
			{Host: "github.com", Port: 443}, {Host: "codeload.github.com", Port: 443}, {Host: "objects.githubusercontent.com", Port: 443},
			{Host: "goproxy.example.com", Port: 443},
		},
		Credentials: []substrate.Credential{testCredential},
	}, egress, "jobs reach serve's run API, the package hosts, and the model API with its key injected by the gateway")

	tls, err := substrateEgress(nil, "https://gavel.example.com", nil)
	require.NoError(t, err)
	assert.Empty(t, tls.HTTP)
	assert.Equal(t, substrate.HostPort{Host: "gavel.example.com", Port: 443}, tls.Passthrough[0], "an HTTPS serve is passed through")

	_, err = substrateEgress(nil, "http://10.0.0.5:8080", nil)
	require.EqualError(t, err, `substrate egress names hosts: GAVEL_SERVER "http://10.0.0.5:8080" must name serve by DNS name, not an IP address`)
}

func TestSubstrateEgressRefusesAPortOutOfRange(t *testing.T) {
	for _, port := range []string{"0", "65536", "4294967297"} {
		server := "http://gavel.example.com:" + port
		_, err := substrateEgress(nil, server, nil)
		require.EqualError(t, err, `substrate egress: GAVEL_SERVER "`+server+`" has port `+port+`, outside 1-65535`)
	}
}

func TestSubstrateProcessRunsTheJobWithItsRunTokenOnly(t *testing.T) {
	run := uuid.New()
	job := Job{
		RunID:       run,
		Image:       "ghcr.io/flanksource/gavel-runner@sha256:" + strings64("a"),
		Command:     []string{"gavel", "runner", "exec", "--run", run.String()},
		Env:         map[string]string{"GAVEL_SERVER": "http://gavel.gavel.svc.cluster.local:8080", "CI": "1"},
		SecretEnv:   map[string]string{"GAVEL_RUN_TOKEN": "run-token"},
		Credentials: []substrate.Credential{testCredential},
		Limits:      Limits{Timeout: 20 * time.Minute},
	}
	process, err := substrateProcess(job)
	require.NoError(t, err)
	assert.Equal(t, substrate.Process{
		Command: job.Command,
		Env:     map[string]string{"GAVEL_SERVER": "http://gavel.gavel.svc.cluster.local:8080", "CI": "1", "GAVEL_RUN_TOKEN": "run-token"},
		Timeout: 20 * time.Minute,
	}, process)

	job.SecretEnv["ANTHROPIC_API_KEY"] = "sk-ant-secret"
	_, err = substrateProcess(job)
	require.EqualError(t, err, "secret ANTHROPIC_API_KEY would enter the substrate environment; give it to substrate's egress gateway as an injected credential instead")
	assert.NotContains(t, err.Error(), "sk-ant-secret")
}

func strings64(c string) string {
	out := ""
	for range 64 {
		out += c
	}
	return out
}

// startSubstrate serves an in-process Substrate whose processes are real, and an executor on it,
// with the lines Submit logs captured.
func startSubstrate(t *testing.T) (*substratetest.Server, Executor, *strings.Builder) {
	t.Helper()
	logged := logtest.Capture(t)
	server := substratetest.Start(t)
	options := server.Options()
	options.TrustBundle = "/run/ate/trust.pem"
	executor, err := NewSubstrate(SubstrateOptions{Options: options})
	require.NoError(t, err)
	return server, executor, logged
}

func testJob(run uuid.UUID, script string) Job {
	return Job{
		RunID:       run,
		Image:       "ghcr.io/flanksource/gavel-runner@sha256:" + strings64("a"),
		Command:     []string{"sh", "-c", script},
		Env:         map[string]string{"GAVEL_SERVER": "https://gavel.example.com"},
		SecretEnv:   map[string]string{"GAVEL_RUN_TOKEN": "run-token"},
		Credentials: []substrate.Credential{testCredential},
	}
}

func TestSubstrateRunsTheJobAndRemovesItsEnvironment(t *testing.T) {
	server, executor, _ := startSubstrate(t)
	run := uuid.MustParse("6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	handle, err := executor.Submit(ctx, testJob(run, `echo "token=$GAVEL_RUN_TOKEN key=$ANTHROPIC_API_KEY ca=$NODE_EXTRA_CA_CERTS"; exit 3`))
	require.NoError(t, err)
	assert.Equal(t, "gavel-run-"+run.String(), handle.Ref())
	policy, ok := server.Policy("ate-env", handle.Ref())
	require.True(t, ok, "the environment gets its egress policy")
	assert.NotEmpty(t, policy.GetRules())

	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	output, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	assert.Equal(t, "token=run-token key="+substrate.Placeholder+" ca=/run/ate/trust.pem\n", string(output),
		"the job gets its run token, a placeholder for its model key, and the gateway's CA")

	result, err := handle.Wait(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, result.ExitCode)
	assert.Equal(t, []string{"ate-env/gavel-run-" + run.String()}, server.Deleted(), "the environment is deleted once the job exits")
}

func TestSubstrateCancelKillsTheJobAndRemovesItsEnvironment(t *testing.T) {
	server, executor, _ := startSubstrate(t)
	run := uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	handle, err := executor.Submit(ctx, testJob(run, "sleep 60"))
	require.NoError(t, err)
	require.NoError(t, handle.Cancel(ctx))
	assert.Equal(t, []string{"ate-env/gavel-run-" + run.String()}, server.Deleted())
}

func TestSubstrateSubmitRemovesAnEnvironmentWhoseJobDoesNotStart(t *testing.T) {
	server, executor, _ := startSubstrate(t)
	run := uuid.New()
	job := testJob(run, "")
	job.Command = []string{"/no/such/program"}

	_, err := executor.Submit(context.Background(), job)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start /no/such/program in substrate environment ate-env/gavel-run-"+run.String())
	assert.Equal(t, []string{"ate-env/gavel-run-" + run.String()}, server.Deleted())
}

func TestSubstrateSubmitRefusesWhatSubstrateCannotRun(t *testing.T) {
	server, executor, _ := startSubstrate(t)
	job := testJob(uuid.New(), "true")
	job.Image = "ghcr.io/flanksource/gavel-runner:1"
	_, err := executor.Submit(context.Background(), job)
	require.EqualError(t, err, `substrate runs images pinned by digest: "ghcr.io/flanksource/gavel-runner:1" must be repo@sha256:<digest>`)

	server.RefusePolicy = errors.New("policy refused")
	_, err = executor.Submit(context.Background(), testJob(uuid.New(), "true"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "set the egress policy of substrate environment ate-env/gavel-run-")
}

func TestSubstrateSubmitLogsTheCredentialsTheGatewayInjects(t *testing.T) {
	_, executor, logged := startSubstrate(t)
	run := uuid.MustParse("6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f")

	_, err := executor.Submit(context.Background(), testJob(run, "true"))
	require.NoError(t, err)
	assert.Contains(t, logged.String(), "sandbox run 6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f in substrate environment ate-env/gavel-run-6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f: the egress gateway injects api.anthropic.com x-api-key from ate-secret://gavel/beta/6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f/anthropic")

	keyless := testJob(uuid.MustParse("7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"), "true")
	keyless.Credentials = nil
	_, err = executor.Submit(context.Background(), keyless)
	require.NoError(t, err)
	assert.Contains(t, logged.String(), "sandbox run 7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d in substrate environment ate-env/gavel-run-7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d: the egress gateway injects no credentials")
}

func TestSubstrateActorIsTheSPIFFEIDOfTheRunsEnvironment(t *testing.T) {
	run := uuid.MustParse("6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f")
	assert.Equal(t, "spiffe://substrate-actor.local/actor/ate-env/gavel-run-6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f", SubstrateActor("ate-env", run))
}
