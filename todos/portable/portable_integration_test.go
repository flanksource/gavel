package portable_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/portable"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortablePostgreSQLImportExportRoundTrip(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres portable TODO tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: "gavel_portable_todos",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })
	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())

	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	repository, err := native.NewRepository(opened.Gorm())
	require.NoError(t, err)
	workDir := t.TempDir()
	workspace, err := repository.CreateWorkspace(t.Context(), native.CreateWorkspaceInput{
		RepoKey: "github.com/example/portable", RootPath: workDir, DisplayName: "Portable",
	})
	require.NoError(t, err)

	id := uuid.MustParse("11111111-2222-4333-8444-555555555555")
	frontmatter := types.TODOFrontmatter{
		Title: "Portable DB round trip", Priority: types.PriorityHigh, Status: types.StatusPending,
	}
	frontmatter.Metadata = map[string]any{
		"id": id.String(), "labels": []string{"database", "portable"},
	}
	body := "Description.\n\n## Acceptance Criteria\n\n- [ ] Round trip\n\n## Verification\n\n```bash\ntrue\n```"
	content, err := todos.WriteFrontmatter(&frontmatter, "\n"+body+"\n")
	require.NoError(t, err)
	inputDir := filepath.Join(workDir, ".todos")
	require.NoError(t, os.MkdirAll(inputDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "portable.md"), []byte(content), 0o644))

	workspaceOptions := todoruntime.WorkspaceOptions{
		Name: "portable", RootPath: workDir, Repositories: []string{"acme/portable"},
	}
	first, err := portable.Import(t.Context(), opened.Gorm(), workspaceOptions, inputDir, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, first.Created)
	issue, err := repository.GetIssue(t.Context(), id)
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, issue.WorkspaceID)
	assert.Equal(t, "Portable DB round trip", issue.Title)
	assert.Equal(t, body, issue.Body)
	assert.Equal(t, "```bash\ntrue\n```", issue.Verification)
	assert.ElementsMatch(t, []string{"database", "portable"}, issue.Labels)

	outputDir := filepath.Join(workDir, "exported")
	exported, err := portable.Export(t.Context(), opened.Gorm(), workspaceOptions, outputDir, []string{id.String()}, false)
	require.NoError(t, err)
	require.Len(t, exported.Files, 1)
	exportedTODO, err := todos.ParseTODO(exported.Files[0])
	require.NoError(t, err)
	assert.Equal(t, "Portable DB round trip", exportedTODO.Title)
	assert.Equal(t, types.PriorityHigh, exportedTODO.Priority)
	assert.Equal(t, types.StatusPending, exportedTODO.Status)
	assert.Equal(t, body, strings.TrimSpace(exportedTODO.MarkdownBody))

	replayed, err := portable.Import(t.Context(), opened.Gorm(), workspaceOptions, outputDir, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, replayed.Created)
	assert.Equal(t, 0, replayed.Updated)
	assert.Equal(t, 1, replayed.Unchanged)
}
