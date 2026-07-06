package status

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// threeProblemFile builds a single changed file carrying three ordered
// problems so the cap / verbose / hint behaviour can be exercised without a
// snapshot round-trip.
func threeProblemFile() FileStatus {
	return FileStatus{
		Path:  "testrunner/runner.go",
		State: StateUnstaged,
		Problems: []Problem{
			{Kind: ProblemKindTest, Severity: "failed", Label: "TestRun/streams", Line: 88, Message: "expected 3, got 2"},
			{Kind: ProblemKindLint, Severity: "error", Label: "errcheck", Line: 142, Message: "Error return value not checked"},
			{Kind: ProblemKindLint, Severity: "warning", Label: "gofmt", Line: 5, Message: "file is not gofmt-ed"},
		},
	}
}

func TestGatherPopulatesProblemsFromSnapshot(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	restore := stubSnapshot(func(string) (string, string, error) {
		return "deadbeef", "", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return &snapshots.Pointer{SHA: "deadbeef", Path: ".gavel/sha-deadbeef.json"}, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		return &testui.Snapshot{
			Tests: []parsers.Test{{
				Name: "TestRun/streams", File: "a.go", Line: 88,
				Failed: true, Message: "expected 3, got 2",
			}},
			Lint: []*linters.LinterResult{{
				Violations: []models.Violation{
					{File: "a.go", Line: 142, Severity: "error",
						Message: models.StringPtr("Error return value not checked"),
						Rule:    &models.Rule{Method: "errcheck"}},
					{File: "a.go", Line: 3, Severity: "info",
						Message: models.StringPtr("informational only")},
				},
			}},
		}, nil
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	problems := result.Files[0].Problems
	// The failing test and the error violation surface; the info violation
	// stays badge-only (LintStatus.Infos) and is excluded from Problems.
	require.Len(t, problems, 2)
	assert.Equal(t, 1, result.Files[0].LintStatus.Infos)

	// Failed test outranks the lint error, so it sorts first.
	assert.Equal(t, ProblemKindTest, problems[0].Kind)
	assert.Equal(t, "TestRun/streams", problems[0].Label)
	assert.Equal(t, "expected 3, got 2", problems[0].Message)
	assert.Equal(t, 88, problems[0].Line)

	assert.Equal(t, ProblemKindLint, problems[1].Kind)
	assert.Equal(t, "errcheck", problems[1].Label)
	assert.Equal(t, "Error return value not checked", problems[1].Message)
}

func TestPrettyProblemsSectionCapsAndHints(t *testing.T) {
	r := &Result{Files: []FileStatus{threeProblemFile()}}

	clean := stripANSI(r.Pretty().ANSI())

	assert.Contains(t, clean, "Problems (3)")
	assert.Contains(t, clean, "testrunner/runner.go")
	// First two problems render; the third is capped away.
	assert.Contains(t, clean, "TestRun/streams")
	assert.Contains(t, clean, "runner.go:88")
	assert.Contains(t, clean, "expected 3, got 2")
	assert.Contains(t, clean, "errcheck")
	assert.NotContains(t, clean, "gofmt")
	assert.NotContains(t, clean, "file is not gofmt-ed")
	// The cap advertises the remaining count and the way to see it.
	assert.Contains(t, clean, "… 1 more")
	assert.Contains(t, clean, "gavel status -v")
}

func TestPrettyProblemsVerboseShowsEveryProblem(t *testing.T) {
	f := threeProblemFile()
	f.Problems[0].Message = "expected 3, got 2\n  at runner.go:88\n  in TestRun"
	r := &Result{Verbose: true, Files: []FileStatus{f}}

	clean := stripANSI(r.Pretty().ANSI())

	// All three problems, including the capped-away warning, are shown.
	assert.Contains(t, clean, "TestRun/streams")
	assert.Contains(t, clean, "errcheck")
	assert.Contains(t, clean, "gofmt")
	assert.Contains(t, clean, "file is not gofmt-ed")
	// Multi-line messages keep their continuation lines in verbose mode.
	assert.Contains(t, clean, "at runner.go:88")
	assert.Contains(t, clean, "in TestRun")
	// Nothing is capped, so no truncation hints appear.
	assert.NotContains(t, clean, "more")
	assert.NotContains(t, clean, "gavel status -v")
}

func TestPrettyProblemsHintShownWhenNotCapped(t *testing.T) {
	f := FileStatus{
		Path:  "a.go",
		State: StateUnstaged,
		Problems: []Problem{
			{Kind: ProblemKindLint, Severity: "error", Label: "errcheck", Line: 10, Message: "unchecked error"},
		},
	}
	r := &Result{Files: []FileStatus{f}}

	clean := stripANSI(r.Pretty().ANSI())

	// A single problem needs no per-file cap, but the footer still teaches
	// the reader how to see full (untruncated) logs.
	assert.NotContains(t, clean, "more")
	assert.Contains(t, clean, "run gavel status -v for full test/lint logs")
}

func TestPrettyProblemsIncludesAIError(t *testing.T) {
	f := FileStatus{
		Path:     "a.go",
		State:    StateUnstaged,
		AIError:  "model timed out after 5m",
		AIStatus: AISummaryStatusFailed,
	}
	r := &Result{Files: []FileStatus{f}}

	clean := stripANSI(r.Pretty().ANSI())

	assert.Contains(t, clean, "Problems (1)")
	assert.Contains(t, clean, "ai summary")
	assert.Contains(t, clean, "model timed out after 5m")
}

func TestPrettyNoProblemsSectionWhenClean(t *testing.T) {
	r := &Result{Files: []FileStatus{{Path: "a.go", State: StateUnstaged}}}

	clean := stripANSI(r.Pretty().ANSI())

	assert.NotContains(t, clean, "Problems")
	assert.False(t, strings.Contains(clean, "gavel status -v"))
}
