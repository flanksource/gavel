package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

func TestTodoAPIRunWarnsUnknownFields(t *testing.T) {
	warnings := captureTodoRequestWarnings(t)
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Run me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	got, _ := stubRunStart(t)
	for field, value := range map[string]any{"driver": "retired", "runMode": "run", "prompt": "triage", "plan": true, "unexpected": 1} {
		body, err := json.Marshal(map[string]any{"ref": todos.TODOReference(created), "step": "plan", field: value})
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()
		s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: run status = %d, want 200; body = %q", field, rec.Code, rec.Body.String())
		}
		if !strings.Contains(warnings.String(), "POST /api/todos/run request body: unknown field "+field) {
			t.Fatalf("missing %s warning: %q", field, warnings.String())
		}
		if got.Options.Step != "plan" {
			t.Fatalf("%s changed the requested step to %q", field, got.Options.Step)
		}
	}
}

func TestTodoAPIRunPlanThreadsPlanOption(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Plan me",
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

	body, _ := json.Marshal(todoRunPayload{
		Ref:  todos.TODOReference(created),
		Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}},
		Step: "plan",
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
	if resp.Step != "plan" || !strings.Contains(resp.Reason, "step") {
		t.Fatalf("response did not echo the requested step and why: %+v", resp)
	}
	if got.Options.Step != "plan" {
		t.Fatalf("run starter did not receive the plan step: %+v", got.Options)
	}
}

// The step is the whole vocabulary: a plan step runs on every runtime (the plan
// prompt's frontmatter carries the plan posture), verify is a step like any
// other and is refused only when the todo has nothing to verify, and an unknown
// step is refused by name with the lifecycle's own steps listed.
func TestTodoAPIRunStepField(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Plan me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	got, _ := stubRunStart(t)

	post := func(payload todoRunPayload) *httptest.ResponseRecorder {
		t.Helper()
		body, _ := json.Marshal(payload)
		rec := httptest.NewRecorder()
		s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
		return rec
	}

	rec := post(todoRunPayload{
		Ref:  todos.TODOReference(created),
		Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent, Effort: "medium"}},
		Step: "plan",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.Step != "plan" {
		t.Fatalf("step not threaded: %+v", got.Options)
	}

	rec = post(todoRunPayload{Ref: todos.TODOReference(created), Step: "verify"})
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "definition of done") {
		t.Fatalf("verify on a todo with nothing to verify: status = %d, body = %q, want a 400 naming the missing definition of done", rec.Code, rec.Body.String())
	}

	rec = post(todoRunPayload{Ref: todos.TODOReference(created), Step: "shape-it"})
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "shape-it") || !strings.Contains(rec.Body.String(), "plan") {
		t.Fatalf("unknown step: status = %d, body = %q, want a 400 naming the step and listing the lifecycle's", rec.Code, rec.Body.String())
	}
}

func TestDashboardRunRuntimeFromSpec(t *testing.T) {
	dir := isolatedTodoWorkspace(t)
	spec, err := dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: api.ModeAgent, Effort: "medium"}}})
	if err != nil {
		t.Fatalf("agent runtime: %v", err)
	}
	if spec.Mode != api.ModeAgent {
		t.Fatalf("got mode=%q", spec.Mode)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "codex", Mode: api.ModeCmux}}})
	if err != nil {
		t.Fatalf("codex cmux runtime: %v", err)
	}
	if providerKey(spec.Model) != "openai" || spec.Name != resolvedName(t, api.Model{Name: "codex", Mode: api.ModeCmux}) || spec.Mode != api.ModeCmux {
		t.Fatalf("got provider=%q model=%q mode=%q", providerKey(spec.Model), spec.Name, spec.Mode)
	}
}

