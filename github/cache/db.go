package cache

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/service"
	"gorm.io/gorm"
)

const (
	// EnvDatabaseDSN and EnvDatabaseDisable are the preferred Gavel-wide names.
	EnvDatabaseDSN     = database.EnvDSN
	EnvDatabaseDisable = database.EnvDisable
	// EnvDSN and EnvDisable are retained for backwards compatibility.
	EnvDSN     = database.LegacyEnvDSN
	EnvDisable = database.LegacyEnvDisable
	// EnvRetention overrides the default 30-day prune horizon for transient
	// http_cache entries.
	EnvRetention = "GAVEL_GITHUB_CACHE_RETENTION"

	defaultRetention = 30 * 24 * time.Hour
)

// Store wraps the underlying postgres connection. A nil Store (or a Store
// returned in disabled mode) is safe to call — every operation becomes a
// pass-through, so callers don't need to branch on whether caching is
// configured.
type Store struct {
	db       *database.DB
	disabled bool
	writeMu  sync.Mutex
}

// Open initializes the GitHub cache store through the shared Gavel database.
// The shared opener resolves GAVEL_DB_DSN, its legacy GitHub-cache alias, and
// finally ~/.config/gavel/db.json written by `gavel system install`.
//
// When db.json selects mode=embedded we launch a per-user postgres via
// commons-db/db.StartEmbedded and register a shutdown hook so the pr-ui
// daemon cleanly stops it on SIGTERM. With neither env nor db.json configured
// the cache remains disabled so the CLI continues to function — but once a
// mode is configured, failures are surfaced rather than swallowed.
func Open() (*Store, error) {
	db, err := database.Open(context.Background())
	if err != nil {
		return nil, err
	}
	if db.Disabled() {
		return &Store{db: db, disabled: true}, nil
	}

	logger.Debugf("github cache ready (postgres, source=%s)", db.DSNSource())
	s := &Store{db: db}
	s.pruneOnOpen()
	return s, nil
}

func openShared() (*Store, error) {
	db, err := database.Shared(context.Background())
	if err != nil {
		return nil, err
	}
	if db.Disabled() {
		return &Store{db: db, disabled: true}, nil
	}
	s := &Store{db: db}
	s.pruneOnOpen()
	return s, nil
}

// Disabled reports whether the store is a no-op pass-through.
func (s *Store) Disabled() bool {
	return s == nil || s.disabled || s.db == nil || s.db.Disabled()
}

var (
	sharedStore   *Store
	sharedStoreMu sync.Mutex
)

// Shared returns a process-wide Store, lazily opened on first access. On
// open failure we return a disabled store and log the error — the CLI keeps
// working, just without caching.
func Shared() *Store {
	sharedStoreMu.Lock()
	defer sharedStoreMu.Unlock()
	if sharedStore != nil {
		return sharedStore
	}
	s, err := openShared()
	if err != nil {
		logger.Warnf("github cache unavailable: %v", err)
		// Do not cache transient failures: a later caller can retry the shared
		// database open/migration path.
		return &Store{disabled: true}
	}
	sharedStore = s
	return sharedStore
}

// Close releases the underlying database handle.
func (s *Store) Close() error {
	if s == nil || s.disabled {
		return nil
	}
	return s.db.Close()
}

// gorm returns the underlying *gorm.DB. Callers must hold writeMu when
// performing writes; reads are safe without locking thanks to postgres MVCC.
func (s *Store) gorm() *gorm.DB {
	return s.db.Gorm()
}

// retention reads GAVEL_GITHUB_CACHE_RETENTION (a Go duration string) and
// falls back to defaultRetention.
func retention() time.Duration {
	if v := os.Getenv(EnvRetention); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
		logger.Warnf("invalid %s=%q, using default", EnvRetention, v)
	}
	return defaultRetention
}

// Status describes the shared Gavel database and cache tables. It is exposed
// through the PR UI so operators can confirm persistence is active and inspect
// row counts.
type Status struct {
	Enabled      bool             `json:"enabled"`
	Driver       string           `json:"driver"`
	DSNSource    string           `json:"dsnSource"`    // env var name, blank if disabled
	DSNMasked    string           `json:"dsnMasked"`    // DSN with credentials redacted
	RetentionSec int64            `json:"retentionSec"` // http_cache_entries prune horizon
	Counts       map[string]int64 `json:"counts"`       // rows per table
	Error        string           `json:"error,omitempty"`
}

// Status returns a point-in-time snapshot of cache config and row counts.
// Safe to call on a disabled store.
func (s *Store) Status() Status {
	st := Status{
		Enabled:      !s.Disabled(),
		Driver:       "postgres",
		RetentionSec: int64(retention().Seconds()),
		Counts:       map[string]int64{},
	}

	// Prefer the DSN we actually opened with (env var OR db.json — including
	// embedded postgres where the DSN is only known at runtime). Fall back to
	// the env var for a disabled Store so the UI can still show what the user
	// configured even when Open() chose not to connect.
	switch {
	case s != nil && s.db != nil && s.db.DSN() != "":
		st.DSNSource = s.db.DSNSource()
		st.DSNMasked = service.MaskDSN(s.db.DSN())
	case os.Getenv(EnvDatabaseDSN) != "":
		st.DSNSource = EnvDatabaseDSN
		st.DSNMasked = service.MaskDSN(os.Getenv(EnvDatabaseDSN))
	case os.Getenv(EnvDSN) != "":
		st.DSNSource = EnvDSN
		st.DSNMasked = service.MaskDSN(os.Getenv(EnvDSN))
	}
	if s != nil && s.db != nil && s.db.DisableSource() != "" {
		st.Error = s.db.DisableSource() + "=off"
		return st
	}
	if !st.Enabled {
		if st.DSNMasked == "" {
			st.Error = EnvDatabaseDSN + " not set"
		}
		return st
	}

	tables := []string{
		"http_cache_entries",
		"workflow_run_caches",
		"job_log_caches",
		"workflow_def_caches",
		"seen_prs",
		"favicon_caches",
		"grite_issue_caches",
		"grite_sync_cursors",
		"commit_stat_caches",
		"commit_stat_cursors",
		"test_run_caches",
		"test_run_cursors",
		"file_scans",
		"violations",
		"linter_executions",
		"debounce_metadata",
	}
	for _, table := range tables {
		var n int64
		if err := s.gorm().Table(table).Count(&n).Error; err != nil {
			st.Error = fmt.Sprintf("count %s: %v", table, err)
			continue
		}
		st.Counts[table] = n
	}
	return st
}

// pruneOnOpen drops http_cache_entries older than the retention horizon.
// We don't prune the immutable caches (workflow runs, job logs, workflow
// definitions) — those are addressable by ID/SHA forever and represent
// data that may have been GC'd from GitHub itself.
func (s *Store) pruneOnOpen() {
	cutoff := time.Now().Add(-retention())
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	res := s.gorm().Where("fetched_at < ?", cutoff).Delete(&HTTPCacheEntry{})
	if res.Error != nil {
		logger.Warnf("github cache prune failed: %v", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		logger.Debugf("github cache pruned %d stale http entries", res.RowsAffected)
	}
}
