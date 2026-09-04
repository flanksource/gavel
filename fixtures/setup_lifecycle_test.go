package fixtures

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// git runs a git command in dir and fails the test if it does not succeed.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %s: %s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

// seedRepo builds a git repository holding one committed file plus one of each
// state a worktree can carry across: a staged edit, an unstaged edit, an
// untracked file, and a gitignored file.
func seedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	git(t, dir, "config", "user.email", "fixtures@example.com")
	git(t, dir, "config", "user.name", "Fixtures")

	write(t, filepath.Join(dir, ".gitignore"), "ignored.txt\n")
	write(t, filepath.Join(dir, "committed.txt"), "committed\n")
	write(t, filepath.Join(dir, "staged.txt"), "original\n")
	write(t, filepath.Join(dir, "unstaged.txt"), "original\n")
	git(t, dir, "add", ".gitignore", "committed.txt", "staged.txt", "unstaged.txt")
	git(t, dir, "commit", "-qm", "initial")

	write(t, filepath.Join(dir, "staged.txt"), "staged edit\n")
	git(t, dir, "add", "staged.txt")
	write(t, filepath.Join(dir, "unstaged.txt"), "unstaged edit\n")
	write(t, filepath.Join(dir, "untracked.txt"), "untracked\n")
	write(t, filepath.Join(dir, "ignored.txt"), "ignored\n")
	return dir
}

// newRunnerForFixture parses a markdown body written into dir and returns a
// Runner holding its tree, ready for prepareSetups.
func newRunnerForFixture(t *testing.T, dir, body string) *Runner {
	t.Helper()
	path := filepath.Join(dir, "lifecycle.fixture.md")
	write(t, path, strings.TrimSpace(body)+"\n")

	runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: dir})
	require.NoError(t, err)
	_, err = runner.Parse()
	require.NoError(t, err)
	return runner
}

func testContext() flanksourceContext.Context {
	return flanksourceContext.NewContext(context.Background())
}

// A worktree checkout must relocate the file's tests into a tree of their own,
// carry every uncommitted and ignored file across by default, start at the same
// commit — and leave the source repository byte-identical, which is the whole
// reason `dirty`/`stash` was replaced: nothing is ever stashed out of your tree.
func TestPrepareSetupsClonesIntoWorktreeAndLeavesSourceUntouched(t *testing.T) {
	repo := seedRepo(t)
	baseDir := t.TempDir()

	runner := newRunnerForFixture(t, repo, `
---
setup:
  baseDir: `+baseDir+`
  checkout:
    mode: local
    path: .
    worktree:
      mode: new
exec: bash
args: ["-c", "true"]
---

# Worktree

`+"```bash\ntrue\n```"+`
`)

	// Captured after the markdown lands in the repo, so the fixture file's own
	// untracked entry is part of the baseline rather than a false positive.
	before := git(t, repo, "status", "--porcelain")

	require.NoError(t, runner.prepareSetups(testContext()))
	require.Len(t, runner.setups, 1)

	var prepared *PreparedSetup
	for _, ps := range runner.setups {
		prepared = ps
	}
	require.NotEmpty(t, prepared.Cwd)
	assert.NotEqual(t, repo, prepared.Cwd, "worktree did not relocate the run")
	assert.Equal(t, git(t, repo, "rev-parse", "HEAD"), git(t, prepared.Cwd, "rev-parse", "HEAD"),
		"worktree started from a different commit than the source HEAD")

	// Defaults are uncommitted: clone and ignored: clone, so all four states
	// travel — a bare `git worktree add` would give you only committed.txt.
	for name, want := range map[string]string{
		"committed.txt": "committed\n",
		"staged.txt":    "staged edit\n",
		"unstaged.txt":  "unstaged edit\n",
		"untracked.txt": "untracked\n",
		"ignored.txt":   "ignored\n",
	} {
		content, err := os.ReadFile(filepath.Join(prepared.Cwd, name))
		require.NoErrorf(t, err, "%s is missing from the worktree", name)
		assert.Equalf(t, want, string(content), "%s has the wrong content in the worktree", name)
	}

	runner.cleanupSetups()
	assert.NoDirExists(t, prepared.Cwd, "worktree survived cleanup")
	assert.Equal(t, before, git(t, repo, "status", "--porcelain"),
		"preparing a worktree mutated the source repository")
}

// `uncommitted: skip` must give a pristine tree at base, while `ignored` stays
// on its own default — the two knobs are independent.
func TestPrepareSetupsHonoursUncommittedSkip(t *testing.T) {
	repo := seedRepo(t)
	baseDir := t.TempDir()

	runner := newRunnerForFixture(t, repo, `
---
setup:
  baseDir: `+baseDir+`
  checkout:
    mode: local
    path: .
    worktree:
      mode: new
      uncommitted: skip
exec: bash
args: ["-c", "true"]
---

# Pristine

`+"```bash\ntrue\n```"+`
`)

	require.NoError(t, runner.prepareSetups(testContext()))
	defer runner.cleanupSetups()

	var prepared *PreparedSetup
	for _, ps := range runner.setups {
		prepared = ps
	}
	require.NotNil(t, prepared)

	staged, err := os.ReadFile(filepath.Join(prepared.Cwd, "staged.txt"))
	require.NoError(t, err)
	assert.Equal(t, "original\n", string(staged), "uncommitted: skip carried the staged edit anyway")
	assert.NoFileExists(t, filepath.Join(prepared.Cwd, "untracked.txt"))
	assert.FileExists(t, filepath.Join(prepared.Cwd, "ignored.txt"), "ignored: clone is still the default")
}