func TestTodoRunContextListsCaptainRuntimeModes(t *testing.T) {
	prev := runCaptainWhoami
	calls := 0
	runCaptainWhoami = func(opts captaincli.WhoamiOptions) (any, error) {
		calls++
		if opts.Mode != "" || opts.Provider != "" || !opts.Models {
			t.Fatalf("whoami opts = %+v, want one unfiltered model snapshot", opts)
		}
		adapters := make([]captaincli.AdapterStatus, 0, 5)
		for _, runtimeMode := range []string{"cmux", "agent", "cli"} {
			adapters = append(adapters, captaincli.AdapterStatus{
				Provider:      "anthropic",
				Mode:          runtimeMode,
				Type:          "cli",
				Authenticated: true,
				Binary:        "/usr/local/bin/claude",
				ModelCount:    3,
				Models:        []string{"claude-sonnet-5", "claude-fable-5", "claude-opus-4-8"},
				ModelDetails: []captainai.ModelDef{
					{ID: "claude-sonnet-5", Name: "Claude Sonnet 5", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
					{ID: "claude-fable-5", Name: "Claude Fable 5", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
					{ID: "claude-opus-4-8", Name: "Claude Opus 4.8", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
				},
			})
		}
		for _, runtimeMode := range []string{"cmux", "agent"} {
			adapters = append(adapters, captaincli.AdapterStatus{
				Provider:      "openai",
				Mode:          runtimeMode,
				Type:          "cli",
				Authenticated: true,
				Binary:        "/usr/local/bin/codex",
				ModelCount:    4,
				Models:        []string{"gpt-5.5", "gpt-5.6-sol", "gpt-5.6-luna", "gpt-5.6-terra"},
				ModelDetails: []captainai.ModelDef{
					{ID: "gpt-5.5", Name: "GPT-5.5", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh}, DefaultEffort: api.EffortMedium},
					{ID: "gpt-5.6-sol", Name: "GPT-5.6 Sol", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax, api.EffortUltra}, DefaultEffort: api.EffortMedium},
					{ID: "gpt-5.6-luna", Name: "GPT-5.6 Luna", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
					{ID: "gpt-5.6-terra", Name: "GPT-5.6 Terra", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
				},
			})
		}
		return captaincli.WhoamiResult{
			Adapters:        adapters,
			Runtimes:        api.RuntimeCatalog(),
			DefaultProvider: "anthropic",
			ProviderDefaults: map[string]captaincli.ProviderDefaultView{
				"anthropic": {Mode: "agent", Model: "claude-sonnet-5"},
				"openai":    {Mode: "agent", Model: "gpt-5.5"},
			},
		}, nil
	}
	t.Cleanup(func() { runCaptainWhoami = prev })

	resp, err := todoRunContext(context.Background(), "")
	if err != nil {
		t.Fatalf("todo run context: %v", err)
	}
	if calls != 1 {
		t.Fatalf("captain whoami calls = %d, want 1", calls)
	}
	if !stringSliceContains(resp.Efforts, "xhigh") {
		t.Fatalf("efforts = %v, want captain xhigh effort", resp.Efforts)
	}
	if resp.DefaultMode != "agent" {
		t.Fatalf("default mode = %q, want agent", resp.DefaultMode)
	}
	modeFor := func(provider, id string) todoRunModeOption {
		t.Helper()
		for _, option := range resp.Modes {
			if option.Provider == provider && option.ID == id {
				return option
			}
		}
		t.Fatalf("missing %s %s runtime in %+v", provider, id, resp.Modes)
		return todoRunModeOption{}
	}
	claudeCmux := modeFor("anthropic", "cmux")
	claudeAgent := modeFor("anthropic", "agent")
	claudeCLI := modeFor("anthropic", "cli")
	codexCmux := modeFor("openai", "cmux")
	codexAgent := modeFor("openai", "agent")
	for _, option := range []todoRunModeOption{claudeCmux, claudeAgent, claudeCLI, codexCmux, codexAgent} {
		if len(option.Models) == 0 {
			t.Fatalf("runtime %s/%s has no model list: %+v", option.Provider, option.ID, option)
		}
	}
	if !todoRunModelsContain(claudeCmux.Models, "claude-sonnet-5") {
		t.Fatalf("claude cmux models = %+v, want claude-sonnet-5 from captain whoami", claudeCmux.Models)
	}
	if !todoRunModelsContain(claudeCmux.Models, "claude-opus-4-8") {
		t.Fatalf("claude cmux models = %+v, want claude-opus-4-8 from captain whoami", claudeCmux.Models)
	}
	if !todoRunModelsContain(claudeAgent.Models, "claude-fable-5") {
		t.Fatalf("claude agent models = %+v, want Fable from captain whoami", claudeAgent.Models)
	}
	if len(claudeCmux.Models) != 3 || claudeCmux.Models[0].ID != "claude-sonnet-5" || claudeCmux.Models[1].ID != "claude-fable-5" || claudeCmux.Models[2].ID != "claude-opus-4-8" {
		t.Fatalf("claude cmux models = %+v, want Captain model-detail order", claudeCmux.Models)
	}
	if todoRunModelsContain(claudeCmux.Models, "claude-agent-sonnet") {
		t.Fatalf("claude cmux models = %+v, should not expose synthetic aliases", claudeCmux.Models)
	}
	if !todoRunModelsContain(codexCmux.Models, "gpt-5.5") {
		t.Fatalf("codex cmux models = %+v, want gpt-5.5 from captain whoami", codexCmux.Models)
	}
	for _, id := range []string{"gpt-5.6-sol", "gpt-5.6-luna", "gpt-5.6-terra"} {
		if !todoRunModelsContain(codexAgent.Models, id) {
			t.Fatalf("codex agent models = %+v, want %s from captain whoami", codexAgent.Models, id)
		}
	}
	sol := todoRunModelByID(codexAgent.Models, "gpt-5.6-sol")
	if !sol.CapabilitiesKnown || !sol.Reasoning || sol.DefaultEffort != "medium" || !stringSliceContains(sol.SupportedEfforts, "ultra") || sol.Temperature == nil || *sol.Temperature {
		t.Fatalf("gpt-5.6-sol capabilities = %+v, want exact effort and temperature metadata", sol)
	}
	if todoRunModelsContain(codexCmux.Models, "gpt-5-codex") {
		t.Fatalf("codex cmux models = %+v, should not include code variant gpt-5-codex", codexCmux.Models)
	}
}

// TestNormalizeTodoRunOptionsCaptainRuntime pins the (provider, mode) each
// payload normalizes to. The provider follows from the model name, so it is
// asserted through providerKey rather than read off the spec.
func TestDashboardRunCaptainRuntime(t *testing.T) {
	dir := isolatedTodoWorkspace(t)
	spec, err := dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: api.ModeCLI, Effort: "xhigh"}}})
	if err != nil {
		t.Fatalf("claude cli runtime: %v", err)
	}
	if spec.Mode != api.ModeCLI || providerKey(spec.Model) != "anthropic" || spec.Effort != "xhigh" {
		t.Fatalf("unexpected claude cli spec: %+v", spec.Model)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "codex", Mode: api.ModeAgent}}})
	if err != nil {
		t.Fatalf("default codex runtime: %v", err)
	}
	if spec.Mode != api.ModeAgent || providerKey(spec.Model) != "openai" {
		t.Fatalf("unexpected codex agent defaults: %+v", spec.Model)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: api.ModeCLI, Name: "gpt-5.5"}}})
	if err != nil {
		t.Fatalf("the model's mode should override the driver: %v", err)
	}
	if spec.Mode != api.ModeCLI || providerKey(spec.Model) != "openai" {
		t.Fatalf("the model's mode did not take precedence: %+v", spec.Model)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "sonnet-4-6", Mode: api.ModeCmux}}})
	if err != nil {
		t.Fatalf("versioned claude cmux model: %v", err)
	}
	if spec.Mode != api.ModeCmux || spec.Name != "claude-sonnet-4-6" {
		t.Fatalf("unexpected versioned claude model spec: %+v", spec.Model)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: api.ModeAgent, Name: "opus-4-8"}}})
	if err != nil {
		t.Fatalf("claude agent opus model: %v", err)
	}
	if spec.Mode != api.ModeAgent || spec.Name != "claude-opus-4-8" {
		t.Fatalf("unexpected normalized opus model spec: %+v", spec.Model)
	}

	if _, err := dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: "claude-agent", Name: "claude-sonnet-5"}}}); err == nil {
		t.Fatal("a composite adapter id in the mode field should be rejected")
	}
}

