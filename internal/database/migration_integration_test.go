package database_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	commonsdb "github.com/flanksource/commons-db/db"
	githubcache "github.com/flanksource/gavel/github/cache"
	analysiscache "github.com/flanksource/gavel/internal/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHCLMigrationFromExistingGitHubSchema(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres migration tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_migrate",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	legacy, err := commonsdb.NewGorm(dsn, commonsdb.DefaultGormConfig())
	require.NoError(t, err)
	require.NoError(t, legacy.AutoMigrate(
		&githubcache.HTTPCacheEntry{},
		&githubcache.WorkflowRunCache{},
		&githubcache.JobLogCache{},
		&githubcache.WorkflowDefCache{},
		&githubcache.SeenPR{},
		&githubcache.FaviconCache{},
		&githubcache.CommitStatCache{},
		&githubcache.CommitStatCursor{},
		&githubcache.TestRunCache{},
		&githubcache.TestRunCursor{},
	))
	require.NoError(t, legacy.Create(&githubcache.HTTPCacheEntry{URL: "https://example.test", Method: "GET", StatusCode: 200}).Error)
	require.NoError(t, legacy.Exec(`
		CREATE TABLE migration_unmanaged (id integer PRIMARY KEY, value text NOT NULL);
		INSERT INTO migration_unmanaged (id, value) VALUES (1, 'preserve me')
	`).Error)
	legacySQL, err := legacy.DB()
	require.NoError(t, err)
	require.NoError(t, legacySQL.Close())

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())

	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	require.False(t, db.Disabled())
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	for _, table := range []string{
		"captain_sessions", "captain_prompt_runs", "captain_plans", "captain_turns", "captain_model_calls",
		"http_cache_entries", "workflow_run_caches", "job_log_caches", "workflow_def_caches",
		"seen_prs", "favicon_caches",
		"commit_stat_caches", "commit_stat_cursors", "test_run_caches", "test_run_cursors",
		"task_run_history",
		"file_scans", "violations", "linter_executions", "debounce_metadata", "migration_unmanaged",
	} {
		require.True(t, db.Gorm().Migrator().HasTable(table), "%s should exist", table)
	}
	var cached githubcache.HTTPCacheEntry
	require.NoError(t, db.Gorm().First(&cached, "url = ? AND method = ?", "https://example.test", "GET").Error)
	require.Equal(t, 200, cached.StatusCode)
	var unmanagedValue string
	require.NoError(t, db.Gorm().Raw(`SELECT value FROM migration_unmanaged WHERE id = 1`).Scan(&unmanagedValue).Error)
	assert.Equal(t, "preserve me", unmanagedValue)

	second, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err, "migration should be idempotent")
	require.NoError(t, second.Close())

	require.NoError(t, db.Gorm().Exec(`DROP TABLE
		http_cache_entries, workflow_run_caches, job_log_caches, workflow_def_caches,
		seen_prs, favicon_caches,
		commit_stat_caches, commit_stat_cursors, test_run_caches, test_run_cursors,
		violations, file_scans, linter_executions, debounce_metadata CASCADE`).Error)
	fresh, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err, "migration should create the complete schema from scratch")
	require.NoError(t, fresh.Close())
	for _, table := range []string{"captain_sessions", "captain_prompt_runs", "captain_plans", "http_cache_entries", "task_run_history", "violations", "file_scans", "linter_executions", "debounce_metadata", "migration_unmanaged"} {
		require.True(t, db.Gorm().Migrator().HasTable(table), "%s should exist after a fresh migration", table)
	}

	violationCache, err := analysiscache.NewViolationCache(db.Gorm())
	require.NoError(t, err)
	sourceFile := filepath.Join(t.TempDir(), "source.go")
	require.NoError(t, os.WriteFile(sourceFile, []byte("package sample\n"), 0o600))
	message := "forbidden dependency"
	code := "ARCH001"
	require.NoError(t, violationCache.StoreViolations(sourceFile, []models.Violation{{
		Line:      7,
		Column:    3,
		Message:   &message,
		Source:    "architecture",
		Severity:  models.SeverityError,
		Fixable:   true,
		Code:      &code,
		Rule:      &models.Rule{Type: models.RuleTypeDeny, Package: "internal/legacy"},
		CreatedAt: time.Now(),
	}}))
	cachedViolations, err := violationCache.GetCachedViolations(sourceFile)
	require.NoError(t, err)
	require.Len(t, cachedViolations, 1)
	assert.Equal(t, "internal/legacy", cachedViolations[0].Rule.Package)
	assert.False(t, cachedViolations[0].CreatedAt.IsZero())
	stats, err := violationCache.GetStats()
	require.NoError(t, err)
	assert.EqualValues(t, 1, stats["cached_files"])
	assert.EqualValues(t, 1, stats["total_violations"])

	linterStats, err := analysiscache.NewLinterStats(db.Gorm())
	require.NoError(t, err)
	require.NoError(t, linterStats.RecordExecution("golangci-lint", "/workspace", 250*time.Millisecond, 1, false))
	shouldSkip, debounce, err := linterStats.ShouldSkipLinter("golangci-lint", "/workspace", "1m")
	require.NoError(t, err)
	assert.True(t, shouldSkip)
	assert.Equal(t, time.Minute, debounce)
	executionStats, err := linterStats.GetStats("golangci-lint", "/workspace")
	require.NoError(t, err)
	assert.EqualValues(t, 1, executionStats.RunCount)
	assert.EqualValues(t, 1, executionStats.ViolationCount)
}

func TestHCLMigrationFreshDatabaseCreatesCurrentSchema(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres migration tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_migrate_fresh",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())

	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	require.False(t, db.Disabled())
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	for _, table := range []string{
		"captain_sessions", "todo_workspaces", "todo_issues",
		"http_cache_entries", "commit_stat_caches", "test_run_caches",
		"task_run_history",
		"file_scans", "violations", "linter_executions", "debounce_metadata",
	} {
		require.True(t, db.Gorm().Migrator().HasTable(table), "%s should exist on a fresh database", table)
	}
}
