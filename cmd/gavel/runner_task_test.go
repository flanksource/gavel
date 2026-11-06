// Tests `gavel runner run-task`, gavel's program in a substrate environment: it rebuilds gavel's
// hooks from the run's host data and runs the step with them, as a local step runs.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/captain/pkg/promptrun/remoterun"
	"github.com/flanksource/captain/pkg/sandbox/substraterun"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/sandbox/runclient"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// substrateTaskDir is a run's directory as the substrate sandbox leaves it in an environment:
// the caller's checkout in the workspace bundle, and the task, carrying hooks as host data.
func substrateTaskDir(t *testing.T, hooks lifecycle.RunHooks) string {
	t.Helper()
	return substrateTaskDirFor(t, func(string, string) lifecycle.RunHooks { return hooks })
}

// substrateTaskDirFor is substrateTaskDir with the hooks hooksFor gives for the caller's checkout
// repo at commit head.
func substrateTaskDirFor(t *testing.T, hooksFor func(repo, head string) lifecycle.RunHooks) string {
	t.Helper()
	repo := t.TempDir()
	gitIn(t, repo, "init", "-q")
	gitIn(t, repo, "commit", "-q", "--allow-empty", "-m", "first")
	hooks := hooksFor(repo, gitIn(t, repo, "rev-parse", "HEAD"))
	const ref = "refs/captain/substrate/0123456789ab"
	gitIn(t, repo, "update-ref", ref, "HEAD")
	dir := t.TempDir()
	gitIn(t, repo, "bundle", "create", "-q", filepath.Join(dir, substraterun.WorkspaceBundle), ref)

	spec := api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}, Prompt: api.Prompt{User: "fix the build"}}
	spec.SetCwd(repo)
	host, err := json.Marshal(hooks)
	require.NoError(t, err)
	job := remoterun.JobFor(promptrun.Input{Resolved: api.ResolvedSpec{Spec: spec}, Timeout: time.Minute}, host)
	job.Root = repo
	task, err := json.Marshal(substraterun.Task{Job: job, Ref: ref})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, substraterun.TaskFile), task, 0o644))
	return dir
}

// recordingAgent answers every turn with done, and keeps the request it was sent.
type recordingAgent struct{ sent chan api.Spec }

func (a recordingAgent) GetModel() string        { return "claude-sonnet-5" }
func (a recordingAgent) GetRuntime() api.Runtime { return api.RuntimeOf(api.Anthropic, api.ModeAgent) }
func (a recordingAgent) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, fmt.Errorf("the recording agent only streams")
}
func (a recordingAgent) ExecuteStream(_ context.Context, req api.Spec) (<-chan captainai.Event, error) {
	a.sent <- req
	ch := make(chan captainai.Event, 2)
	ch <- captainai.Event{Kind: captainai.EventText, Text: "done"}
	ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
	close(ch)
	return ch, nil
}

func TestRunTaskRunsTheStepWithGavelsHooks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := substrateTaskDir(t, lifecycle.RunHooks{
		Env:    map[string]string{"GAVEL_ISSUE_ID": "todo-1"},
		Labels: map[string]string{"todo": "todo-1"},
	})
	agent := recordingAgent{sent: make(chan api.Spec, 1)}
	stdoutR, stdoutW := io.Pipe()
	stdinR, stdinW := io.Pipe()
	ran := make(chan error, 1)
	go func() {
		ran <- runnerRunTask(context.Background(), runTaskOptions{Dir: dir, Agent: agent}, stdoutW, stdinR)
		_ = stdoutW.Close()
	}()
	var started []api.Spec
	result, err := remoterun.Relay(t.Context(), promptrun.Input{}, func(spec api.Spec) error {
		started = append(started, spec)
		return nil
	}, stdoutR, stdinW)
	_ = stdinW.Close()
	require.NoError(t, err)
	require.NoError(t, <-ran)

	assert.Equal(t, "done", result.Response.Text)
	require.Len(t, started, 1)
	assert.Equal(t, "todo-1", started[0].Labels["todo"], "gavel's labels hook ran in the environment")
	sent := <-agent.sent
	require.NotNil(t, sent.Setup)
	assert.Contains(t, sent.Setup.Env, "GAVEL_ISSUE_ID=todo-1", "and its run environment hook, whose variables the agent gets")
}

