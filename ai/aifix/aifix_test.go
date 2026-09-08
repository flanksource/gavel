package aifix

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

func ptr(s string) *string { return &s }

func violation(file, message, rule string, line int) models.Violation {
	v := models.Violation{File: file, Line: line, Source: "betterleaks", Message: ptr(message)}
	if rule != "" {
		v.Rule = &models.Rule{Method: rule}
	}
	return v
}

func resultsWith(linter string, vs ...models.Violation) []*linters.LinterResult {
	return []*linters.LinterResult{{Linter: linter, Violations: vs}}
}

func renderLintPrompt(options ResolveOptions) (api.Spec, error) {
	layers, err := Layers(options)
	if err != nil {
		return api.Spec{}, err
	}
	resolved, err := api.ResolveSpecLayers(api.ResolveSpecOptions{Layers: layers, RequireModel: true})
	return resolved.Spec, err
}

func unexpectedRequestBuild([]*linters.LinterResult) (captainai.Request, error) {
	return captainai.Request{}, errors.New("unexpected request rebuild")
}

// fakeRuntime is the runtime the loop tests register their scripted provider
// under. It has to be an api-mode cell: captain refuses a local mode whose
// executable is missing (`codex`, `tsx`) before it ever reaches the registry, so
// an agent-mode runtime would make these tests assert nothing more than whether
// the machine happens to have an agent CLI installed. The api cell requires no
// binary, and every request pairs it with an explicit APIKey so no environment
// credential is consulted either.
var fakeRuntime = captainai.Runtime{Provider: "openai", Mode: captainai.ModeAPI}

func TestHasViolations_TrueWhenAtLeastOneNonSkippedHasViolations(t *testing.T) {
	res := resultsWith("betterleaks", violation("a.go", "leaked secret", "AWS", 12))
	if !hasViolations(res) {
		t.Fatal("hasViolations = false, want true")
	}
}

func TestHasViolations_FalseWhenAllSkipped(t *testing.T) {
	res := []*linters.LinterResult{{Linter: "x", Skipped: true, Violations: []models.Violation{
		violation("a.go", "msg", "RULE", 1),
	}}}
	if hasViolations(res) {
		t.Error("hasViolations = true on skipped result; want false")
	}
}

func TestHasViolations_FalseWhenNoViolations(t *testing.T) {
	if hasViolations([]*linters.LinterResult{{Linter: "x"}}) {
		t.Error("hasViolations = true on empty result")
	}
}

func TestResolveSpecFormatsViolationsWithRuleAndLocation(t *testing.T) {
	res := resultsWith("betterleaks",
		violation(".env", "AWS access key", "AWS_KEY", 3),
		violation("config.yaml", "GCP key", "", 0),
	)
	spec, err := renderLintPrompt(ResolveOptions{Base: api.Spec{Model: api.Model{Name: "agent:sonnet"}}, Dir: "/repo", Linters: []string{"betterleaks"}, Results: res})
	if err != nil {
		t.Fatalf("ResolveSpec err: %v", err)
	}
	out := spec.Prompt.User
	if !strings.Contains(out, ".env:3 [betterleaks/AWS_KEY] AWS access key") {
		t.Errorf("missing first violation line; out=%q", out)
	}
	if !strings.Contains(out, "config.yaml [betterleaks] GCP key") {
		t.Errorf("missing second violation line; out=%q", out)
	}
}

func TestResolveSpecSkipsSkippedAndEmptyResults(t *testing.T) {
	res := []*linters.LinterResult{
		{Linter: "skipped", Skipped: true, Violations: []models.Violation{violation("x", "x", "X", 1)}},
		{Linter: "empty"},
		{Linter: "real", Violations: []models.Violation{violation("a.go", "msg", "R", 5)}},
	}
	spec, err := renderLintPrompt(ResolveOptions{Base: api.Spec{Model: api.Model{Name: "agent:sonnet"}}, Dir: "/repo", Results: res})
	if err != nil {
		t.Fatalf("ResolveSpec err: %v", err)
	}
	out := spec.Prompt.User
	if strings.Contains(out, "skipped/") || strings.Contains(out, "[empty]") {
		t.Errorf("prompt included skipped/empty linters: %q", out)
	}
	if !strings.Contains(out, "[real/R] msg") {
		t.Errorf("prompt missing real linter line: %q", out)
	}
}

func TestResolveSpecSystemPromptMentionsLintersWhenProvided(t *testing.T) {
	spec, err := renderLintPrompt(ResolveOptions{Base: api.Spec{Model: api.Model{Name: "agent:sonnet"}}, Dir: "/repo", Linters: []string{"betterleaks", "ruff"}, Results: resultsWith("ruff", violation("x.py", "bad", "R", 1))})
	if err != nil {
		t.Fatalf("ResolveSpec err: %v", err)
	}
	out := spec.Prompt.System
	if !strings.Contains(out, "betterleaks, ruff") {
		t.Errorf("system prompt missing linter list: %q", out)
	}
	if !strings.Contains(out, "/repo") {
		t.Errorf("system prompt missing workdir: %q", out)
	}
}