// A setup that declares no checkout must leave the run where the markdown is.
// commons-db invents a scratch <baseDir>/tmp/<uuid> when Cwd is blank, and it
// defaults baseDir to <sourceDir>/.shell — both would be silent surprises.
func TestPrepareSetupsKeepsCwdAndKeepsBaseDirOutOfTheRepo(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".env.fixture"), "FIXTURE_TOKEN=from-dotenv\n")

	runner := newRunnerForFixture(t, dir, `
---
setup:
  dotenv: [.env.fixture]
exec: bash
args: ["-c", "true"]
---

# Dotenv only

`+"```bash\ntrue\n```"+`
`)

	require.NoError(t, runner.prepareSetups(testContext()))
	defer runner.cleanupSetups()

	var prepared *PreparedSetup
	for _, ps := range runner.setups {
		prepared = ps
	}
	require.NotNil(t, prepared)
	assert.Equal(t, dir, prepared.Cwd, "a checkout-less setup relocated the run")
	assert.Equal(t, "from-dotenv", prepared.Env["FIXTURE_TOKEN"])
	assert.NoDirExists(t, filepath.Join(dir, ".shell"), "commons-db's .shell fallback leaked into the fixture's repo")
}

// Without a database a `connection://` reference fails deep inside commons-db
// as a bare "db is not configured" with nothing tying it to a fixture. Name the
// file and say what is missing.
func TestPrepareSetupsRejectsConnectionRefWithoutDatabase(t *testing.T) {
	dir := t.TempDir()
	runner := newRunnerForFixture(t, dir, `
---
setup:
  connections:
    kubernetes:
      connection: connection://kubernetes/sandbox
exec: bash
args: ["-c", "true"]
---

# Needs a connection

`+"```bash\ntrue\n```"+`
`)

	err := runner.prepareSetups(testContext())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "needs a database")
	assert.Contains(t, err.Error(), "lifecycle.fixture.md", "the error does not name the offending file")
}

// Cleanup is reached twice on a normal run — the runner's defer and the
// shutdown hook — and must not tear the same worktree down twice.
func TestPreparedSetupCleanupRunsOnce(t *testing.T) {
	calls := 0
	prepared := &PreparedSetup{cleanup: func() error { calls++; return nil }}

	require.NoError(t, prepared.Cleanup())
	require.NoError(t, prepared.Cleanup())
	assert.Equal(t, 1, calls)

	var nilSetup *PreparedSetup
	assert.NoError(t, nilSetup.Cleanup())
	assert.Nil(t, nilSetup.Environ())
	assert.Equal(t, "fallback", nilSetup.Dir("fallback"))
}

// build: is run-wide but a setup is per file, so the build must land in the
// prepared tree rather than the runner's WorkDir — otherwise it builds repo A
// while the tests exercise worktree B, and passes.
func TestBuildCommandRunsInThePreparedTree(t *testing.T) {
	repo := seedRepo(t)
	baseDir := t.TempDir()

	runner := newRunnerForFixture(t, repo, `
---
setup:
  baseDir: `+baseDir+`
  checkout:
    mode: local
    path: .
    worktree:
      mode: new
build: pwd > built-in.txt
exec: bash
args: ["-c", "true"]
---

# Build

`+"```bash\ntrue\n```"+`
`)

	ctx := testContext()
	require.NoError(t, runner.prepareSetups(ctx))
	defer runner.cleanupSetups()

	buildCmd, buildSetup := runner.getBuildCommand()
	require.Equal(t, "pwd > built-in.txt", buildCmd)
	require.NotNil(t, buildSetup, "build did not pick up the setup of the file that declared it")
	require.NoError(t, runner.executeBuildCommand(ctx, buildCmd, buildSetup))

	assert.NoFileExists(t, filepath.Join(repo, "built-in.txt"), "build ran in the source repo")
	out, err := os.ReadFile(filepath.Join(buildSetup.Cwd, "built-in.txt"))
	require.NoError(t, err, "build did not run in the prepared tree")
	// Through EvalSymlinks because macOS resolves /var to /private/var, so the
	// shell's own pwd never matches the path git handed back verbatim.
	wantPwd, err := filepath.EvalSymlinks(buildSetup.Cwd)
	require.NoError(t, err)
	assert.Equal(t, wantPwd, strings.TrimSpace(string(out)))
}

// A file with no setup: must stay exactly where it was, with no prepared tree
// and nothing to tear down.
func TestPrepareSetupsIsANoOpWithoutSetup(t *testing.T) {
	dir := t.TempDir()
	runner := newRunnerForFixture(t, dir, `
---
exec: bash
args: ["-c", "true"]
---

# No setup

`+"```bash\ntrue\n```"+`
`)

	require.NoError(t, runner.prepareSetups(testContext()))
	assert.Empty(t, runner.setups)

	buildCmd, buildSetup := runner.getBuildCommand()
	assert.Empty(t, buildCmd)
	assert.Nil(t, buildSetup)
}

func TestSetupBaseDirIsStablePerFileAndDistinctBetweenFiles(t *testing.T) {
	first, err := setupBaseDir("a/one.fixture.md")
	require.NoError(t, err)
	again, err := setupBaseDir("a/one.fixture.md")
	require.NoError(t, err)
	second, err := setupBaseDir("a/two.fixture.md")
	require.NoError(t, err)

	assert.Equal(t, first, again, "the same fixture file got a different base dir on a second run")
	assert.NotEqual(t, first, second, "two fixture files share a base dir and would fight over one worktree")

	cache, err := os.UserCacheDir()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(first, filepath.Join(cache, "gavel", "fixtures")), "base dir escaped the user cache: %s", first)
}
