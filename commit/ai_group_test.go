package commit

import (
	"context"
	"strings"
	"testing"

	"github.com/flanksource/gavel/status"
	"github.com/flanksource/repomap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildStatusTable(t *testing.T) {
	orig := gatherStatusFunc
	t.Cleanup(func() { gatherStatusFunc = orig })
	gatherStatusFunc = func(workDir string) (*status.Result, error) {
		return &status.Result{Files: []status.FileStatus{
			{Path: "api/a.go", FileMap: &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeType("api")}}},
			{Path: "README.md"},
		}}, nil
	}

	changes := []stagedChange{
		{Path: "api/a.go", Status: "updated", Adds: 3, Dels: 1},
		{Path: "README.md", Status: "updated", Adds: 2, Dels: 2},
		// Not present in the gathered status — must still get a row (general scope),
		// never dropped, so the LLM can assign it.
		{Path: "ghost.go", Status: "updated", Adds: 1, Dels: 1},
	}

	got, err := buildStatusTable("/repo", changes, false)
	require.NoError(t, err)

	// A markdown table with the expected columns and one row per change.
	assert.Contains(t, got, "SCOPE")
	assert.Contains(t, got, "FILE")
	assert.Contains(t, got, "api/a.go")
	assert.Contains(t, got, "go · api")
	assert.Contains(t, got, "README.md")
	// Change absent from status falls back to the "general" scope rather than being omitted.
	assert.Contains(t, got, "ghost.go")
	assert.Contains(t, got, scopeGeneralFallback)
}

func TestSortGroupingRowsByScope(t *testing.T) {
	rows := func() []groupingRow {
		return []groupingRow{
			{Scope: "go · b", File: "z.go"},
			{Scope: "go · a", File: "y.go"},
			{Scope: "go · b", File: "a.go"},
		}
	}

	byScope := rows()
	sortGroupingRows(byScope, true)
	assert.Equal(t, []string{"y.go", "a.go", "z.go"}, []string{byScope[0].File, byScope[1].File, byScope[2].File},
		"group-by-scope orders by scope then file")

	byFile := rows()
	sortGroupingRows(byFile, false)
	assert.Equal(t, []string{"a.go", "y.go", "z.go"}, []string{byFile[0].File, byFile[1].File, byFile[2].File},
		"flat ordering is by file alone")
}

func TestGroupWithMaxCommitsConsolidates(t *testing.T) {
	// First call overshoots (3 groups), feedback round returns 2 — within the cap.
	responses := []aiGroupingSchema{
		{Groups: []aiGroup{{Label: "a"}, {Label: "b"}, {Label: "c"}}, Ignore: []string{"lock"}},
		{Groups: []aiGroup{{Label: "ab"}, {Label: "c"}}, Ignore: []string{"lock"}},
	}
	var calls []string
	exec := func(feedback string) (aiGroupingSchema, error) {
		calls = append(calls, feedback)
		r := responses[len(calls)-1]
		return r, nil
	}

	got, err := groupWithMaxCommits(2, exec)
	require.NoError(t, err)
	assert.Len(t, got.Groups, 2, "stops once within the limit")
	require.Len(t, calls, 2)
	assert.Empty(t, calls[0], "first call uses the base prompt (no feedback)")
	assert.Contains(t, calls[1], "limit is 2", "feedback re-prompt cites the cap")
}

func TestGroupWithMaxCommitsIgnoreDoesNotCount(t *testing.T) {
	// 2 groups + a large ignore list; the cap of 2 is satisfied without feedback.
	exec := func(feedback string) (aiGroupingSchema, error) {
		return aiGroupingSchema{
			Groups: []aiGroup{{Label: "a"}, {Label: "b"}},
			Ignore: []string{"x", "y", "z"},
		}, nil
	}
	got, err := groupWithMaxCommits(2, callCounter(t, exec, 1))
	require.NoError(t, err)
	assert.Len(t, got.Groups, 2)
}

func TestGroupWithMaxCommitsStopsAtAttemptBound(t *testing.T) {
	// LLM never gets under the cap; the loop must stop after maxGroupingAttempts
	// calls and return the last grouping rather than erroring or looping forever.
	var calls int
	exec := func(feedback string) (aiGroupingSchema, error) {
		calls++
		return aiGroupingSchema{Groups: []aiGroup{{Label: "a"}, {Label: "b"}, {Label: "c"}}}, nil
	}
	got, err := groupWithMaxCommits(2, exec)
	require.NoError(t, err)
	assert.Equal(t, maxGroupingAttempts, calls)
	assert.Len(t, got.Groups, 3)
}

