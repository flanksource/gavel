package aifix

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
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

func TestBuildPrompt_FormatsViolationsWithRuleAndLocation(t *testing.T) {
	res := resultsWith("betterleaks",
		violation(".env", "AWS access key", "AWS_KEY", 3),
		violation("config.yaml", "GCP key", "", 0),
	)
	out := buildPrompt("/repo", res)
	if !strings.Contains(out, ".env:3 [betterleaks/AWS_KEY] AWS access key") {
		t.Errorf("missing first violation line; out=%q", out)
	}
	if !strings.Contains(out, "config.yaml [betterleaks] GCP key") {
		t.Errorf("missing second violation line; out=%q", out)
	}
}

func TestBuildPrompt_SkipsSkippedAndEmptyResults(t *testing.T) {
	res := []*linters.LinterResult{
		{Linter: "skipped", Skipped: true, Violations: []models.Violation{violation("x", "x", "X", 1)}},
		{Linter: "empty"},
		{Linter: "real", Violations: []models.Violation{violation("a.go", "msg", "R", 5)}},
	}
	out := buildPrompt("/repo", res)
	if strings.Contains(out, "skipped/") || strings.Contains(out, "[empty]") {
		t.Errorf("prompt included skipped/empty linters: %q", out)
	}
	if !strings.Contains(out, "[real/R] msg") {
		t.Errorf("prompt missing real linter line: %q", out)
	}
}

func TestBuildSystemPrompt_MentionsLintersWhenProvided(t *testing.T) {
	out := buildSystemPrompt("/repo", []string{"betterleaks", "ruff"})
	if !strings.Contains(out, "betterleaks, ruff") {
		t.Errorf("system prompt missing linter list: %q", out)
	}
	if !strings.Contains(out, "/repo") {
		t.Errorf("system prompt missing workdir: %q", out)
	}
}

func TestBuildSystemPrompt_OmitsLinterClauseWhenEmpty(t *testing.T) {
	out := buildSystemPrompt("/repo", nil)
	if strings.Contains(out, "active linters") {
		t.Errorf("system prompt should not mention linters when none given: %q", out)
	}
}

func TestRun_ShortCircuitsOnCleanInitial(t *testing.T) {
	res, err := Run(context.Background(), Request{
		WorkDir: "/repo",
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
		WorkDir: "/repo",
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
	backend  captainai.Backend
	requests []captainai.Request
}

func (f *fakeStreaming) GetModel() string              { return f.model }
func (f *fakeStreaming) GetBackend() captainai.Backend { return f.backend }
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

func (f *fakeBuffered) GetModel() string              { return "buf" }
func (f *fakeBuffered) GetBackend() captainai.Backend { return captainai.Backend("buffered-only") }
func (f *fakeBuffered) Execute(ctx context.Context, req captainai.Request) (*captainai.Response, error) {
	return &captainai.Response{}, nil
}

// TestRun_UsesAIConfigFromCaller asserts the provider receives exactly the
// fields callers set on AIConfig + AIRequestProto — the saved captain
// configure defaults that gavel just learned to honour.
func TestRun_UsesAIConfigFromCaller(t *testing.T) {
	p := &fakeStreaming{model: "gpt-5.5", backend: captainai.Backend("test-streaming")}
	captainai.RegisterProvider(captainai.Backend("test-streaming"), func(cfg captainai.Config) captainai.Provider {
		p.model = cfg.Model
		return p
	})

	res, err := Run(context.Background(), Request{
		WorkDir:       "/repo",
		Linters:       []string{"fakelint"},
		Initial:       resultsWith("fakelint", violation("x.go", "missing comma", "RULE", 7)),
		MaxIterations: 1,
		AIConfig: captainai.Config{
			Backend: captainai.Backend("test-streaming"),
			Model:   "gpt-5.5",
		},
		AIRequestProto: captainai.Request{
			NoMCP:           true,
			NoHooks:         true,
			NoSkills:        true,
			NoUser:          true,
			NoProject:       true,
			NoMemory:        true,
			MaxTokens:       16000,
			ReasoningEffort: "high",
		},
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
		"NoMCP": got.NoMCP, "NoHooks": got.NoHooks, "NoSkills": got.NoSkills,
		"NoUser": got.NoUser, "NoProject": got.NoProject, "NoMemory": got.NoMemory,
	} {
		if !b {
			t.Errorf("%s = false, want true (propagated from AIRequestProto)", name)
		}
	}
	if got.MaxTokens != 16000 {
		t.Errorf("MaxTokens = %d, want 16000", got.MaxTokens)
	}
	if got.ReasoningEffort != "high" {
		t.Errorf("ReasoningEffort = %q, want high", got.ReasoningEffort)
	}
	if got.SystemPrompt == "" {
		t.Error("SystemPrompt unset, expected aifix to fill it")
	}
	if got.Prompt == "" {
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
		Initial:  resultsWith("fakelint", violation("a", "x", "R", 1)),
		ReLint:   func(ctx context.Context) ([]*linters.LinterResult, error) { return nil, nil },
		AIConfig: captainai.Config{},
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
	p := &fakeStreaming{model: "rl", backend: captainai.Backend("test-relint-err")}
	captainai.RegisterProvider(captainai.Backend("test-relint-err"), func(cfg captainai.Config) captainai.Provider {
		return p
	})
	boom := errors.New("re-lint command failed: exit status 1")
	res, err := Run(context.Background(), Request{
		Initial:       resultsWith("fakelint", violation("a", "x", "R", 1)),
		MaxIterations: 3,
		AIConfig: captainai.Config{
			Backend: captainai.Backend("test-relint-err"),
			Model:   "rl",
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

// TestRun_NonStreamingBackendErrors guards against backends that only
// implement buffered Execute. Aifix needs streaming for live progress, so
// it must error rather than silently degrade to one-shot calls.
func TestRun_NonStreamingBackendErrors(t *testing.T) {
	captainai.RegisterProvider(captainai.Backend("test-buffered-only"), func(cfg captainai.Config) captainai.Provider {
		return &fakeBuffered{}
	})
	_, err := Run(context.Background(), Request{
		Initial: resultsWith("fakelint", violation("a", "x", "R", 1)),
		ReLint:  func(ctx context.Context) ([]*linters.LinterResult, error) { return nil, nil },
		AIConfig: captainai.Config{
			Backend: captainai.Backend("test-buffered-only"),
			Model:   "buf",
		},
	})
	if err == nil {
		t.Fatal("expected error for non-streaming backend, got nil")
	}
	if !strings.Contains(err.Error(), "not streaming") {
		t.Errorf("error %q should explain the backend is not streaming", err.Error())
	}
}
