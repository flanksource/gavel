package main

import (
	"os"
	"path/filepath"
	"testing"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func TestCollectViolationTypes(t *testing.T) {
	msg1 := "error return value not checked"
	msg2 := "unused variable"

	results := []*linters.LinterResult{
		{
			Violations: []models.Violation{
				{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}, File: "a.go", Message: &msg1},
				{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}, File: "b.go", Message: &msg1},
				{Source: "golangci-lint", Rule: &models.Rule{Method: "unused"}, File: "a.go", Message: &msg2},
			},
		},
		{
			Violations: []models.Violation{
				{Source: "eslint", Rule: &models.Rule{Method: "no-unused-vars"}, File: "x.ts"},
			},
		},
		nil,
	}

	types := collectViolationTypes(results)

	assert.Len(t, types, 3)

	// Sorted by count descending
	assert.Equal(t, "errcheck", types[0].Rule)
	assert.Equal(t, "golangci-lint", types[0].Source)
	assert.Equal(t, 2, types[0].Count)
	assert.ElementsMatch(t, []string{"a.go", "b.go"}, types[0].Files)
	assert.Equal(t, msg1, types[0].Example)

	assert.Equal(t, 1, types[1].Count)
	assert.Equal(t, 1, types[2].Count)
}

func TestCollectViolationTypes_Empty(t *testing.T) {
	types := collectViolationTypes(nil)
	assert.Empty(t, types)
}

func TestCollectViolationTypes_NilRule(t *testing.T) {
	msg := "some error"
	results := []*linters.LinterResult{
		{
			Violations: []models.Violation{
				{Source: "custom", File: "f.go", Message: &msg},
				{Source: "custom", File: "g.go"},
			},
		},
	}

	types := collectViolationTypes(results)
	assert.Len(t, types, 1)
	assert.Equal(t, "", types[0].Rule)
	assert.Equal(t, "custom", types[0].Source)
	assert.Equal(t, 2, types[0].Count)
}

func TestHandleCommitLintFindings_ContinueOnce(t *testing.T) {
	repo := t.TempDir()
	prev := promptLintFindingsAction
	t.Cleanup(func() { promptLintFindingsAction = prev })
	promptLintFindingsAction = func() lintFindingsAction { return lintActionContinueOnce }

	msg := "boom"
	result := &commitpkg.Result{
		Lint: &commitpkg.LintGateResult{
			Violations: 1,
			Results: []*linters.LinterResult{{
				Linter: "fake",
				Violations: []models.Violation{{
					Source: "fake", Rule: &models.Rule{Method: "rule1"},
					File: "a.go", Message: &msg,
				}},
			}},
		},
	}

	got := handleCommitLintFindings(repo, result, false)
	assert.Equal(t, lintFindingsContinueOnce, got)

	// Continue-once must NOT write any rules to .gavel.yaml.
	_, err := os.Stat(filepath.Join(repo, ".gavel.yaml"))
	assert.True(t, os.IsNotExist(err), "continue-once must not persist .gavel.yaml")
}

func TestHandleCommitLintFindings_Cancel(t *testing.T) {
	repo := t.TempDir()
	prev := promptLintFindingsAction
	t.Cleanup(func() { promptLintFindingsAction = prev })
	promptLintFindingsAction = func() lintFindingsAction { return lintActionCancel }

	msg := "boom"
	result := &commitpkg.Result{
		Lint: &commitpkg.LintGateResult{
			Violations: 1,
			Results: []*linters.LinterResult{{
				Linter: "fake",
				Violations: []models.Violation{{
					Source: "fake", File: "a.go", Message: &msg,
				}},
			}},
		},
	}

	got := handleCommitLintFindings(repo, result, false)
	assert.Equal(t, lintFindingsBlocked, got)

	_, err := os.Stat(filepath.Join(repo, ".gavel.yaml"))
	assert.True(t, os.IsNotExist(err), "cancel must not persist .gavel.yaml")
}

// With assumeYes the lint-findings prompt must never open; the AI-fix path is
// taken directly. We use a violation with no file so runCommitAIFix bails at
// its "no files reported violations" guard before invoking any AI, which keeps
// the test hermetic while still proving the prompt was bypassed.
func TestHandleCommitLintFindings_AssumeYesSkipsPrompt(t *testing.T) {
	prev := promptLintFindingsAction
	t.Cleanup(func() { promptLintFindingsAction = prev })
	promptLintFindingsAction = func() lintFindingsAction {
		t.Fatal("prompt must not be called when assumeYes is set")
		return lintActionCancel
	}

	msg := "boom"
	result := &commitpkg.Result{
		Lint: &commitpkg.LintGateResult{
			Violations: 1,
			Results: []*linters.LinterResult{{
				Linter: "fake",
				Violations: []models.Violation{{
					Source: "fake", File: "", Message: &msg,
				}},
			}},
		},
	}

	got := handleCommitLintFindings(t.TempDir(), result, true)
	assert.Equal(t, lintFindingsBlocked, got)
}

func TestHandleCommitLintFindings_NilResultBlocks(t *testing.T) {
	got := handleCommitLintFindings(t.TempDir(), nil, false)
	assert.Equal(t, lintFindingsBlocked, got)

	got = handleCommitLintFindings(t.TempDir(), &commitpkg.Result{Lint: nil}, false)
	assert.Equal(t, lintFindingsBlocked, got)
}

// lintActionAIFix must be the first entry in the iota so the prompt
// presents "AI Fix" above Triage. If someone reorders the const block, the
// commit-time prompt and runCommit's switch would silently route the user
// to the wrong branch — this guard fails the build instead.
func TestLintActionIotaOrder(t *testing.T) {
	assert.Equal(t, lintFindingsAction(0), lintActionAIFix)
	assert.Equal(t, lintFindingsAction(1), lintActionTriage)
	assert.Equal(t, lintFindingsAction(2), lintActionContinueOnce)
	assert.Equal(t, lintFindingsAction(3), lintActionCancel)
}