// callCounter wraps exec asserting it is invoked exactly want times.
func callCounter(t *testing.T, exec func(string) (aiGroupingSchema, error), want int) func(string) (aiGroupingSchema, error) {
	t.Helper()
	var calls int
	t.Cleanup(func() { assert.Equal(t, want, calls) })
	return func(feedback string) (aiGroupingSchema, error) {
		calls++
		return exec(feedback)
	}
}

func TestAssembleGroupsMapsGroupsIgnoreAndLeftovers(t *testing.T) {
	changes := []stagedChange{
		{Path: "feature/a.go"},
		{Path: "feature/b.go"},
		{Path: "pnpm-lock.yaml"},
		{Path: "orphan.txt"},
	}

	groups := assembleGroups(changes, aiGroupingSchema{
		Groups: []aiGroup{
			// "ghost.go" is unknown and must be skipped, not committed.
			{Label: "feat: feature", Files: []string{"feature/a.go", "feature/b.go", "ghost.go"}},
		},
		Ignore: []string{"pnpm-lock.yaml"},
	})

	require.Len(t, groups, 3)

	assert.Equal(t, "feat: feature", groups[0].Label)
	assert.Empty(t, groups[0].Message)
	assert.Equal(t, []string{"feature/a.go", "feature/b.go"}, groups[0].Files())

	// Unassigned files are committed in a trailing "other" group, never dropped.
	assert.Equal(t, "other", groups[1].Label)
	assert.Equal(t, []string{"orphan.txt"}, groups[1].Files())

	// Ignored files become the chore group last, with the preset message.
	assert.Equal(t, choreGroupLabel, groups[2].Label)
	assert.Equal(t, choreGroupMessage, groups[2].Message)
	assert.Equal(t, []string{"pnpm-lock.yaml"}, groups[2].Files())
}

func TestAssembleGroupsNoIgnoreNoLeftovers(t *testing.T) {
	changes := []stagedChange{{Path: "a.go"}, {Path: "b.go"}}

	groups := assembleGroups(changes, aiGroupingSchema{
		Groups: []aiGroup{
			{Label: "g1", Files: []string{"a.go"}},
			{Label: "g2", Files: []string{"b.go"}},
		},
	})

	require.Len(t, groups, 2)
	assert.Equal(t, "g1", groups[0].Label)
	assert.Equal(t, "g2", groups[1].Label)
}

func TestRunAIGroupSplitsAndCreatesChoreCommit(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "feature/a.go", "package feature\n")
	writeFileInDir(t, repo, "bugfix/b.go", "package bugfix\n")
	writeFile(t, repo, "pnpm-lock.yaml", "lockfileVersion: 9\n")
	gitRun(t, repo, "add", "feature/a.go", "bugfix/b.go", "pnpm-lock.yaml")

	t.Setenv(testEnvVar, "1")

	orig := groupChangesByAIFunc
	t.Cleanup(func() { groupChangesByAIFunc = orig })
	groupChangesByAIFunc = func(ctx context.Context, opts Options, source stagedSource) ([]commitGroup, error) {
		return assembleGroups(source.Changes, aiGroupingSchema{
			Groups: []aiGroup{
				{Label: "feat: a", Files: []string{"feature/a.go"}},
				{Label: "fix: b", Files: []string{"bugfix/b.go"}},
			},
			Ignore: []string{"pnpm-lock.yaml"},
		}), nil
	}

	result, err := Run(context.Background(), Options{WorkDir: repo, AIGroup: true})
	require.NoError(t, err)

	require.Len(t, result.Commits, 3)
	assert.Equal(t, []string{"feature/a.go"}, result.Commits[0].Files)
	assert.Equal(t, []string{"bugfix/b.go"}, result.Commits[1].Files)
	assert.Equal(t, []string{"pnpm-lock.yaml"}, result.Commits[2].Files)
	assert.Equal(t, choreGroupMessage, result.Commits[2].Message)
	for _, c := range result.Commits {
		assert.NotEmpty(t, c.Hash)
	}

	// 1 initial commit + 3 created here.
	assert.Equal(t, "4", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
	assert.Empty(t, strings.TrimSpace(gitOutput(t, repo, "status", "--short")))
}

func TestRunAIGroupRejectsMessage(t *testing.T) {
	repo := initCommitRepo(t)
	_, err := Run(context.Background(), Options{WorkDir: repo, AIGroup: true, Message: "x"})
	assert.ErrorIs(t, err, ErrAIGroupWithMessage)
}
