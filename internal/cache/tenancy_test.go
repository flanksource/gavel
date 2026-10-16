// Tests that cached violations and linter statistics are isolated per organization:
// one org's scans, findings, executions and debounce state never affect another's.
package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openTenancyDB(t *testing.T) (*gorm.DB, uuid.UUID) {
	t.Helper()
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_analysis_tenancy"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	other := uuid.New()
	require.NoError(t, db.Gorm().Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, other, other.String()).Error)
	return db.Gorm(), other
}

func TestViolationCacheIsolatesOrganizations(t *testing.T) {
	db, otherOrg := openTenancyDB(t)
	local, err := NewViolationCache(db)
	require.NoError(t, err)
	require.Equal(t, tenant.LocalOrgID, local.org)
	other := local.ForOrg(otherOrg)

	file := filepath.Join(t.TempDir(), "x.go")
	require.NoError(t, os.WriteFile(file, []byte("package x\n"), 0o600))
	message := "local finding"
	require.NoError(t, local.StoreViolations(file, []models.Violation{{Line: 1, Source: "lint", Message: &message, Severity: models.SeverityError}}))

	needs, err := other.NeedsRescan(file)
	require.NoError(t, err)
	assert.True(t, needs, "another org has never scanned the file")
	for name, read := range map[string]func(*ViolationCache) ([]models.Violation, error){
		"by file":    func(c *ViolationCache) ([]models.Violation, error) { return c.GetCachedViolations(file) },
		"all":        func(c *ViolationCache) ([]models.Violation, error) { return c.GetAllViolations() },
		"by source":  func(c *ViolationCache) ([]models.Violation, error) { return c.GetViolationsBySource("lint") },
		"by sources": func(c *ViolationCache) ([]models.Violation, error) { return c.GetViolationsBySources([]string{"lint"}) },
	} {
		found, err := read(other)
		require.NoError(t, err, name)
		assert.Empty(t, found, name)
		found, err = read(local)
		require.NoError(t, err, name)
		assert.Len(t, found, 1, name)
	}
	files, err := other.GetAllCachedFiles()
	require.NoError(t, err)
	assert.Empty(t, files)
	stats, err := other.GetStats()
	require.NoError(t, err)
	assert.EqualValues(t, 0, stats["cached_files"])
	assert.EqualValues(t, 0, stats["total_violations"])

	otherMessage := "other finding"
	require.NoError(t, other.StoreViolations(file, []models.Violation{
		{Line: 2, Source: "lint", Message: &otherMessage, Severity: models.SeverityError},
		{Line: 3, Source: "lint", Message: &otherMessage, Severity: models.SeverityError},
	}))
	found, err := local.GetCachedViolations(file)
	require.NoError(t, err)
	require.Len(t, found, 1, "another org's store must not replace this org's scan")
	assert.Equal(t, "local finding", *found[0].Message)

	require.NoError(t, other.ClearFileCache([]string{file}))
	require.NoError(t, other.ClearCache())
	deleted, err := other.ClearViolations(time.Time{}, "")
	require.NoError(t, err)
	assert.Zero(t, deleted)
	found, err = local.GetAllViolations()
	require.NoError(t, err)
	assert.Len(t, found, 1, "another org's clears must not touch this org's rows")

	require.NoError(t, other.StoreViolations(file, []models.Violation{{Line: 2, Source: "lint", Severity: models.SeverityError}}))
	deleted, err = local.ClearViolations(time.Time{}, "")
	require.NoError(t, err)
	assert.EqualValues(t, 1, deleted)
	found, err = other.GetAllViolations()
	require.NoError(t, err)
	assert.Len(t, found, 1)
}

func TestLinterStatsIsolateOrganizations(t *testing.T) {
	db, otherOrg := openTenancyDB(t)
	local, err := NewLinterStats(db)
	require.NoError(t, err)
	require.Equal(t, tenant.LocalOrgID, local.org)
	other := local.ForOrg(otherOrg)

	require.NoError(t, local.RecordExecution("golangci-lint", "/ws", 250*time.Millisecond, 1, true))

	skip, _, err := other.ShouldSkipLinter("golangci-lint", "/ws", "1h")
	require.NoError(t, err)
	assert.False(t, skip, "another org has never run the linter")
	stats, err := other.GetStats("golangci-lint", "/ws")
	require.NoError(t, err)
	assert.Zero(t, stats.RunCount)
	debounce, err := other.GetIntelligentDebounce("golangci-lint", "/ws")
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, debounce, "no history means the default debounce")
	linters, err := other.GetLinterHistory("/ws")
	require.NoError(t, err)
	assert.Empty(t, linters)

	require.NoError(t, other.RecordExecution("golangci-lint", "/ws", 2*time.Second, 0, true))
	require.NoError(t, other.RecordExecution("golangci-lint", "/ws", 2*time.Second, 0, true))
	stats, err = other.GetStats("golangci-lint", "/ws")
	require.NoError(t, err)
	assert.EqualValues(t, 2, stats.RunCount)
	stats, err = local.GetStats("golangci-lint", "/ws")
	require.NoError(t, err)
	assert.EqualValues(t, 1, stats.RunCount)

	var counters []int
	require.NoError(t, db.Raw(`SELECT consecutive_no_violations FROM debounce_metadata WHERE linter_name = 'golangci-lint' AND work_dir = '/ws' ORDER BY org_id = ? DESC`, tenant.LocalOrgID).Scan(&counters).Error)
	assert.Equal(t, []int{0, 2}, counters, "each org keeps its own debounce state")
}
