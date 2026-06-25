package commit

import (
	"testing"

	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStageAllSkipsEmbeddedRepoWithoutError reproduces the reported failure:
// `git ls-files --others` reports an embedded repository as a single `pr/ui/`
// directory entry that `git add` then refuses, aborting the entire stage (the
// origin of the `git add -A … paths are ignored` error). Staging must skip such
// entries and still commit the other files instead of failing the whole run.
func TestStageAllSkipsEmbeddedRepoWithoutError(t *testing.T) {
	dir := initCommitRepo(t)
	writeFile(t, dir, ".gitignore", "*.js\n")
	gitRun(t, dir, "add", ".gitignore")
	gitRun(t, dir, "commit", "-m", "add gitignore")

	// Embedded git repo whose only file is ignored by the parent's *.js rule.
	gitRun(t, dir, "init", "pr/ui")
	writeFileInDir(t, dir, "pr/ui/index.js", "console.log(1)\n")
	writeFile(t, dir, "main.go", "package main\n")

	err := stageFiles(dir, StageAll, verify.CommitConfig{})
	require.NoError(t, err)

	staged := mustStagedFiles(t, dir)
	assert.Contains(t, staged, "main.go")
	for _, f := range staged {
		assert.NotContains(t, f, "pr/ui", "embedded repo path must not be staged")
	}
}

// TestStageAllStripsTrackedButIgnoredModifications encodes the rule that
// .gitignore is authoritative for `gavel commit`: a force-tracked bundle such as
// dist/bundle.js must be left out of the commit, while a !-negated sibling and
// normal changes are staged. The bundle stays tracked (non-destructive); the
// release CI re-commits it with raw `git add -f`.
func TestStageAllStripsTrackedButIgnoredModifications(t *testing.T) {
	dir := initCommitRepo(t)
	writeFile(t, dir, ".gitignore", "dist/*\n!dist/keep.js\n")
	gitRun(t, dir, "add", ".gitignore")
	gitRun(t, dir, "commit", "-m", "ignore dist")

	writeFileInDir(t, dir, "dist/bundle.js", "v1\n")
	writeFileInDir(t, dir, "dist/keep.js", "v1\n")
	gitRun(t, dir, "add", "-f", "dist/bundle.js")
	gitRun(t, dir, "add", "dist/keep.js")
	gitRun(t, dir, "commit", "-m", "vendor bundles")

	writeFileInDir(t, dir, "dist/bundle.js", "v2\n")
	writeFileInDir(t, dir, "dist/keep.js", "v2\n")
	writeFileInDir(t, dir, "src/app.go", "package app\n")

	err := stageFiles(dir, StageAll, verify.CommitConfig{})
	require.NoError(t, err)

	staged := mustStagedFiles(t, dir)
	assert.NotContains(t, staged, "dist/bundle.js", "force-tracked .gitignore'd file must not be staged")
	assert.Contains(t, staged, "dist/keep.js", "!-negated file must be staged")
	assert.Contains(t, staged, "src/app.go")
	assert.Contains(t, gitOutput(t, dir, "ls-files"), "dist/bundle.js", "stripped file stays tracked")
}

// TestStageAllPreservesManuallyStagedGitIgnored confirms an explicit `git add`
// overrides .gitignore for that commit: gavel only strips what it stages itself.
func TestStageAllPreservesManuallyStagedGitIgnored(t *testing.T) {
	dir := initCommitRepo(t)
	writeFile(t, dir, ".gitignore", "dist/\n")
	gitRun(t, dir, "add", ".gitignore")
	gitRun(t, dir, "commit", "-m", "ignore dist")

	writeFileInDir(t, dir, "dist/bundle.js", "v1\n")
	gitRun(t, dir, "add", "-f", "dist/bundle.js")
	gitRun(t, dir, "commit", "-m", "vendor bundle")

	writeFileInDir(t, dir, "dist/bundle.js", "v2\n")
	gitRun(t, dir, "add", "-f", "dist/bundle.js") // manual stage before gavel runs

	err := stageFiles(dir, StageAll, verify.CommitConfig{})
	require.NoError(t, err)

	assert.Contains(t, mustStagedFiles(t, dir), "dist/bundle.js", "manually staged file must be preserved")
}

// TestStageAllFiltersGavelGitignore verifies the .gavel.yaml commit.gitignore
// rules are honored at stage time, and that commit.allow overrides them.
func TestStageAllFiltersGavelGitignore(t *testing.T) {
	t.Run("pattern excludes untracked file", func(t *testing.T) {
		dir := initCommitRepo(t)
		writeFile(t, dir, "secret.env", "TOKEN=1\n")
		writeFile(t, dir, "keep.txt", "ok\n")

		err := stageFiles(dir, StageAll, verify.CommitConfig{GitIgnore: []string{"*.env"}})
		require.NoError(t, err)

		staged := mustStagedFiles(t, dir)
		assert.Contains(t, staged, "keep.txt")
		assert.NotContains(t, staged, "secret.env", "*.env must be filtered by commit.gitignore")
	})

	t.Run("allow re-includes a matched file", func(t *testing.T) {
		dir := initCommitRepo(t)
		writeFile(t, dir, "secret.env", "TOKEN=1\n")

		err := stageFiles(dir, StageAll, verify.CommitConfig{
			GitIgnore: []string{"*.env"},
			Allow:     []string{"secret.env"},
		})
		require.NoError(t, err)

		assert.Contains(t, mustStagedFiles(t, dir), "secret.env", "commit.allow must override commit.gitignore")
	})
}

// TestStageUnstagedExcludesUntracked confirms `unstaged` stages tracked changes
// only and leaves brand-new untracked files alone.
func TestStageUnstagedExcludesUntracked(t *testing.T) {
	dir := initCommitRepo(t)
	writeFile(t, dir, "README.md", "# changed\n")
	writeFile(t, dir, "new.txt", "new\n")

	err := stageFiles(dir, StageUnstaged, verify.CommitConfig{})
	require.NoError(t, err)

	staged := mustStagedFiles(t, dir)
	assert.Contains(t, staged, "README.md")
	assert.NotContains(t, staged, "new.txt", "unstaged mode must not add untracked files")
}

func mustStagedFiles(t *testing.T, dir string) []string {
	t.Helper()
	files, err := stagedFiles(dir)
	require.NoError(t, err)
	return files
}
