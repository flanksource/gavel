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

func TestBuildChangeSummary(t *testing.T) {
	changes := []stagedChange{
		{Path: "a/x.go", Status: "updated", Adds: 3, Dels: 1},
		{Path: "new.txt", Status: "inserted", Adds: 5, Dels: 0},
		{Path: "b/new.go", PreviousPath: "b/old.go", Status: "renamed", Adds: 2, Dels: 2},
	}

	want := "updated  a/x.go (+3/-1)\n" +
		"inserted new.txt (+5/-0)\n" +
		"renamed  b/old.go -> b/new.go (+2/-2)\n"

	assert.Equal(t, want, buildChangeSummary(changes))
}

func TestBuildScopeGroupedSummary(t *testing.T) {
	orig := gatherStatusFunc
	t.Cleanup(func() { gatherStatusFunc = orig })
	gatherStatusFunc = func(workDir string) (*status.Result, error) {
		return &status.Result{Files: []status.FileStatus{
			{Path: "api/a.go", FileMap: &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeType("api")}}},
			{Path: "api/b.go", FileMap: &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeType("api")}}},
			{Path: "README.md"},
			// In status but not in the commit set — must be omitted.
			{Path: "other.go", FileMap: &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeType("api")}}},
		}}, nil
	}

	changes := []stagedChange{
		{Path: "api/a.go", Status: "updated", Adds: 3, Dels: 1},
		{Path: "api/b.go", Status: "inserted", Adds: 5, Dels: 0},
		{Path: "README.md", Status: "updated", Adds: 2, Dels: 2},
		// Not present in status — must be omitted from the summary.
		{Path: "ghost.go", Status: "updated", Adds: 1, Dels: 1},
	}

	got, err := buildScopeGroupedSummary("/repo", changes)
	require.NoError(t, err)

	want := "[scope: go · api]\n" +
		"updated  api/a.go (+3/-1)\n" +
		"inserted api/b.go (+5/-0)\n" +
		"\n" +
		"[scope: general]\n" +
		"updated  README.md (+2/-2)\n" +
		"\n"
	assert.Equal(t, want, got)
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