func TestTodoAPIRunThreadsCaptainRuntimeMode(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Run headless",
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

	body, _ := json.Marshal(todoRunPayload{
		Ref:  todos.TODOReference(created),
		Spec: api.Spec{Model: api.Model{Mode: api.ModeAgent, Name: "claude-sonnet-5", Effort: "medium"}},
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
	if resp.RuntimeMode != "agent" || resp.Model != "claude-sonnet-5" {
		t.Fatalf("response did not thread mode/model: %+v", resp)
	}
	if got.Options.Request.Mode != api.ModeAgent || got.Options.Request.Name != "claude-sonnet-5" {
		t.Fatalf("run starter did not receive mode/model: %+v", got.Options)
	}
}

func TestTodoRunModelLabelFormatsVersionedClaudeModels(t *testing.T) {
	cases := map[string]string{
		"sonnet-5":              "Sonnet 5",
		"sonnet-4-6":            "Sonnet 4.6",
		"opus-4-8":              "Opus 4.8",
		"claude-agent-opus-4-6": "Opus 4.6",
		"claude-sonnet-5":       "Sonnet 5",
		"claude-agent-sonnet":   "Sonnet",
		"gpt-5-codex":           "GPT 5 Codex",
		"codex-gpt-5-codex":     "GPT 5 Codex",
		"claude-code-haiku-4-5": "Haiku 4.5",
		"claude-agent-fable-5":  "Fable 5",
	}
	for in, want := range cases {
		if got := modelLabel(in); got != want {
			t.Errorf("modelLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func stringSliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func todoRunModelsContain(items []todoRunModelOption, want string) bool {
	for _, item := range items {
		if item.ID == want {
			return true
		}
	}
	return false
}

func todoRunModelByID(items []todoRunModelOption, want string) todoRunModelOption {
	for _, item := range items {
		if item.ID == want {
			return item
		}
	}
	return todoRunModelOption{}
}

func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
