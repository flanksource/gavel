package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// isolatedTodoWorkspace returns a workspace whose .gavel.yaml layers are empty
// apart from a model, so a resolution test reads only what it declares. The home
// layer is already redirected package-wide by TestMain.
//
// The model is not optional any more: gavel has no built-in default, so a run
// resolved from a genuinely empty workspace fails with "model name is required"
// before it reaches whatever the test is actually about. Declaring it here keeps
// these tests about run options rather than about configuration.
func isolatedTodoWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("ai:\n  model: agent:claude-sonnet-5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// Auto-commit is sourced from Workflow.Commits. The lifecycle's run step
// declares one, so a default dashboard run commits; a payload stanza replaces
// it; and an empty list cannot clear it — captain's merge reads an empty slice
// as "not stated", which is why the CLI refuses `--commit=false` outright
// rather than pretending the request layer removed anything.
func TestDashboardRunCommitFromWorkflow(t *testing.T) {
	dir := isolatedTodoWorkspace(t)
	// The run step is named: a pending todo with no plan would otherwise be
	// planned first, and a plan never commits.
	base := todoRunPayload{Step: "run", Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}}}

	cases := []struct {
		name     string
		workflow *api.Workflow
		want     []api.Commit
	}{
		{"no workflow inherits the lifecycle step's commit", nil, []api.Commit{{On: api.CommitOnRun, Stage: "worktree", Gates: api.CommitGatesFull}}},
		{"a commit policy replaces it", &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun, Gates: api.CommitGatesCheap}}}, []api.Commit{{On: api.CommitOnRun, Gates: api.CommitGatesCheap}}},
		{"an empty commit list leaves the step's commit", &api.Workflow{Commits: []api.Commit{}}, []api.Commit{{On: api.CommitOnRun, Stage: "worktree", Gates: api.CommitGatesFull}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := base
			payload.Spec.Workflow = tc.workflow
			spec, err := dashboardRunSpec(t, dir, payload)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			if !reflect.DeepEqual(run.Commits(spec), tc.want) {
				t.Fatalf("commits = %+v, want %+v", run.Commits(spec), tc.want)
			}
		})
	}
}

func TestDashboardRunToolPreferences(t *testing.T) {
	dir := isolatedTodoWorkspace(t)
	base := todoRunPayload{Step: "run", Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}}}

	t.Run("valid prefs and permission mode are threaded", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{
			Mode: api.PermissionDefault,
			Tools: api.Tools{
				"Bash": api.ToolPolicyAllow, "Write": api.ToolPolicyDeny, "Read": api.ToolPolicyAuto, "Glob": api.ToolPolicyAuto,
			},
		}
		spec, err := dashboardRunSpec(t, dir, payload)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if spec.Permissions.Mode != api.PermissionDefault {
			t.Fatalf("PermissionMode = %q, want default", spec.Permissions.Mode)
		}
		policies := spec.Permissions.Tools.Policies()
		if policies["Bash"] != api.ToolPolicyAllow || policies["Write"] != api.ToolPolicyDeny || policies["Read"] != api.ToolPolicyAuto {
			t.Fatalf("tool policies = %v, want Bash=allow Write=deny Read=auto", policies)
		}
		// auto is carried rather than dropped now that Tools is the policy map.
		// It means the same thing either way — an absent tool inherits the posture
		// and auto defers to it — but the map now says what was configured.
		if policies["Glob"] != api.ToolPolicyAuto {
			t.Fatalf("tool policies = %v, want Glob carried as auto", policies)
		}
	})

	t.Run("empty prefs resolve to the dashboard posture", func(t *testing.T) {
		spec, err := dashboardRunSpec(t, dir, base)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		// The dashboard resolves as the approval-serving host, so a payload that
		// states no posture still leaves with permissions.mode: default — the mode
		// that makes the broker the thing a tool call is checked against. The
		// prompt's own preset may allow the read-only tools; nothing here polices
		// Bash, which is what the broker exists to ask about.
		if policy := spec.Permissions.Tools.Policies()["Bash"]; policy != "" || spec.Permissions.Mode != api.PermissionDefault {
			t.Fatalf("want an unpoliced Bash and the dashboard posture, got %v / %q", spec.Permissions.Tools, spec.Permissions.Mode)
		}
	})

	t.Run("invalid tool policy fails at the wire", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{Tools: api.Tools{"Bash": "maybe"}}
		if _, err := buildTodoRunOptions(payload, nil); err == nil {
			t.Fatal("expected error for invalid tool policy, got nil")
		}
	})

	t.Run("invalid permission mode fails at the wire", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{Mode: "yolo"}
		if _, err := buildTodoRunOptions(payload, nil); err == nil {
			t.Fatal("expected error for invalid permission mode, got nil")
		}
	})
}

