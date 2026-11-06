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

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRunAPI struct {
	spec    sandbox.RunSpec
	bundle  []byte
	token   string
	mu      sync.Mutex
	results []string
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

// sourceCheckout is a checkout, on a branch of its own, of the first of two commits of README.md,
// as captain's substrate sandbox checks a job's commit out, and returns it and the commit.
func sourceCheckout(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("first"), 0o644))
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-q", "-m", "first")
	first := gitIn(t, dir, "rev-parse", "HEAD")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("second"), 0o644))
	gitIn(t, dir, "commit", "-q", "-am", "second")
	gitIn(t, dir, "switch", "-q", "-c", "substrate-run", first)
	return dir, first
}

func TestRunnerExecClonesItsSourceFromALocalCheckout(t *testing.T) {
	for name, spec := range map[string]func(first string) sandbox.RunSpec{
		"at the spec's commit": func(first string) sandbox.RunSpec {
			return sandbox.RunSpec{Kind: "test", Ref: "refs/heads/main", SHA: first}
		},
		"at the checkout's head without one": func(string) sandbox.RunSpec { return sandbox.RunSpec{Kind: "test", Ref: "refs/heads/main"} },
	} {
		t.Run(name, func(t *testing.T) {
			run := uuid.New()
			source, first := sourceCheckout(t)
			api := &fakeRunAPI{spec: spec(first), token: "run-token"}
			handler := api.handler(run)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/source") {
					t.Errorf("runner exec downloaded its source, which the local checkout has")
				}
				handler.ServeHTTP(w, r)
			}))
			t.Cleanup(server.Close)
			workspace := t.TempDir()

			command := func(_ context.Context, dir string, _ string, _ []string, _ io.Writer) (int, error) {
				assert.Equal(t, filepath.Join(workspace, "src"), dir, "the job runs in a clone of its own, so what it writes stays out of the checkout")
				readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
				require.NoError(t, err)
				assert.Equal(t, "first", string(readme), "the job's commit is checked out")
				return 0, nil
			}
			code, err := runnerExec(context.Background(), runnerOptions{
				Run: run.String(), Workspace: workspace, Source: source, Getenv: runnerEnv(server.URL, "run-token", run), Command: command, Output: io.Discard,
			})
			require.NoError(t, err)
			assert.Equal(t, 0, code)
			assert.Equal(t, "", gitIn(t, source, "status", "--porcelain"), "the checkout is left as it was")
		})
	}
}

func TestRunnerExecPushCheckRunsTheRepositorysSSHCommand(t *testing.T) {
	// A clone of the bundle, which has no HEAD, makes a local branch only for the default branch
	// name git guesses; one other than the ref's leaves the ref only under refs/remotes.
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "init.defaultBranch")
	t.Setenv("GIT_CONFIG_VALUE_0", "trunk")
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

func TestRunnerExecRefusesAnAgentRun(t *testing.T) {
	run := uuid.New()
	bundle, first := sourceBundle(t, map[string]string{"README.md": "first"})
	fake := &fakeRunAPI{spec: sandbox.RunSpec{Kind: "agent", Ref: "refs/heads/main", SHA: first}, bundle: bundle, token: "run-token"}
	server := httptest.NewServer(fake.handler(run))
	t.Cleanup(server.Close)

	_, err := runnerExec(context.Background(), runnerOptions{
		Run: run.String(), Workspace: t.TempDir(), Getenv: runnerEnv(server.URL, "run-token", run), Output: io.Discard,
	})
	require.EqualError(t, err, "an agent run runs whole on substrate, through `gavel runner run-task`, not as a sandbox job")
}
