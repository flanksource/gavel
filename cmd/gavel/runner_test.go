// Tests gavel runner exec against a fake run API: it fetches its spec and source, checks out the
// commit, runs the job's command there, posts the runs it produced and exits with its code.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/todos/remote"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunAPI struct {
	spec     sandbox.RunSpec
	bundle   []byte
	token    string
	mu       sync.Mutex
	results  []string
	messages []remote.Message
}

func (f *fakeRunAPI) handler(run uuid.UUID) http.Handler {
	mux := http.NewServeMux()
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer "+f.token {
				http.Error(w, `{"error":"bad token"}`, http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
	prefix := "/api/runs/" + run.String()
	mux.HandleFunc("GET "+prefix+"/spec", auth(func(w http.ResponseWriter, _ *http.Request) { _ = json.NewEncoder(w).Encode(f.spec) }))
	mux.HandleFunc("GET "+prefix+"/source", auth(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(f.bundle) }))
	mux.HandleFunc("POST "+prefix+"/results", auth(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.results = append(f.results, string(data))
		f.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"runId":"run-x"}`))
	}))
	mux.HandleFunc("POST "+prefix+"/messages", auth(func(w http.ResponseWriter, r *http.Request) {
		var message remote.Message
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			http.Error(w, `{"error":"bad message"}`, http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		f.messages = append(f.messages, message)
		f.mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	return mux
}

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}

// sourceBundle commits two versions of README.md and returns a bundle of main and the first commit.
func sourceBundle(t *testing.T, files map[string]string) ([]byte, string) {
	t.Helper()
	work := t.TempDir()
	gitIn(t, work, "init", "-q", "-b", "main")
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(work, name), []byte(content), 0o644))
	}
	gitIn(t, work, "add", ".")
	gitIn(t, work, "commit", "-q", "-m", "first")
	first := gitIn(t, work, "rev-parse", "HEAD")
	require.NoError(t, os.WriteFile(filepath.Join(work, "README.md"), []byte("second"), 0o644))
	gitIn(t, work, "commit", "-q", "-am", "second")
	bundle := filepath.Join(t.TempDir(), "src.bundle")
	gitIn(t, work, "bundle", "create", bundle, "refs/heads/main")
	data, err := os.ReadFile(bundle)
	require.NoError(t, err)
	return data, first
}

func runnerEnv(server, token string, run uuid.UUID) func(string) string {
	env := map[string]string{"GAVEL_SERVER": server, "GAVEL_RUN_TOKEN": token, "GAVEL_RUN_ID": run.String(), "GAVEL_DB": "off"}
	return func(key string) string { return env[key] }
}

func TestRunnerExecRunsTheJobAndPostsItsResults(t *testing.T) {
	run := uuid.New()
	bundle, first := sourceBundle(t, map[string]string{"README.md": "first"})
	api := &fakeRunAPI{spec: sandbox.RunSpec{Kind: "test", Args: []string{"./..."}, Ref: "refs/heads/main", SHA: first}, bundle: bundle, token: "run-token"}
	server := httptest.NewServer(api.handler(run))
	t.Cleanup(server.Close)
	workspace := t.TempDir()

	var ran []string
	command := func(_ context.Context, dir string, name string, args []string, _ io.Writer) (int, error) {
		readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
		require.NoError(t, err)
		assert.Equal(t, "first", string(readme), "the spec's commit is checked out")
		ran = append([]string{name}, args...)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".gavel"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel", "run-2026-10-15T09-00-00Z.json"), []byte(`{"tests":[]}`), 0o644))
		return 1, nil
	}
	code, err := runnerExec(context.Background(), runnerOptions{Run: run.String(), Workspace: workspace, Getenv: runnerEnv(server.URL, "run-token", run), Command: command, Output: io.Discard})
	require.NoError(t, err)
	assert.Equal(t, 1, code, "the job's exit code")
	gavel, _ := os.Executable()
	assert.Equal(t, []string{gavel, "test", "--work-dir", filepath.Join(workspace, "src"), "./..."}, ran)
	assert.Equal(t, []string{`{"tests":[]}`}, api.results)
}

func TestRunnerExecPushCheckRunsTheRepositorysSSHCommand(t *testing.T) {
	run := uuid.New()
	bundle, _ := sourceBundle(t, map[string]string{"README.md": "first", ".gavel.yaml": "ssh:\n  cmd: make check\n"})
	api := &fakeRunAPI{spec: sandbox.RunSpec{Kind: "push-check", Ref: "refs/heads/main"}, bundle: bundle, token: "run-token"}
	server := httptest.NewServer(api.handler(run))
	t.Cleanup(server.Close)

	var ran []string
	command := func(_ context.Context, _ string, name string, args []string, _ io.Writer) (int, error) {
		ran = append([]string{name}, args...)
		return 0, nil
	}
	code, err := runnerExec(context.Background(), runnerOptions{Run: run.String(), Workspace: t.TempDir(), Getenv: runnerEnv(server.URL, "run-token", run), Command: command, Output: io.Discard})
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, []string{"sh", "-c", "make check"}, ran)
}

func TestRunnerExecRefusesAnUnsafeEnvironment(t *testing.T) {
	run := uuid.New()
	base := map[string]string{"GAVEL_SERVER": "https://gavel.example", "GAVEL_RUN_TOKEN": "t", "GAVEL_RUN_ID": run.String(), "GAVEL_DB": "off"}
	for name, tc := range map[string]struct {
		change func(map[string]string)
		run    string
		want   string
	}{
		"a database":       {func(env map[string]string) { env["GAVEL_DB"] = "" }, run.String(), "gavel runner exec needs GAVEL_DB=off: a sandbox job never reaches a database"},
		"no token":         {func(env map[string]string) { delete(env, "GAVEL_RUN_TOKEN") }, run.String(), "GAVEL_SERVER, GAVEL_RUN_TOKEN and GAVEL_RUN_ID must be set"},
		"another run's id": {func(map[string]string) {}, uuid.NewString(), "--run "},
	} {
		t.Run(name, func(t *testing.T) {
			env := map[string]string{}
			for key, value := range base {
				env[key] = value
			}
			tc.change(env)
			_, err := runnerExec(context.Background(), runnerOptions{Run: tc.run, Workspace: t.TempDir(), Getenv: func(key string) string { return env[key] }, Output: io.Discard})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestRunnerExecReportsAnAPIError(t *testing.T) {
	run := uuid.New()
	api := &fakeRunAPI{token: "right"}
	server := httptest.NewServer(api.handler(run))
	t.Cleanup(server.Close)
	_, err := runnerExec(context.Background(), runnerOptions{Run: run.String(), Workspace: t.TempDir(), Getenv: runnerEnv(server.URL, "wrong", run), Output: io.Discard})
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("get the spec of run %s: HTTP 401: bad token", run))
}

// scriptedRunnerAgent answers every turn with one scripted response in place of an agent CLI.
type scriptedRunnerAgent struct{ text string }

func (a scriptedRunnerAgent) GetModel() string { return "claude-sonnet-5" }
func (a scriptedRunnerAgent) GetRuntime() api.Runtime {
	return api.RuntimeOf(api.Anthropic, api.ModeAgent)
}
func (a scriptedRunnerAgent) Execute(context.Context, api.Spec) (*api.Response, error) {
	return nil, fmt.Errorf("the scripted agent only streams")
}
func (a scriptedRunnerAgent) ExecuteStream(context.Context, api.Spec) (<-chan captainai.Event, error) {
	ch := make(chan captainai.Event, 2)
	ch <- captainai.Event{Kind: captainai.EventText, Text: a.text}
	ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
	close(ch)
	return ch, nil
}

func TestRunnerExecRunsAnAgentJobAndReportsIt(t *testing.T) {
	run := uuid.New()
	bundle, first := sourceBundle(t, map[string]string{"README.md": "first"})
	spec := api.Spec{Model: api.Model{Name: "claude-sonnet-5", Provider: api.Anthropic, Mode: api.ModeAgent}, Prompt: api.Prompt{User: "fix the build"}}
	spec.SetCwd("/srv/gavel/workspaces/org/repo")
	job, err := json.Marshal(remote.Job{Resolved: api.ResolvedSpec{Spec: spec}, Config: remote.JobConfig{Model: spec.Model}, Timeout: time.Minute, Root: "/srv/gavel/workspaces/org/repo"})
	require.NoError(t, err)
	fake := &fakeRunAPI{spec: sandbox.RunSpec{Kind: "agent", Ref: "refs/heads/main", SHA: first, Agent: job}, bundle: bundle, token: "run-token"}
	server := httptest.NewServer(fake.handler(run))
	t.Cleanup(server.Close)
	workspace := t.TempDir()

	code, err := runnerExec(context.Background(), runnerOptions{
		Run: run.String(), Workspace: workspace, Getenv: runnerEnv(server.URL, "run-token", run), Output: io.Discard,
		Agent: scriptedRunnerAgent{text: "done"},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, code)

	require.NotEmpty(t, fake.messages)
	assert.Equal(t, remote.KindStarted, fake.messages[0].Kind)
	assert.Equal(t, filepath.Join(workspace, "src"), fake.messages[0].Spec.Cwd(), "the agent works in the job's checkout")
	last := fake.messages[len(fake.messages)-1]
	require.Equal(t, remote.KindResult, last.Kind)
	assert.Empty(t, last.Error)
	assert.Equal(t, "done", last.Result.Unwrap().Response.Text)
}
