package cache

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/flanksource/clicky/shutdown"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/commons/logger"
	internalcache "github.com/flanksource/gavel/internal/cache"
	"github.com/flanksource/gavel/service"
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
	db        *internalcache.DB
	disabled  bool
	writeMu   sync.Mutex
	dsn       string // resolved DSN, captured for Status() reporting
	dsnSource string // EnvDSN | "db.json (dsn)" | "db.json (embedded)"
}

// Open initializes the GitHub cache store. It resolves the DSN in this order:
//  1. $GAVEL_GITHUB_CACHE_DSN (unchanged fast path)
//  2. ~/.config/gavel/db.json written by `gavel system install`
//
// When db.json selects mode=embedded we launch a per-user postgres via
// commons-db/db.StartEmbedded and register a shutdown hook so the pr-ui
// daemon cleanly stops it on SIGTERM. With neither env nor db.json configured
// the cache remains disabled so the CLI continues to function — but once a
// mode is configured, failures are surfaced rather than swallowed.
func Open() (*Store, error) {
	if os.Getenv(EnvDisable) == "off" {
		logger.Debugf("github cache disabled via %s=off", EnvDisable)
		return &Store{disabled: true}, nil
	}
	var (
		dsn       = os.Getenv(EnvDSN)
		dsnSource string
	)
	if dsn != "" {
		dsnSource = EnvDSN
	} else {
		resolved, src, err := resolveConfiguredDSN()
		if err != nil {
			return nil, err
		}
		dsn = resolved
		dsnSource = src
	}
	if dsn == "" {
		logger.Debugf("github cache disabled: %s not set and no db config", EnvDSN)
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
		&FaviconCache{},
	); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate github cache: %w", err)
	}

	logger.Debugf("github cache ready (postgres, source=%s)", dsnSource)
	s := &Store{db: db, dsn: dsn, dsnSource: dsnSource}
	s.pruneOnOpen()
	return s, nil
}

// resolveConfiguredDSN reads ~/.config/gavel/db.json. For mode=dsn it returns
// the stored DSN. For mode=embedded it starts a managed postgres via
// commons-db and registers a shutdown hook that stops it on daemon exit.
// Returns ("", "", nil) when no config is present — caller treats that as
// "cache stays disabled". The second return value names the source so the
// UI can show "embedded postgres" vs "db.json DSN" in the status panel.
func resolveConfiguredDSN() (string, string, error) {
	cfg, err := service.LoadDBConfig()
	if err != nil {
		return "", "", fmt.Errorf("load db config: %w", err)
	}
	switch cfg.Mode {
	case "":
		return "", "", nil
	case service.DBModeDSN:
		if cfg.DSN == "" {
			return "", "", fmt.Errorf("db config: mode=dsn but DSN is empty")
		}
		return cfg.DSN, "db.json (dsn)", nil
	case service.DBModeEmbedded:
		dataDir, err := service.EmbeddedDataDir()
		if err != nil {
			return "", "", err
		}
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  dataDir,
			Database: "gavel",
		})
		if err != nil {
			return "", "", fmt.Errorf("start embedded postgres: %w", err)
		}
		shutdown.AddHook("embedded-postgres", func() {
			if err := stop(); err != nil {
				logger.Warnf("stop embedded postgres: %v", err)
			}
		})
		return dsn, "db.json (embedded)", nil
	default:
		return "", "", fmt.Errorf("unknown db mode %q (expected %q or %q)", cfg.Mode, service.DBModeDSN, service.DBModeEmbedded)
	}
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
	case s.dsn != "":
		st.DSNSource = s.dsnSource
		st.DSNMasked = service.MaskDSN(s.dsn)
	case os.Getenv(EnvDSN) != "":
		st.DSNSource = EnvDSN
		st.DSNMasked = service.MaskDSN(os.Getenv(EnvDSN))
	}
	if os.Getenv(EnvDisable) == "off" {
		st.Error = EnvDisable + "=off"
		return st
	}
	if !st.Enabled {
		if st.DSNMasked == "" {
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
		{"favicon_caches", &FaviconCache{}},
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