func TestResolveSpecSystemPromptOmitsLinterClauseWhenEmpty(t *testing.T) {
	spec, err := renderLintPrompt(ResolveOptions{Base: api.Spec{Model: api.Model{Name: "agent:sonnet"}}, Dir: "/repo", Results: resultsWith("ruff", violation("x.py", "bad", "R", 1))})
	if err != nil {
		t.Fatalf("ResolveSpec err: %v", err)
	}
	out := spec.Prompt.System
	if strings.Contains(out, "active linters") {
		t.Errorf("system prompt should not mention linters when none given: %q", out)
	}
}

func TestRun_ShortCircuitsOnCleanInitial(t *testing.T) {
	res, err := Run(context.Background(), Request{
		Initial: []*linters.LinterResult{{Linter: "x"}}, // no violations
		ReLint: func(ctx context.Context) ([]*linters.LinterResult, error) {
			t.Fatal("ReLint should not be called when initial is clean")
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("Run err: %v", err)
	}
	if res.StopReason != "condition-met" {
		t.Errorf("StopReason = %q, want condition-met", res.StopReason)
	}
}

func TestRun_ErrorsWhenReLintMissingAndViolationsPresent(t *testing.T) {
	_, err := Run(context.Background(), Request{
		Initial: resultsWith("betterleaks", violation("a", "x", "R", 1)),
	})
	if err == nil || !strings.Contains(err.Error(), "ReLint is required") {
		t.Fatalf("err = %v, want ReLint required", err)
	}
}

// fakeStreaming mirrors captain/pkg/ai/loop_test.go's fakeStreamingProvider:
// records every Request and returns a single scripted Result event per call.
type fakeStreaming struct {
	mu       sync.Mutex
	model    string
	runtime  captainai.Runtime
	requests []captainai.Request
}

func (f *fakeStreaming) GetModel() string              { return f.model }
func (f *fakeStreaming) GetRuntime() captainai.Runtime { return f.runtime }
func (f *fakeStreaming) Execute(ctx context.Context, req captainai.Request) (*captainai.Response, error) {
	return nil, errors.New("not used")
}
func (f *fakeStreaming) ExecuteStream(ctx context.Context, req captainai.Request) (<-chan captainai.Event, error) {
	f.mu.Lock()
	f.requests = append(f.requests, req)
	f.mu.Unlock()
	ch := make(chan captainai.Event, 1)
	ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
	close(ch)
	return ch, nil
}

// fakeBuffered implements only the non-streaming Provider interface, so
// aifix.Run should refuse to drive it through the loop.
type fakeBuffered struct{}

func (f *fakeBuffered) GetModel() string { return "buf" }
func (f *fakeBuffered) GetRuntime() captainai.Runtime {
	return captainai.Runtime{Provider: "anthropic", Mode: captainai.ModeAPI}
}
func (f *fakeBuffered) Execute(ctx context.Context, req captainai.Request) (*captainai.Response, error) {
	return &captainai.Response{}, nil
}

// TestRun_UsesAIConfigFromCaller asserts the provider receives exactly the
// fields callers set on AIConfig + AIRequestProto — the saved captain
// configure defaults that gavel just learned to honour.
func TestRun_UsesAIConfigFromCaller(t *testing.T) {
	p := &fakeStreaming{model: "gpt-5.5", runtime: fakeRuntime}
	captainai.RegisterProvider(fakeRuntime, func(cfg captainai.Config) (captainai.Provider, error) {
		p.model = cfg.Model.Name
		return p, nil
	})

	res, err := Run(context.Background(), Request{
		Initial:       resultsWith("fakelint", violation("x.go", "missing comma", "RULE", 7)),
		MaxIterations: 1,
		AIConfig: captainai.Config{
			Model:  api.Model{Name: "gpt-5.5", Mode: fakeRuntime.Mode},
			APIKey: "example-key",
		},
		AIRequestProto: captainai.Request{
			Prompt:      api.Prompt{System: "Repair lint failures", User: "x.go:7 missing comma"},
			Model:       api.Model{Effort: api.EffortHigh},
			Budget:      api.Budget{MaxTokens: 16000},
			Memory:      api.Memory{SkipHooks: true, SkipSkills: true, SkipUser: true, SkipProject: true, SkipMemory: true},
			Permissions: api.Permissions{MCP: api.MCP{Disabled: true}},
		},
		BuildRequest: unexpectedRequestBuild,
		ReLint: func(ctx context.Context) ([]*linters.LinterResult, error) {
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("Run err: %v", err)
	}
	if len(p.requests) == 0 {
		t.Fatal("streaming provider never invoked")
	}
	got := p.requests[0]
	for name, b := range map[string]bool{
		"MCP.Disabled": got.Permissions.MCP.Disabled, "Memory.SkipHooks": got.Memory.SkipHooks,
		"Memory.SkipSkills": got.Memory.SkipSkills, "Memory.SkipUser": got.Memory.SkipUser,
		"Memory.SkipProject": got.Memory.SkipProject, "Memory.SkipMemory": got.Memory.SkipMemory,
	} {
		if !b {
			t.Errorf("%s = false, want true (propagated from AIRequestProto)", name)
		}
	}
	if got.Budget.MaxTokens != 16000 {
		t.Errorf("MaxTokens = %d, want 16000", got.Budget.MaxTokens)
	}
	if got.Effort != api.EffortHigh {
		t.Errorf("Effort = %q, want high", got.Effort)
	}
	if got.Prompt.System == "" {
		t.Error("SystemPrompt unset, expected aifix to fill it")
	}
	if got.Prompt.User == "" {
		t.Error("Prompt unset, expected aifix to fill it with violation list")
	}
	if res.StopReason == "error" {
		t.Errorf("unexpected error stop reason; res=%+v", res)
	}
}

// TestRun_NoModelErrors verifies an empty AIConfig.Model surfaces captain's
// "run captain configure" error verbatim. This is the user-facing failure
// when captain configure has never been run and no --model flag is passed.
func TestRun_NoModelErrors(t *testing.T) {
	_, err := Run(context.Background(), Request{
		Initial:      resultsWith("fakelint", violation("a", "x", "R", 1)),
		ReLint:       func(ctx context.Context) ([]*linters.LinterResult, error) { return nil, nil },
		AIConfig:     captainai.Config{},
		BuildRequest: unexpectedRequestBuild,
	})
	if err == nil {
		t.Fatal("expected error for empty model, got nil")
	}
	if !strings.Contains(err.Error(), "captain configure") {
		t.Errorf("error %q should mention 'captain configure'", err.Error())
	}
}

// TestRun_SurfacesReLintError verifies a failing ReLint between iterations
// is reported back to the caller instead of being silently swallowed. The
// loop must stop fast so the model isn't asked to fix stale violations.
func TestRun_SurfacesReLintError(t *testing.T) {
	p := &fakeStreaming{model: "gpt-5.6-sol", runtime: fakeRuntime}
	captainai.RegisterProvider(fakeRuntime, func(cfg captainai.Config) (captainai.Provider, error) {
		return p, nil
	})
	boom := errors.New("re-lint command failed: exit status 1")
	res, err := Run(context.Background(), Request{
		Initial:        resultsWith("fakelint", violation("a", "x", "R", 1)),
		MaxIterations:  3,
		AIRequestProto: captainai.Request{Prompt: api.Prompt{User: "Repair the lint failure"}},
		BuildRequest:   unexpectedRequestBuild,
		AIConfig: captainai.Config{
			Model:  api.Model{Name: "gpt-5.6-sol", Mode: fakeRuntime.Mode},
			APIKey: "example-key",
		},
		ReLint: func(ctx context.Context) ([]*linters.LinterResult, error) {
			return nil, boom
		},
	})
	if err == nil {
		t.Fatal("expected ReLint error to surface, got nil")
	}
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want wrap of %v", err, boom)
	}
	if res == nil || res.StopReason != "relint-error" {
		t.Errorf("StopReason = %v, want relint-error; res=%+v", res, res)
	}
}

// TestRun_NonStreamingRuntimeErrors guards against runtimes that only
// implement buffered Execute. Aifix needs streaming for live progress, so
// it must error rather than silently degrade to one-shot calls.
func TestRun_NonStreamingRuntimeErrors(t *testing.T) {
	captainai.RegisterProvider(captainai.Runtime{Provider: "anthropic", Mode: captainai.ModeAPI}, func(cfg captainai.Config) (captainai.Provider, error) {
		return &fakeBuffered{}, nil
	})
	_, err := Run(context.Background(), Request{
		Initial:      resultsWith("fakelint", violation("a", "x", "R", 1)),
		ReLint:       func(ctx context.Context) ([]*linters.LinterResult, error) { return nil, nil },
		BuildRequest: unexpectedRequestBuild,
		AIConfig: captainai.Config{
			Model: api.Model{Name: "claude-sonnet-5", Mode: captainai.ModeAPI},
		},
	})
	if err == nil {
		t.Fatal("expected error for a non-streaming runtime, got nil")
	}
	if !strings.Contains(err.Error(), "not streaming") {
		t.Errorf("error %q should explain the runtime is not streaming", err.Error())
	}
}