// api.Spec declares a value-receiver MarshalJSON; when todoRunPayload embedded it
// that method was promoted onto the payload, so marshaling emitted a bare spec and
// every sibling field vanished — silently emptying the review API's `options`
// object and stripping the ref off every request body a test builds. Spec is now a
// named field under its own `spec` key, and this asserts the whole payload survives
// the wire in both directions.
func TestTodoRunPayloadRoundTripsSpecAndSiblings(t *testing.T) {
	payload := todoRunPayload{
		Dir:    "/repos/gavel",
		Ref:    "todo-1",
		Refs:   []string{"todo-1", "todo-2"},
		Step:   "plan",
		Resume: true,
		Force:  true,
		Spec: api.Spec{
			Model:  api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium", Fallbacks: api.ModelList{api.Model{Name: "claude-sonnet-5"}.WithExplicit("/model")}},
			Prompt: api.Prompt{User: "Implement the reviewed plan.", System: "Keep the patch narrow."},
			Budget: api.Budget{Cost: 2.5, MaxTurns: 8, Timeout: "20m"},
			Memory: api.Memory{Skills: []string{"gavel-todos"}},
			Permissions: api.Permissions{
				Mode:    api.PermissionAcceptEdits,
				Tools:   api.Tools{"Bash": api.ToolPolicyAsk},
				MCP:     api.MCP{Servers: []string{"postgres"}},
				Plugins: api.ResourcePolicies{"review": api.ResourceEnabled},
				Skills:  api.ResourcePolicies{"gavel-todos": api.ResourceEnabled},
			},
			Setup:     &shell.Setup{Cwd: "workspace", Checkout: &shell.Checkout{Mode: shell.CheckoutLocal, Path: ".", Worktree: &shell.Worktree{Mode: shell.WorktreeNew, Prefix: "todo"}}},
			Workflow:  &api.Workflow{Verify: &api.Verify{Commands: []string{"go test ./todos"}, Scope: api.VerifyScopeChanged, MaxIterations: 3}, Commits: []api.Commit{{On: api.CommitOnRun, Gates: api.CommitGatesFull}}},
			SessionID: "sess-1",
			CLIArgs:   map[string]any{"fullAuto": true},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	var got todoRunPayload
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if !reflect.DeepEqual(payload, got) {
		t.Fatalf("payload did not round-trip\n got: %+v\nwant: %+v\nwire: %s", got, payload, body)
	}
}

// The response's Commit reports the RESOLVED run: a payload stanza commits, and
// so does a payload that says nothing, because the lifecycle's run step declares
// a commit of its own.
func TestTodoAPIRunThreadsCommitOption(t *testing.T) {
	cases := []struct {
		name     string
		workflow *api.Workflow
		want     bool
	}{
		{"a commit policy auto-commits", &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun}}}, true},
		{"absent workflow inherits the run step's commit", nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workDir := t.TempDir()
			s := &Server{ghOpts: github.Options{WorkDir: workDir}}
			created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
				Title:  "Run me",
				Status: types.StatusPending,
			})
			if err != nil {
				t.Fatalf("seed create: %v", err)
			}

			oldStart := run.Start
			var got todoRunRequest
			run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
				got = req
				return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
			}
			t.Cleanup(func() { run.Start = oldStart })

			payload := todoRunPayload{
				Ref:  todos.TODOReference(created),
				Step: "run",
				Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}},
			}
			payload.Spec.Workflow = tc.workflow
			body, _ := json.Marshal(payload)
			rec := httptest.NewRecorder()
			s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
			if rec.Code != http.StatusOK {
				t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
			}
			var resp todoRunResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal run response: %v", err)
			}
			if resp.Commit != tc.want {
				t.Fatalf("response Commit = %v, want %v", resp.Commit, tc.want)
			}
			if got := specCommit(resolvedRun(t, got).Spec); got != tc.want {
				t.Fatalf("run starter commit = %v, want %v", got, tc.want)
			}
		})
	}
}

// A dry run (every Workflow.Commits stanza marked dryRun) still executes the
// agent — it only reports what it would commit — so handleTodoRun starts it and
// reports Commit:false.
func TestTodoAPIRunDryRunStartsButSkipsCommit(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Run me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	oldStart := run.Start
	started := false
	run.Start = func(todoRunRequest) (todoRunStartResult, error) {
		started = true
		return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	payload := todoRunPayload{
		Ref:  todos.TODOReference(created),
		Step: "run",
		Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}},
	}
	// A commit policy is declared but marked dryRun: the run executes, the commit
	// is reported rather than cut.
	payload.Spec.Workflow = &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun, DryRun: true}}}
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if resp.Status != "started" {
		t.Fatalf("status = %q, want started (a dry run still executes)", resp.Status)
	}
	if resp.Commit {
		t.Fatal("response Commit = true, want false (dry run suppresses the commit)")
	}
	if !started {
		t.Fatal("dry run did not start the agent run")
	}
}

func TestTodoAPIRunRejectsMultipleTodos(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := uiTestProviderFor(workDir)
	first, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:  "First todo",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed first: %v", err)
	}
	second, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:  "Second todo",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed second: %v", err)
	}

	oldStart := run.Start
	run.Start = func(todoRunRequest) (todoRunStartResult, error) {
		t.Fatal("grouped native run must be rejected before dispatch")
		return todoRunStartResult{}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	body, _ := json.Marshal(todoRunPayload{
		Refs: []string{todos.TODOReference(first), todos.TODOReference(second), todos.TODOReference(first)},
		Spec: api.Spec{Model: api.Model{Name: "sonnet", Mode: api.ModeCmux, Effort: "medium"}},
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("run status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "one issue at a time") {
		t.Fatalf("unexpected grouped-run error: %q", rec.Body.String())
	}
}
