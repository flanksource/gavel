package snapshots

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadClean(t *testing.T) {
	repo := initRepo(t)

	snap := &testui.Snapshot{
		Status: testui.SnapshotStatus{LintRun: true},
	}
	path, err := Save(repo, snap)
	require.NoError(t, err)
	require.FileExists(t, path)
	assert.NotContains(t, filepath.Base(path), "-dirty-")

	pointer, err := LoadPointer(repo, PointerLast)
	require.NoError(t, err)
	require.NotNil(t, pointer)
	assert.Empty(t, pointer.Uncommitted)
	assert.NotEmpty(t, pointer.SHA)
	assert.Equal(t, filepath.Join(Dir, filepath.Base(path)), pointer.Path)

	loaded, err := LoadByPointer(repo, pointer)
	require.NoError(t, err)
	assert.True(t, loaded.Status.LintRun)

	// Branch pointer must exist and match.
	branch := branchPointerName(repo)
	require.NotEmpty(t, branch)
	bp, err := LoadPointer(repo, branch)
	require.NoError(t, err)
	assert.Equal(t, pointer.Path, bp.Path)
}

func TestSaveAndLoadUncommitted(t *testing.T) {
	repo := initRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(repo, "new.txt"), []byte("uncommitted\n"), 0o644))

	snap := &testui.Snapshot{}
	path1, err := Save(repo, snap)
	require.NoError(t, err)
	base := filepath.Base(path1)
	assert.Regexp(t, `^sha-[0-9a-f]{40}-[0-9a-f]{8}\.json$`, base)

	pointer, err := LoadPointer(repo, PointerLast)
	require.NoError(t, err)
	require.NotNil(t, pointer)
	assert.NotEmpty(t, pointer.Uncommitted)

	// Re-saving with the same uncommitted state must reuse the same file.
	path2, err := Save(repo, snap)
	require.NoError(t, err)
	assert.Equal(t, path1, path2)

	// Change uncommitted state → new snapshot file.
	require.NoError(t, os.WriteFile(filepath.Join(repo, "new.txt"), []byte("different\n"), 0o644))
	path3, err := Save(repo, snap)
	require.NoError(t, err)
	assert.NotEqual(t, path1, path3)
}

func TestLoadMissingPointer(t *testing.T) {
	repo := initRepo(t)
	p, err := LoadPointer(repo, PointerLast)
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestLoadCorruptPointer(t *testing.T) {
	repo := initRepo(t)
	snap := &testui.Snapshot{}
	_, err := Save(repo, snap)
	require.NoError(t, err)

	pointer, err := LoadPointer(repo, PointerLast)
	require.NoError(t, err)
	require.NotNil(t, pointer)

	// Delete the snapshot file but keep the pointer.
	require.NoError(t, os.Remove(filepath.Join(repo, pointer.Path)))

	_, err = LoadByPointer(repo, pointer)
	require.Error(t, err)
}

func TestSanitiseBranch(t *testing.T) {
	assert.Equal(t, "feat-foo", SanitiseBranch("feat/foo"))
	assert.Equal(t, "feat-foo-bar", SanitiseBranch("feat/foo/bar"))
	assert.Equal(t, "main", SanitiseBranch("main"))
	assert.Equal(t, "detached", SanitiseBranch(""))
	assert.Equal(t, "detached", SanitiseBranch("/"))
}

func TestSnapshotIDStableAcrossIdenticalDirtyStates(t *testing.T) {
	repo := initRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "file.txt"), []byte("content\n"), 0o644))

	sha1, u1, err := SnapshotID(repo)
	require.NoError(t, err)
	assert.NotEmpty(t, u1)

	sha2, u2, err := SnapshotID(repo)
	require.NoError(t, err)
	assert.Equal(t, sha1, sha2)
	assert.Equal(t, u1, u2)
}

func TestPointerRoundTripJSON(t *testing.T) {
	repo := initRepo(t)
	snap := &testui.Snapshot{}
	_, err := Save(repo, snap)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(repo, Dir, "last.json"))
	require.NoError(t, err)

	var p Pointer
	require.NoError(t, json.Unmarshal(data, &p))
	assert.NotEmpty(t, p.SHA)
	assert.NotEmpty(t, p.Path)
}

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")
	run("config", "commit.gpgsign", "false")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# r\n"), 0o644))
	run("add", "README.md")
	run("commit", "-m", "initial")
	return dir
}
