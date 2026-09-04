package types

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/snapshots"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptyGitRepo is a runner step's workdir: resolveStepWorkDir walks up to the
// git root, so the step needs a real repo to land its artifact in.
func emptyGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "--quiet", dir)
	require.NoError(t, cmd.Run())
	return dir
}

func runnerStepFixture(name, kind, config, dir string) fixtures.FixtureTest {
	return fixtures.FixtureTest{
		Name:       name,
		SourceDir:  dir,
		RunnerStep: &fixtures.RunnerStepSpec{Kind: kind, Config: config},
	}
}

func TestRunLintStepRecordsRunArtifact(t *testing.T) {
	dir := emptyGitRepo(t)

	result := RunLintStep(
		runnerStepFixture("lint the repo", fixtures.RunnerKindLint, "workdir: "+dir, dir),
		fixtures.RunOptions{WorkDir: dir},
	)

	require.NotNil(t, result.Run, "a lint step must record its engine output by reference")
	assert.Equal(t, "lint", result.Run.Kind)
	assert.Equal(t, result.Run.RunID+".json", filepath.Base(result.Run.Path))

	_, err := os.Stat(result.Run.Path)
	require.NoError(t, err, "the .gavel artifact the fixture result points at must exist")

	_, err = os.Stat(filepath.Join(dir, snapshots.Dir, "last.json"))
	assert.True(t, os.IsNotExist(err), "a fixture step must not move the last.json pointer")

	// The existing contract that fixtureFeedback and the CEL predicate read.
	assert.Contains(t, result.Metadata, "violations")
}

func TestRunTestStepRecordsRunArtifactOnEngineCrash(t *testing.T) {
	dir := emptyGitRepo(t)

	// No Go module, no tests: the engine produces no results, and the step must
	// still record why rather than leaving the attempt unexplained.
	result := RunTestStep(
		runnerStepFixture("run the suite", fixtures.RunnerKindTest, "workdir: "+dir+"\npaths: [./...]", dir),
		fixtures.RunOptions{WorkDir: dir},
	)

	require.NotNil(t, result.Run, "an empty or crashed run still records an artifact")
	assert.Equal(t, "test", result.Run.Kind)
	assert.Zero(t, result.Run.Total)

	_, err := os.Stat(result.Run.Path)
	require.NoError(t, err)
	assert.Contains(t, result.Metadata, "summary")
}
