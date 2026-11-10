// Tests the agent-substrate executor on a real in-process Substrate through captain's substrate
// sandbox: a job is a bash run whose command runs in a checkout of its commit, reaching serve's run
// API over the run's channel, with no token, and its model keys injected by the gateway.
package sandbox

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/promptrun/remoterun"
	"github.com/flanksource/captain/pkg/sandbox/substraterun"
	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/captain/pkg/substrate/substratetest"
	"github.com/flanksource/gavel/internal/logtest"
	"github.com/flanksource/gavel/sandbox/runclient"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var testCredential = substrate.Credential{
	Host: "api.anthropic.com", Header: "x-api-key", URI: "ate-secret://gavel/beta/6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f/anthropic", Env: "ANTHROPIC_API_KEY",
}

// The test binary is the environment's program and the job too, as the runner image's gavel is.
const (
	substrateRunTaskArg = "substrate-run-task"
	substrateJobArg     = "substrate-job"
)

// runSubstrateTask is the environment's program, as `gavel runner run-task` is: the run's side of
// a whole run, serving the run API to what the run starts.
func runSubstrateTask() int {
	var stop func() error
	err := substraterun.RunTask(context.Background(), ".", os.Stdout, os.Stdin, func(_ remoterun.Job, _ string, call remoterun.HostCall) (remoterun.Host, error) {
		var err error
		stop, err = runclient.ServeRun(call)
		return remoterun.Host{}, err
	})
	if stop != nil {
		err = errors.Join(err, stop())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// runSubstrateJob is the job, as `gavel runner exec` is: it reads its spec through the run API,
// says what it found, posts its results and exits with code, or, given sleep, sleeps until killed.
func runSubstrateJob(args []string) int {
	if len(args) == 1 && args[0] == "sleep" {
		fmt.Println("sleeping")
		time.Sleep(time.Minute)
		return 0
	}
	code, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	client, err := runclient.FromEnv(os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	resp, err := client.Do(context.Background(), http.MethodGet, "spec", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	spec, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	readme, _ := os.ReadFile("README.md")
	fmt.Printf("spec=%s readme=%s server=%q token=%q key=%s ca=%s\n", spec, readme,
		os.Getenv("GAVEL_SERVER"), os.Getenv("GAVEL_RUN_TOKEN"), os.Getenv("ANTHROPIC_API_KEY"), os.Getenv("NODE_EXTRA_CA_CERTS"))
	if _, err := client.Do(context.Background(), http.MethodPost, "results", strings.NewReader(`{"tests":[]}`)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	return code
}

// fakeRunAPI is serve's run API: a run's spec, and the results it is posted, each asked with the
// run's token.
type fakeRunAPI struct {
	mu      sync.Mutex
	results []string
	// onSpec, when set, runs as the job asks its spec, while its environment is up.
	onSpec func(run string)
}

func (f *fakeRunAPI) handler() http.Handler {
	mux := http.NewServeMux()
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer run-token" {
				http.Error(w, `{"error":"bad token"}`, http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
	prefix := "/api/runs/{run}"
	mux.HandleFunc("GET "+prefix+"/spec", auth(func(w http.ResponseWriter, r *http.Request) {
		if f.onSpec != nil {
			f.onSpec(r.PathValue("run"))
		}
		_, _ = io.WriteString(w, `{"kind":"test"}`)
	}))
	mux.HandleFunc("POST "+prefix+"/results", auth(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.results = append(f.results, string(body))
		f.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	return mux
}

// gitRun runs git in dir as a test author.
func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}

// jobRepository is serve's bare repository of a job's repository, with two commits of README.md,
// and the first, the job's.
func jobRepository(t *testing.T) (string, string) {
	t.Helper()
	work := t.TempDir()
	gitRun(t, work, "init", "-q", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(work, "README.md"), []byte("first"), 0o644))
	gitRun(t, work, "add", ".")
	gitRun(t, work, "commit", "-q", "-m", "first")
	first := gitRun(t, work, "rev-parse", "HEAD")
	require.NoError(t, os.WriteFile(filepath.Join(work, "README.md"), []byte("second"), 0o644))
	gitRun(t, work, "commit", "-q", "-am", "second")
	bare := filepath.Join(t.TempDir(), "app.git")
	gitRun(t, work, "clone", "-q", "--bare", work, bare)
	return bare, first
}

// startSubstrate serves an in-process Substrate whose processes are real, and an executor on it
// whose jobs reach api, with the lines it logs captured.
func startSubstrate(t *testing.T, api http.Handler) (*substratetest.Server, Executor, *strings.Builder) {
	t.Helper()
	logged := logtest.Capture(t)
	server := substratetest.Start(t)
	options := server.Options()
	options.TrustBundle = "/run/ate/trust.pem"
	options.Template = "gavel-runner"
	executor, err := NewSubstrate(SubstrateOptions{Options: options, RunAPI: api, Program: []string{os.Args[0], substrateRunTaskArg}})
	require.NoError(t, err)
	return server, executor, logged
}

// testJob is a job of run whose command is the test binary's job with args, at the first commit
// of serve's repository.
func testJob(t *testing.T, run uuid.UUID, args ...string) Job {
	t.Helper()
	repo, first := jobRepository(t)
	return Job{
		RunID:       run,
		Command:     append([]string{os.Args[0], substrateJobArg}, args...),
		Env:         map[string]string{"GAVEL_SERVER": "https://gavel.example.com", "GAVEL_RUN_ID": run.String(), "GAVEL_DB": "off"},
		SecretEnv:   map[string]string{"GAVEL_RUN_TOKEN": "run-token"},
		Credentials: []substrate.Credential{testCredential},
		Limits:      Limits{Timeout: time.Minute},
		Source:      &JobSource{GitDir: repo, Commit: first},
	}
}

func TestSubstrateRunsTheJobInACheckoutOfItsCommitReachingServeOverItsChannel(t *testing.T) {
	run := uuid.MustParse("6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f")
	api := &fakeRunAPI{}
	server, executor, _ := startSubstrate(t, api.handler())
	var rules int
	api.onSpec = func(run string) {
		if policy, ok := server.Policy("ate-env", "gavel-run-"+run); ok {
			rules = len(policy.GetRules())
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	handle, err := executor.Submit(ctx, testJob(t, run, "3"))
	require.NoError(t, err)
	assert.Equal(t, "gavel-run-"+run.String(), handle.Ref())
	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	output, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	result, err := handle.Wait(ctx)
	require.NoError(t, err)

	// The log holds the environment's program's own lines too.
	assert.Contains(t, string(output), "\n"+`spec={"kind":"test"} readme=first server="" token="" key=`+substrate.Placeholder+" ca=/run/ate/trust.pem\n",
		"the job ran at its commit and read its spec through the run's channel, holding neither serve's address nor a token, "+
			"with a placeholder for its model key and the gateway's CA")
	assert.Equal(t, 3, result.ExitCode)
	assert.Equal(t, []string{`{"tests":[]}`}, api.results, "its results reached serve as the run")
	assert.NotZero(t, rules, "the environment had its egress policy while the job ran")
	assert.Equal(t, []string{"ate-env/gavel-run-" + run.String()}, server.Deleted(), "the environment is deleted once the job exits")
}

func TestSubstrateWarnsWhenAFinishedJobsEnvironmentIsNotDeleted(t *testing.T) {
	run := uuid.New()
	api := &fakeRunAPI{}
	server, executor, logged := startSubstrate(t, api.handler())
	server.RefuseDelete = status.Error(codes.Internal, "object-store plugin upload failed")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	handle, err := executor.Submit(ctx, testJob(t, run, "0"))
	require.NoError(t, err)
	result, err := handle.Wait(ctx)
	require.NoError(t, err, "the job finished, so its run stands")
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, logged.String(), "ended, but deleting its environment failed, so it is left behind until it is deleted by hand: "+
		"delete substrate environment ate-env/gavel-run-"+run.String()+": rpc error: code = Internal desc = object-store plugin upload failed")
}

func TestSubstrateCancelKillsTheJobAndRemovesItsEnvironment(t *testing.T) {
	run := uuid.New()
	server, executor, _ := startSubstrate(t, (&fakeRunAPI{}).handler())
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	handle, err := executor.Submit(ctx, testJob(t, run, "sleep"))
	require.NoError(t, err)
	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	lines := bufio.NewScanner(logs)
	started := false
	for !started && lines.Scan() {
		started = lines.Text() == "sleeping"
	}
	require.True(t, started, "the job started")
	// The log is read to its end, as the dispatcher reads it.
	go func() { _, _ = io.Copy(io.Discard, logs) }()
	require.NoError(t, handle.Cancel(ctx))
	assert.Equal(t, []string{"ate-env/gavel-run-" + run.String()}, server.Deleted())
	_, err = handle.Wait(ctx)
	require.Error(t, err, "a cancelled job did not finish")
}

func TestSubstrateReportsAJobItsSubstrateCannotRun(t *testing.T) {
	run := uuid.New()
	server, executor, _ := startSubstrate(t, (&fakeRunAPI{}).handler())
	server.RefusePolicy = errors.New("policy refused")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	handle, err := executor.Submit(ctx, testJob(t, run, "0"))
	require.NoError(t, err)
	_, err = handle.Wait(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "the job in substrate environment gavel-run-"+run.String()+": ")
	assert.Contains(t, err.Error(), "policy refused")
}

func TestSubstrateSubmitRefusesASecretThatWouldEnterTheEnvironment(t *testing.T) {
	run := uuid.New()
	_, executor, _ := startSubstrate(t, (&fakeRunAPI{}).handler())
	job := testJob(t, run, "0")
	job.SecretEnv["ANTHROPIC_API_KEY"] = "sk-ant-secret"

	_, err := executor.Submit(context.Background(), job)
	require.EqualError(t, err, "secret ANTHROPIC_API_KEY would enter the substrate environment; give it to substrate's egress gateway as an injected credential instead")
	assert.NotContains(t, err.Error(), "sk-ant-secret")
}

func TestSubstrateSubmitRefusesAJobWithoutItsSource(t *testing.T) {
	run := uuid.New()
	_, executor, _ := startSubstrate(t, (&fakeRunAPI{}).handler())
	job := testJob(t, run, "0")
	job.Source = nil

	_, err := executor.Submit(context.Background(), job)
	require.EqualError(t, err, "sandbox run "+run.String()+" names no source: a substrate job runs in a checkout of its commit, which its dispatcher's Sources finds")
}

func TestSubstrateSubmitLogsTheCredentialsTheGatewayInjects(t *testing.T) {
	run := uuid.MustParse("6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f")
	_, executor, logged := startSubstrate(t, (&fakeRunAPI{}).handler())
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	handle, err := executor.Submit(ctx, testJob(t, run, "0"))
	require.NoError(t, err)
	_, err = handle.Wait(ctx)
	require.NoError(t, err)
	assert.Contains(t, logged.String(), "sandbox run 6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f in substrate environment ate-env/gavel-run-6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f: the egress gateway injects api.anthropic.com x-api-key from ate-secret://gavel/beta/6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f/anthropic")

	keyless := uuid.MustParse("7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d")
	job := testJob(t, keyless, "0")
	job.Credentials = nil
	handle, err = executor.Submit(ctx, job)
	require.NoError(t, err)
	_, err = handle.Wait(ctx)
	require.NoError(t, err)
	assert.Contains(t, logged.String(), "sandbox run 7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d in substrate environment ate-env/gavel-run-7a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d: the egress gateway injects no credentials")
}

func TestSubstrateBackendIsCaptainsSubstrateSandboxOptions(t *testing.T) {
	backend := SubstrateBackend(substrate.Options{
		EnvAPI: "envapi.example:443", Control: "control.example:443", ControlTokenFile: "/run/token",
		Atespace: "ate-env", Template: "gavel-runner", TrustBundle: "/run/ate/trust.pem",
	}, []string{"goproxy.example.com"})
	assert.Equal(t, map[string]any{
		"envAPI": "envapi.example:443", "control": "control.example:443", "tokenFile": "/run/token",
		"atespace": "ate-env", "template": "gavel-runner", "trustBundle": "/run/ate/trust.pem",
		"egress": []string{
			"proxy.golang.org:443", "sum.golang.org:443", "registry.npmjs.org:443", "pypi.org:443", "files.pythonhosted.org:443",
			"github.com:443", "codeload.github.com:443", "objects.githubusercontent.com:443", "goproxy.example.com:443",
		},
	}, backend, "a run reaches the package hosts and GitHub, and never serve, which it asks over its channel")
}

func TestSubstrateCredentialOptionsNameEachInjectedKeyInFull(t *testing.T) {
	assert.Equal(t, []any{map[string]any{
		"host": "api.anthropic.com", "header": "x-api-key", "prefix": "", "env": "ANTHROPIC_API_KEY",
		"uri": "ate-secret://gavel/beta/6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f/anthropic",
	}}, SubstrateCredentialOptions([]substrate.Credential{testCredential}))
}

func TestShellCommandQuotesWhatTheShellWouldSplit(t *testing.T) {
	assert.Equal(t, `gavel runner exec --run 6f1c --source .`, shellCommand([]string{"gavel", "runner", "exec", "--run", "6f1c", "--source", "."}))
	assert.Equal(t, `sh -c 'echo "$HOME"; it'\''s' ''`, shellCommand([]string{"sh", "-c", `echo "$HOME"; it's`, ""}))
}

func TestSubstrateActorIsTheSPIFFEIDOfTheRunsEnvironment(t *testing.T) {
	run := uuid.MustParse("6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f")
	assert.Equal(t, "spiffe://substrate-actor.local/actor/ate-env/gavel-run-6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f", SubstrateActor("ate-env", run))
}