func TestRunTaskRefusesHostDataThatIsNotGavelsHooks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := substrateTaskDir(t, lifecycle.RunHooks{})
	task := filepath.Join(dir, substraterun.TaskFile)
	data, err := os.ReadFile(task)
	require.NoError(t, err)
	var decoded substraterun.Task
	require.NoError(t, json.Unmarshal(data, &decoded))
	decoded.Job.Host = json.RawMessage(`["not", "hooks"]`)
	data, err = json.Marshal(decoded)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(task, data, 0o644))

	stdoutR, stdoutW := io.Pipe()
	stdinR, stdinW := io.Pipe()
	ran := make(chan error, 1)
	go func() {
		ran <- runnerRunTask(context.Background(), runTaskOptions{Dir: dir, Agent: scriptedRunnerAgent{text: "done"}}, stdoutW, stdinR)
		_ = stdoutW.Close()
	}()
	_, err = remoterun.Relay(t.Context(), promptrun.Input{}, func(api.Spec) error { return nil }, stdoutR, stdinW)
	_ = stdinW.Close()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode gavel's hooks from the run's host data: ")
	require.Error(t, <-ran)
}

func TestRunTaskReachesTheStepsPRThroughItsCaller(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var head string
	dir := substrateTaskDirFor(t, func(repo, sha string) lifecycle.RunHooks {
		head = sha
		return lifecycle.RunHooks{PR: &lifecycle.PRHooks{
			URL: "https://github.com/acme/app/pull/7", Number: 7, WorkDir: repo,
			Target: provider.RunTarget{RepositoryURL: "https://github.com/acme/app.git", Branch: "topic", HeadSHA: sha},
		}}
	})
	// The caller's run API, as serve's answers a run's PR endpoints.
	var asked []string
	runAPI := http.NewServeMux()
	runAPI.HandleFunc("GET /api/runs/run-1/pr/status", func(w http.ResponseWriter, r *http.Request) {
		asked = append(asked, "status")
		_ = json.NewEncoder(w).Encode(model.StatusSnapshot{PR: &model.PRInfo{
			State: "OPEN", HeadRefOID: head, HeadRefName: "topic", HeadRepositoryURL: "https://github.com/acme/app.git",
		}})
	})
	runAPI.HandleFunc("GET /api/runs/run-1/pr/head", func(w http.ResponseWriter, r *http.Request) {
		asked = append(asked, "head")
		_ = json.NewEncoder(w).Encode(map[string]string{"sha": head})
	})
	stdoutR, stdoutW := io.Pipe()
	stdinR, stdinW := io.Pipe()
	ran := make(chan error, 1)
	go func() {
		ran <- runnerRunTask(context.Background(), runTaskOptions{Dir: dir, Agent: scriptedRunnerAgent{text: "done"}}, stdoutW, stdinR)
		_ = stdoutW.Close()
	}()
	in := promptrun.Input{HostCalls: map[string]func(context.Context, json.RawMessage) (json.RawMessage, error){
		runclient.HostCallName: runclient.AnswerHostCall(runAPI, "run-1", "run-token"),
	}}
	_, err := remoterun.Relay(t.Context(), in, func(api.Spec) error { return nil }, stdoutR, stdinW)
	_ = stdinW.Close()
	require.NoError(t, err)
	require.NoError(t, <-ran)
	assert.Equal(t, []string{"status", "head"}, asked, "the PR hook read its PR through the caller, holding no credentials")
	assert.False(t, provider.InJob(), "the process's run API is the run's alone")
}
