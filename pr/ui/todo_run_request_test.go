package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

func TestTodoAPIRunStartsSelectedTodo(t *testing.T) {
	workDir := t.TempDir()
	configureAutomaticPlanToolPolicies(t, workDir)
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:    "Run me",
		Priority: types.PriorityMedium,
		Status:   types.StatusPending,
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

	body, _ := json.Marshal(todoRunPayload{
		Ref: todos.TODOReference(created),
		Spec: api.Spec{
			Model:  api.Model{Name: "codex", Mode: api.ModeCmux, Effort: "high"},
			Budget: api.Budget{Cost: 1.25, MaxTurns: 12, Timeout: "45m"},
			// Dirty-worktree now rides the spec's
			// Setup.Checkout.Worktree.Uncommitted (Workspace section) instead of a
			// sibling flag.
			Setup: &shell.Setup{Checkout: &shell.Checkout{
				Worktree: &shell.Worktree{Mode: shell.WorktreeNew, Uncommitted: shell.CloneClone},
			}},
		},
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if resp.Status != "started" || resp.Provider != "openai" || resp.RuntimeMode != "cmux" {
		t.Fatalf("unexpected run response: %+v", resp)
	}
	if resp.Count != 1 {
		t.Fatalf("run count = %d, want 1", resp.Count)
	}
	if got.Todo == nil || got.Todo.Title != "Run me" {
		t.Fatalf("run starter did not receive selected todo: %+v", got.Todo)
	}
	if got.Dir != workDir {
		t.Fatalf("unexpected run source: dir=%q", got.Dir)
	}
	// The request layer carries exactly what the dialog sent; the fold — model
	// expansion, the lifecycle step's own layers — happens when the run resolves.
	request := got.Options.Request
	if request.Name != "codex" || request.Mode != captainai.ModeCmux || request.Effort != "high" || request.Budget.Cost != 1.25 || request.Budget.MaxTurns != 12 || !specDirty(request) {
		t.Fatalf("unexpected request layer: %+v", got.Options)
	}
	if got.Options.Host != lifecycle.HostDashboard || got.Options.Step != "" {
		t.Fatalf("run options = %+v, want the dashboard host and the lifecycle's own step choice", got.Options)
	}
	if resp.Step == "" || resp.Model != resolvedName(t, api.Model{Name: "codex", Mode: captainai.ModeCmux}) {
		t.Fatalf("response did not report the resolved step and model: %+v", resp)
	}
}

// dashboardRunSpec resolves a dashboard payload for a fresh todo in dir the way
// /api/todos/run does — wire validation, then the lifecycle host's fold — and
// returns the spec captain would be handed.
func dashboardRunSpec(t *testing.T, dir string, payload todoRunPayload) (api.Spec, error) {
	t.Helper()
	provider := uiTestProviderFor(dir)
	todo, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Resolvable", Status: types.StatusPending})
	if err != nil {
		t.Fatalf("seed todo: %v", err)
	}
	opts, err := buildTodoRunOptions(payload, nil)
	if err != nil {
		return api.Spec{}, err
	}
	prepared, err := run.Resolve(t.Context(), todoRunRequest{Provider: provider, Todo: todo, Dir: dir, Options: opts})
	if err != nil {
		return api.Spec{}, err
	}
	return prepared.Resolution.Spec, nil
}

func TestTodoAPIRunPreviewReturnsPrompt(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:    "Fix the parser",
		Body:     "The parser drops trailing commas.",
		Priority: types.PriorityMedium,
		Status:   types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	ref := todos.TODOReference(created)

	preview := func(payload todoRunPayload) todoRunPreviewResponse {
		t.Helper()
		body, _ := json.Marshal(payload)
		rec := httptest.NewRecorder()
		s.handleTodoRunPreview(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
		if rec.Code != http.StatusOK {
			t.Fatalf("preview status = %d, want 200; body = %q", rec.Code, rec.Body.String())
		}
		var resp todoRunPreviewResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal preview response: %v", err)
		}
		return resp
	}

	cmuxResp := preview(todoRunPayload{Ref: ref, Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "high"}}})
	if cmuxResp.Count != 1 || cmuxResp.Provider != "anthropic" || cmuxResp.RuntimeMode != "cmux" {
		t.Fatalf("unexpected preview meta: %+v", cmuxResp)
	}
	if !strings.Contains(cmuxResp.Prompt, "## Fix the parser") {
		t.Fatalf("cmux preview should contain the title heading: %q", cmuxResp.Prompt)
	}
	if !strings.Contains(cmuxResp.Prompt, "The parser drops trailing commas.") {
		t.Fatalf("cmux preview should inline the body: %q", cmuxResp.Prompt)
	}
	if !strings.Contains(cmuxResp.Prompt, "Think hard and reason thoroughly") {
		t.Fatalf("cmux preview should include the effort directive: %q", cmuxResp.Prompt)
	}
	if !strings.Contains(cmuxResp.Prompt, "## Instructions") {
		t.Fatalf("cmux preview should include the instructions section: %q", cmuxResp.Prompt)
	}

	planResp := preview(todoRunPayload{Ref: ref, Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}}, Step: "plan"})
	if !strings.Contains(planResp.Prompt, "## Fix the parser") {
		t.Fatalf("plan preview should contain the title: %q", planResp.Prompt)
	}
	if planResp.Step != "plan" {
		t.Fatalf("plan preview step = %q, want plan", planResp.Step)
	}

	agentResp := preview(todoRunPayload{Ref: ref, Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent, Effort: "medium"}}})
	if strings.HasPrefix(agentResp.Prompt, "# Fix the parser") {
		t.Fatalf("agent preview should be the bare claude prompt, not the cmux instruction: %q", agentResp.Prompt)
	}
	if !strings.Contains(agentResp.Prompt, "The parser drops trailing commas.") {
		t.Fatalf("agent preview should include the body: %q", agentResp.Prompt)
	}
}

func TestTodoRunPreviewAbsolutizesAttachmentURLs(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Screenshot todo",
		Body:   "See the bug:\n\n![screen.png](" + attachmentURLPrefix + "abc.png)",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	ref := todos.TODOReference(created)

	body, _ := json.Marshal(todoRunPayload{Ref: ref, Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent}}})
	rec := httptest.NewRecorder()
	s.handleTodoRunPreview(rec, httptest.NewRequest(http.MethodPost, "http://gavel.example:9092/api/todos/run/preview", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunPreviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal preview response: %v", err)
	}

	wantURL := "http://gavel.example:9092" + attachmentURLPrefix + "abc.png"
	if !strings.Contains(resp.Prompt, wantURL) {
		t.Fatalf("prompt should reference the absolute attachment URL %q: %q", wantURL, resp.Prompt)
	}
	if strings.Contains(resp.Prompt, "]("+attachmentURLPrefix) {
		t.Fatalf("prompt should not retain relative attachment links: %q", resp.Prompt)
	}
}
