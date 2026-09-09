package native_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/gavel/todos/native"
)

func TestLabelDefinitionStorage(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()

	one, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/one"})
	require.NoError(t, err)
	two, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/two"})
	require.NoError(t, err)

	t.Run("upsert is idempotent for a workspace row", func(t *testing.T) {
		first, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{
			WorkspaceID: &one.ID, Name: "bug", Color: "red", Icon: "debug", Description: "broken",
		})
		require.NoError(t, err)
		assert.Equal(t, "red", first.Color)

		second, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{
			WorkspaceID: &one.ID, Name: "bug", Color: "rose", Description: "still broken",
		})
		require.NoError(t, err)
		assert.Equal(t, first.ID, second.ID, "a second save must update, not insert a duplicate")
		assert.Equal(t, "rose", second.Color)
		assert.Equal(t, "", second.Icon, "an omitted icon clears the previous one")
	})

	t.Run("upsert is idempotent for a global row", func(t *testing.T) {
		first, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "flaky", Color: "amber"})
		require.NoError(t, err)
		require.Nil(t, first.WorkspaceID)

		second, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "flaky", Color: "teal"})
		require.NoError(t, err)
		assert.Equal(t, first.ID, second.ID)
		assert.Equal(t, "teal", second.Color)
	})

	t.Run("a workspace row and a global row may share a name", func(t *testing.T) {
		_, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "shared", Color: "sky"})
		require.NoError(t, err)
		scoped, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{
			WorkspaceID: &one.ID, Name: "shared", Color: "violet",
		})
		require.NoError(t, err)
		assert.Equal(t, "violet", scoped.Color)
	})

	t.Run("the same name may be defined in two workspaces", func(t *testing.T) {
		_, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{
			WorkspaceID: &two.ID, Name: "bug", Color: "lime",
		})
		require.NoError(t, err)
	})

	t.Run("names and colors are normalized on write", func(t *testing.T) {
		row, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{
			WorkspaceID: &one.ID, Name: "  MiXeD  ", Color: "  SKY  ",
		})
		require.NoError(t, err)
		assert.Equal(t, "mixed", row.Name)
		assert.Equal(t, "sky", row.Color)
	})

	t.Run("invalid input is rejected before it reaches the database", func(t *testing.T) {
		for name, input := range map[string]native.LabelDefinitionInput{
			"empty name":    {Name: "   ", Color: "red"},
			"off-palette":   {Name: "x", Color: "chartreuse"},
			"missing color": {Name: "x"},
			"unknown icon":  {Name: "x", Color: "red", Icon: "no-such-glyph"},
			"nil workspace": {WorkspaceID: &uuid.Nil, Name: "x", Color: "red"},
		} {
			t.Run(name, func(t *testing.T) {
				_, err := repo.SetLabelDefinition(ctx, input)
				require.ErrorIs(t, err, native.ErrInvalidInput)
			})
		}
	})

	t.Run("list returns the workspace rows plus every global row", func(t *testing.T) {
		rows, err := repo.ListLabelDefinitions(ctx, one.ID)
		require.NoError(t, err)

		byName := map[string]native.LabelDefinition{}
		for _, row := range rows {
			byName[row.Name] = row
		}
		assert.Equal(t, "rose", byName["bug"].Color, "workspace one's own row")
		assert.Equal(t, "violet", byName["shared"].Color, "the workspace row, not the global")
		assert.Contains(t, byName, "flaky", "global rows are included")

		for _, row := range rows {
			if row.WorkspaceID != nil {
				assert.Equal(t, one.ID, *row.WorkspaceID, "another workspace's rows must not leak in")
			}
		}
	})

	t.Run("global list excludes scoped rows", func(t *testing.T) {
		rows, err := repo.ListGlobalLabelDefinitions(ctx)
		require.NoError(t, err)
		for _, row := range rows {
			assert.Nil(t, row.WorkspaceID)
		}
	})

	t.Run("deleting an override leaves the global row", func(t *testing.T) {
		_, err := repo.DeleteLabelDefinition(ctx, &one.ID, "shared")
		require.NoError(t, err)

		rows, err := repo.ListLabelDefinitions(ctx, one.ID)
		require.NoError(t, err)
		for _, row := range rows {
			if row.Name == "shared" {
				assert.Nil(t, row.WorkspaceID, "only the global row should remain")
			}
		}
	})

	t.Run("deleting an absent definition is a loud miss", func(t *testing.T) {
		_, err := repo.DeleteLabelDefinition(ctx, &one.ID, "never-defined")
		require.ErrorIs(t, err, native.ErrLabelNotFound)
	})

	t.Run("a workspace-scoped delete does not remove the global row of the same name", func(t *testing.T) {
		_, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "both", Color: "sky"})
		require.NoError(t, err)
		_, err = repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{
			WorkspaceID: &one.ID, Name: "both", Color: "pink",
		})
		require.NoError(t, err)

		_, err = repo.DeleteLabelDefinition(ctx, &one.ID, "both")
		require.NoError(t, err)

		globals, err := repo.ListGlobalLabelDefinitions(ctx)
		require.NoError(t, err)
		names := map[string]bool{}
		for _, row := range globals {
			names[row.Name] = true
		}
		assert.True(t, names["both"], "the global row must survive a workspace-scoped delete")
	})
}

