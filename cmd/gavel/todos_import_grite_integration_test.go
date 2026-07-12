package main

import (
	"os"
	"path/filepath"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/migrategrite"
	"github.com/flanksource/gavel/todos/native"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefuseExistingGriteImportWithoutBaselineDetectsAliasOnly(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres Grite CLI tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: "gavel_import_grite_cli",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })
	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })

	workspace := native.NormalizeImportWorkspace(native.ImportWorkspace{
		RepoKey: " GitHub.COM/Flanksource/Gavel-Alias-Only ", RootPath: "/tmp/gavel-alias-only", DisplayName: " alias only ",
	})
	workspace.ID = native.DeterministicImportWorkspaceID(workspace)
	snapshot := griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 2, EventCount: 1},
		Issues: []griteexport.Issue{{
			IssueID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Title: "Alias evidence", State: "open", CreatedTS: 1, UpdatedTS: 1,
		}},
		Events: []griteexport.Event{{
			EventID: "alias-created", IssueID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TimestampMS: 1,
			Kind: griteexport.Kind{"IssueCreated": []byte(`{"title":"Alias evidence","body":"body"}`)},
		}},
	}
	service, err := migrategrite.NewService(opened.Gorm())
	require.NoError(t, err)
	report, err := service.Import(t.Context(), snapshot, migrategrite.ImportOptions{Workspace: workspace})
	require.NoError(t, err)
	assert.Equal(t, workspace.ID, report.WorkspaceID)

	require.NoError(t, opened.Gorm().Exec(`DELETE FROM todo_issue_events WHERE source = ?`, native.DefaultImportSource).Error)
	var eventCount, aliasCount int64
	require.NoError(t, opened.Gorm().Table("todo_issue_events").Where("source = ?", native.DefaultImportSource).Count(&eventCount).Error)
	require.NoError(t, opened.Gorm().Table("todo_issue_aliases").Where("workspace_id = ? AND kind = 'grite'", workspace.ID).Count(&aliasCount).Error)
	assert.Zero(t, eventCount)
	assert.EqualValues(t, 1, aliasCount)
	require.ErrorContains(t, refuseExistingGriteImportWithoutBaseline(t.Context(), opened.Gorm(), workspace.ID), "Grite aliases")
}
