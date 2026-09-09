package native_test

import (
	"testing"

	"github.com/flanksource/gavel/todos/native"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateWorkspaceRepointsThePrimaryPath pins the one thing a workspace's
// root path has to do: name where the workspace is *now*. Moving a checkout
// re-registers it, and every later lookup — a deep link resolving a todo it has
// no workspace for, most of all — has to answer with the new location, not the
// one the workspace was first created at.
func TestUpdateWorkspaceRepointsThePrimaryPath(t *testing.T) {
	repo, db := openRepository(t)

	oldRoot := t.TempDir()
	newRoot := t.TempDir()

	workspace, err := repo.CreateWorkspace(t.Context(), native.CreateWorkspaceInput{
		RepoKey: "github.com/example/moved", RootPath: oldRoot, DisplayName: "Moved",
	})
	require.NoError(t, err)
	require.Equal(t, oldRoot, workspace.RootPath)

	updated, err := repo.UpdateWorkspace(t.Context(), workspace.ID, native.UpdateWorkspaceInput{RootPath: &newRoot})
	require.NoError(t, err)
	assert.Equal(t, newRoot, updated.RootPath, "the update's own answer names the new root")

	// Exactly one row may claim the workspace, otherwise the read below joins
	// against whichever of them the planner happens to return first.
	var primaries []string
	require.NoError(t, db.WithContext(t.Context()).Raw(
		`SELECT path FROM todo_workspace_paths WHERE workspace_id = ? AND is_primary ORDER BY path`,
		workspace.ID,
	).Scan(&primaries).Error)
	assert.Equal(t, []string{newRoot}, primaries, "the former root is retained but demoted")

	reread, err := repo.GetWorkspace(t.Context(), workspace.ID)
	require.NoError(t, err)
	assert.Equal(t, newRoot, reread.RootPath, "a re-read resolves the workspace to its current root")

	// The retained path still resolves the workspace; it just no longer names it.
	byOldPath, err := repo.GetWorkspaceByPath(t.Context(), oldRoot)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, byOldPath.ID)
	assert.Equal(t, newRoot, byOldPath.RootPath)
}