// TestRemovingAProjectLabelStripsItFromTodos covers the asymmetry between the
// two scopes: a project removal retires the label from the backlog, a global
// removal is presentation only.
func TestRemovingAProjectLabelStripsItFromTodos(t *testing.T) {
	repo, gorm := openRepository(t)
	ctx := t.Context()

	workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/retire"})
	require.NoError(t, err)
	other, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/spectator"})
	require.NoError(t, err)

	create := func(workspaceID uuid.UUID, title string, labels ...string) *native.Issue {
		t.Helper()
		issue, err := repo.CreateIssue(ctx, native.CreateIssueInput{
			WorkspaceID: workspaceID, Title: title, Labels: labels,
		})
		require.NoError(t, err)
		return issue
	}

	tagged := create(workspace.ID, "tagged", "stale", "bug")
	alsoTagged := create(workspace.ID, "also tagged", "stale")
	untouched := create(workspace.ID, "untouched", "bug")
	elsewhere := create(other.ID, "elsewhere", "stale")

	_, err = repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{
		WorkspaceID: &workspace.ID, Name: "stale", Color: "amber",
	})
	require.NoError(t, err)

	removal, err := repo.DeleteLabelDefinition(ctx, &workspace.ID, "stale")
	require.NoError(t, err)

	t.Run("the removal reports both halves of what it did", func(t *testing.T) {
		assert.True(t, removal.Definition)
		assert.Equal(t, int64(2), removal.Todos)
	})

	t.Run("the label is gone from this workspace's todos", func(t *testing.T) {
		for _, issue := range []*native.Issue{tagged, alsoTagged} {
			reloaded, err := repo.GetIssue(ctx, issue.ID)
			require.NoError(t, err)
			assert.NotContains(t, reloaded.Labels, "stale")
		}
	})

	t.Run("other labels on the same todo survive", func(t *testing.T) {
		reloaded, err := repo.GetIssue(ctx, tagged.ID)
		require.NoError(t, err)
		assert.Equal(t, []string{"bug"}, reloaded.Labels)
	})

	t.Run("todos that never carried it are not rewritten", func(t *testing.T) {
		reloaded, err := repo.GetIssue(ctx, untouched.ID)
		require.NoError(t, err)
		assert.Equal(t, untouched.Version, reloaded.Version, "an untouched issue must not be bumped")
	})

	t.Run("another workspace keeps its own label", func(t *testing.T) {
		reloaded, err := repo.GetIssue(ctx, elsewhere.ID)
		require.NoError(t, err)
		assert.Contains(t, reloaded.Labels, "stale")
	})

	t.Run("the strip bumps the version so stale writers conflict", func(t *testing.T) {
		reloaded, err := repo.GetIssue(ctx, tagged.ID)
		require.NoError(t, err)
		assert.Equal(t, tagged.Version+1, reloaded.Version)
	})

	t.Run("history records the edit rather than the label vanishing", func(t *testing.T) {
		var events []struct {
			Kind    string `gorm:"column:kind"`
			Payload string `gorm:"column:payload"`
		}
		require.NoError(t, gorm.Raw(`
			SELECT kind, payload::text AS payload
			FROM todo_issue_events WHERE issue_id = ? ORDER BY sequence DESC LIMIT 1`,
			tagged.ID).Scan(&events).Error)
		require.Len(t, events, 1)
		assert.Equal(t, "updated", events[0].Kind)
		assert.Contains(t, events[0].Payload, "bug")
		assert.NotContains(t, events[0].Payload, "stale")
	})

	t.Run("a label carried by todos but never defined can still be retired", func(t *testing.T) {
		removal, err := repo.DeleteLabelDefinition(ctx, &workspace.ID, "bug")
		require.NoError(t, err)
		assert.False(t, removal.Definition, "'bug' resolves through a built-in, so no row exists")
		assert.Equal(t, int64(2), removal.Todos)
	})

	t.Run("a global removal never touches todo content", func(t *testing.T) {
		_, err := repo.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "shared-hue", Color: "teal"})
		require.NoError(t, err)
		carrier := create(workspace.ID, "carries the global label", "shared-hue")

		removal, err := repo.DeleteLabelDefinition(ctx, nil, "shared-hue")
		require.NoError(t, err)
		assert.True(t, removal.Definition)
		assert.Zero(t, removal.Todos)

		reloaded, err := repo.GetIssue(ctx, carrier.ID)
		require.NoError(t, err)
		assert.Contains(t, reloaded.Labels, "shared-hue",
			"a global scope spans every project, so removing it must not delete content")
	})
}

func TestCountIssuesByLabel(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()

	workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/counts"})
	require.NoError(t, err)
	other, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/other"})
	require.NoError(t, err)

	createLabelled := func(workspaceID uuid.UUID, title string, labels ...string) {
		t.Helper()
		_, err := repo.CreateIssue(ctx, native.CreateIssueInput{
			WorkspaceID: workspaceID, Title: title, Labels: labels,
		})
		require.NoError(t, err)
	}

	createLabelled(workspace.ID, "one", "bug", "ui")
	createLabelled(workspace.ID, "two", "bug")
	createLabelled(workspace.ID, "three", "undefined-label")
	createLabelled(other.ID, "elsewhere", "bug")

	counts, err := repo.CountIssuesByLabel(ctx, workspace.ID)
	require.NoError(t, err)

	assert.Equal(t, 2, counts["bug"], "another workspace's issue must not be counted")
	assert.Equal(t, 1, counts["ui"])
	assert.Equal(t, 1, counts["undefined-label"], "labels with no definition still count")
	assert.NotContains(t, counts, "missing")
}
