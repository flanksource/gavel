package checks

import (
	"context"
	"strings"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/todos/types"
)

func withLintRunner(t *testing.T, fn LintFunc) {
	t.Helper()
	orig := lintRunner
	t.Cleanup(func() { lintRunner = orig })
	SetLintRunner(fn)
}

func TestRunLintOnlyReportsViolations(t *testing.T) {
	withLintRunner(t, func(ctx context.Context, workDir string, cfg types.AgentLintConfig) ([]*linters.LinterResult, error) {
		return []*linters.LinterResult{{
			Linter:  "golangci-lint",
			Success: true,
			Violations: []models.Violation{
				{File: "pkg/foo.go", Line: 10, Source: "errcheck", Message: models.StringPtr("error not checked")},
			},
		}}, nil
	})

	res, err := Run(context.Background(), t.TempDir(), types.AgentChecksConfig{Lint: &types.AgentLintConfig{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.OK {
		t.Errorf("expected OK=false with a violation present")
	}
	if res.Violations != 1 {
		t.Errorf("expected 1 violation, got %d", res.Violations)
	}
	if !strings.Contains(res.Summary, "pkg/foo.go:10") {
		t.Errorf("summary should name the violation location, got:\n%s", res.Summary)
	}
}

func TestRunCleanLintIsOK(t *testing.T) {
	withLintRunner(t, func(ctx context.Context, workDir string, cfg types.AgentLintConfig) ([]*linters.LinterResult, error) {
		return []*linters.LinterResult{{Linter: "golangci-lint", Success: true}}, nil
	})

	res, err := Run(context.Background(), t.TempDir(), types.AgentChecksConfig{Lint: &types.AgentLintConfig{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.OK {
		t.Errorf("expected OK=true with no violations, summary:\n%s", res.Summary)
	}
	if res.Summary != "" {
		t.Errorf("expected empty summary when OK, got:\n%s", res.Summary)
	}
}

func TestRunNoChecksIsOK(t *testing.T) {
	// Neither test nor lint configured → nothing to run → trivially OK.
	res, err := Run(context.Background(), t.TempDir(), types.AgentChecksConfig{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.OK {
		t.Errorf("expected OK with no checks configured")
	}
}

func TestRunSkipsLintWhenRunnerUnset(t *testing.T) {
	// Lint requested but no runner registered → lint silently skipped, OK.
	withLintRunner(t, nil)
	res, err := Run(context.Background(), t.TempDir(), types.AgentChecksConfig{Lint: &types.AgentLintConfig{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.OK {
		t.Errorf("expected OK when lint runner is unset")
	}
}
