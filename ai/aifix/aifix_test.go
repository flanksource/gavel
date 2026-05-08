package aifix

import (
	"context"
	"strings"
	"testing"

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
