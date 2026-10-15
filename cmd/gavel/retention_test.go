// Tests gavel serve's retention wiring: loading the properties file, and the retention pass that
// runs at startup and then every retention.interval.
package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLoadServeProperties(t *testing.T) {
	t.Cleanup(func() { properties.Set("retention.job_logs", "") })
	previous := defaultPropertiesFile
	t.Cleanup(func() { defaultPropertiesFile = previous })

	t.Run("a missing --properties-file is an error", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "gavel.properties")
		err := loadServeProperties(missing)
		require.Error(t, err)
		assert.Contains(t, err.Error(), missing)
	})

	t.Run("--properties-file is loaded", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "gavel.properties")
		require.NoError(t, os.WriteFile(path, []byte("retention.job_logs=7d\n"), 0o644))
		require.NoError(t, loadServeProperties(path))
		assert.Equal(t, "7d", properties.String("", "retention.job_logs"))
	})

	t.Run("without the default file nothing is loaded", func(t *testing.T) {
		properties.Set("retention.job_logs", "")
		defaultPropertiesFile = filepath.Join(t.TempDir(), "gavel.properties")
		require.NoError(t, loadServeProperties(""))
		assert.Empty(t, properties.String("", "retention.job_logs"))
	})

	t.Run("the default file is loaded when it exists", func(t *testing.T) {
		defaultPropertiesFile = filepath.Join(t.TempDir(), "gavel.properties")
		require.NoError(t, os.WriteFile(defaultPropertiesFile, []byte("retention.job_logs=9d\n"), 0o644))
		require.NoError(t, loadServeProperties(""))
		assert.Equal(t, "9d", properties.String("", "retention.job_logs"))
	})
}

func TestStartRetentionRunsAtStartupAndOnEachInterval(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_serve_retention"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	properties.Set("retention.interval", "200ms")
	t.Cleanup(func() { properties.Set("retention.interval", "") })

	insertOld := func() {
		t.Helper()
		require.NoError(t, db.Exec(`INSERT INTO linter_executions (linter_name, work_dir, executed_at, duration_ms, violation_count, success) VALUES ('x', '/w', ?, 1, 0, true)`,
			time.Now().Add(-90*24*time.Hour)).Error)
	}
	remaining := func() int {
		var n int
		require.NoError(t, db.Raw(`SELECT count(*) FROM linter_executions`).Scan(&n).Error)
		return n
	}
	insertOld()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	require.NoError(t, startRetention(ctx, db, nil))
	require.Eventually(t, func() bool { return remaining() == 0 }, 5*time.Second, 20*time.Millisecond, "the startup pass deletes it")

	insertOld()
	require.Eventually(t, func() bool { return remaining() == 0 }, 5*time.Second, 20*time.Millisecond, "a later pass deletes the next one")
}

func TestStartRetentionReportsAnInvalidInterval(t *testing.T) {
	properties.Set("retention.interval", "hourly")
	t.Cleanup(func() { properties.Set("retention.interval", "") })
	err := startRetention(context.Background(), (*gorm.DB)(nil), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retention.interval=hourly")
}
