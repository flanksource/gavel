package cache

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	internalcache "github.com/flanksource/gavel/internal/cache"
	"gorm.io/gorm"
)

const (
	// EnvDSN is the postgres connection string used by the GitHub cache.
	// When unset (or empty) the cache is disabled and Open returns a no-op
	// store, so the CLI continues to work without postgres available.
	EnvDSN = "GAVEL_GITHUB_CACHE_DSN"
	// EnvDisable forces the cache off regardless of EnvDSN.
	EnvDisable = "GAVEL_GITHUB_CACHE"
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
	db       *internalcache.DB
	disabled bool
	writeMu  sync.Mutex
}

// Open initializes the GitHub cache store. It returns a disabled (but
// non-nil) Store when GAVEL_GITHUB_CACHE=off or when GAVEL_GITHUB_CACHE_DSN
// is not set, so callers can use the result unconditionally.
func Open() (*Store, error) {
	if os.Getenv(EnvDisable) == "off" {
		logger.Debugf("github cache disabled via %s=off", EnvDisable)
		return &Store{disabled: true}, nil
	}
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		logger.Debugf("github cache disabled: %s not set", EnvDSN)
		return &Store{disabled: true}, nil
	}

	db, err := internalcache.NewDBRaw("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open github cache: %w", err)
	}

	if err := db.GormDB().AutoMigrate(
		&HTTPCacheEntry{},
		&WorkflowRunCache{},
		&JobLogCache{},
		&WorkflowDefCache{},
		&SeenPR{},
	); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate github cache: %w", err)
	}

	logger.Debugf("github cache ready (postgres)")
	s := &Store{db: db}
	s.pruneOnOpen()
	return s, nil
}

// Disabled reports whether the store is a no-op pass-through.
func (s *Store) Disabled() bool {
	return s == nil || s.disabled
}

var (
	sharedStore     *Store
	sharedStoreOnce sync.Once
)

// Shared returns a process-wide Store, lazily opened on first access. On
// open failure we return a disabled store and log the error — the CLI keeps
// working, just without caching.
func Shared() *Store {
	sharedStoreOnce.Do(func() {
		s, err := Open()
		if err != nil {
			logger.Warnf("github cache unavailable: %v", err)
			sharedStore = &Store{disabled: true}
			return
		}
		sharedStore = s
	})
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
	return s.db.GormDB()
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

// Status describes the runtime state of the GitHub cache. Exposed via the
// PR UI's /api/activity/cache endpoint so operators can confirm whether
// caching is actually active and see how much is being stored.
type Status struct {
	Enabled      bool              `json:"enabled"`
	Driver       string            `json:"driver"`
	DSNSource    string            `json:"dsnSource"`    // env var name, blank if disabled
	DSNMasked    string            `json:"dsnMasked"`    // DSN with credentials redacted
	RetentionSec int64             `json:"retentionSec"` // http_cache_entries prune horizon
	Counts       map[string]int64  `json:"counts"`       // rows per table
	Error        string            `json:"error,omitempty"`
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

	dsn := os.Getenv(EnvDSN)
	if dsn != "" {
		st.DSNSource = EnvDSN
		st.DSNMasked = maskDSN(dsn)
	}
	if os.Getenv(EnvDisable) == "off" {
		st.Error = EnvDisable + "=off"
		return st
	}
	if !st.Enabled {
		if dsn == "" {
			st.Error = EnvDSN + " not set"
		}
		return st
	}

	tables := []struct {
		name  string
		model any
	}{
		{"http_cache_entries", &HTTPCacheEntry{}},
		{"workflow_run_caches", &WorkflowRunCache{}},
		{"job_log_caches", &JobLogCache{}},
		{"workflow_def_caches", &WorkflowDefCache{}},
		{"seen_prs", &SeenPR{}},
	}
	for _, t := range tables {
		var n int64
		if err := s.gorm().Model(t.model).Count(&n).Error; err != nil {
			st.Error = fmt.Sprintf("count %s: %v", t.name, err)
			continue
		}
		st.Counts[t.name] = n
	}
	return st
}

// maskDSN hides the password in a postgres DSN so it can be surfaced in the
// UI without leaking credentials. Handles both URL form (postgres://u:p@h/d)
// and key-value form (host=... password=...).
func maskDSN(dsn string) string {
	// URL form: postgres://user:password@host/db
	if strings.Contains(dsn, "://") {
		if at := strings.LastIndex(dsn, "@"); at > 0 {
			if colon := strings.Index(dsn[:at], "://"); colon >= 0 {
				schemeEnd := colon + 3
				creds := dsn[schemeEnd:at]
				if user, _, ok := strings.Cut(creds, ":"); ok {
					return dsn[:schemeEnd] + user + ":REDACTED" + dsn[at:]
				}
			}
		}
		return dsn
	}
	// Key-value form: host=... password=... dbname=...
	parts := strings.Fields(dsn)
	for i, p := range parts {
		if strings.HasPrefix(p, "password=") {
			parts[i] = "password=REDACTED"
		}
	}
	return strings.Join(parts, " ")
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
