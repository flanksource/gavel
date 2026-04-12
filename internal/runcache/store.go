package runcache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry is the value persisted under a package fingerprint. Only successful
// runs (ExitCode == 0) are written; failures always re-execute.
type Entry struct {
	// Version is the on-disk schema version. Bump to invalidate existing
	// entries on format changes.
	Version int `json:"version"`

	// ImportPath identifies the package for debugging / `gavel cache inspect`.
	ImportPath string `json:"import_path"`

	// Framework is "go test", "ginkgo", etc.
	Framework string `json:"framework"`

	// ExitCode is always 0 for entries we persist.
	ExitCode int `json:"exit_code"`

	// PassCount / FailCount / SkipCount are summary counts from the original
	// run. FailCount is always 0 for persisted entries.
	PassCount int `json:"pass_count"`
	FailCount int `json:"fail_count"`
	SkipCount int `json:"skip_count"`

	// DurationNanos is the wall-clock duration of the original run.
	DurationNanos int64 `json:"duration_nanos"`

	// RecordedAt is the unix-nano timestamp when this entry was written.
	RecordedAt int64 `json:"recorded_at"`
}

// Duration returns the recorded run duration as time.Duration.
func (e Entry) Duration() time.Duration { return time.Duration(e.DurationNanos) }

// schemaVersion is the current on-disk format. Bumping invalidates all
// prior entries (Load returns miss on mismatch).
const schemaVersion = 1

// Store is a content-addressed cache of per-package run outcomes. It is safe
// for concurrent use by multiple goroutines; filesystem atomicity is provided
// by rename().
type Store struct {
	Dir string
}

// DefaultDir returns the default cache directory, ~/.cache/gavel/runs.
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}
	return filepath.Join(home, ".cache", "gavel", "runs"), nil
}

// Open returns a Store rooted at dir, creating the directory if needed.
// If dir is empty, DefaultDir is used.
func Open(dir string) (*Store, error) {
	if dir == "" {
		d, err := DefaultDir()
		if err != nil {
			return nil, err
		}
		dir = d
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return &Store{Dir: dir}, nil
}

// Lookup returns the entry for a given fingerprint, or (zero, false) on miss.
// Corrupted files, schema mismatches, and IO errors all surface as a miss
// rather than an error — the cache must never be load-bearing.
func (s *Store) Lookup(fingerprint string) (Entry, bool) {
	path := s.entryPath(fingerprint)
	data, err := os.ReadFile(path)
	if err != nil {
		return Entry{}, false
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return Entry{}, false
	}
	if e.Version != schemaVersion {
		return Entry{}, false
	}
	return e, true
}

// Record persists a successful run under the given fingerprint. Only
// entries with ExitCode == 0 and FailCount == 0 are written; all other
// inputs are silently ignored. If tooRecent is true (see Fingerprint), the
// write is skipped to defend against mtime granularity races.
func (s *Store) Record(fingerprint string, tooRecent bool, entry Entry) error {
	if entry.ExitCode != 0 || entry.FailCount > 0 {
		return nil
	}
	if tooRecent {
		return nil
	}
	entry.Version = schemaVersion
	if entry.RecordedAt == 0 {
		entry.RecordedAt = time.Now().UnixNano()
	}

	path := s.entryPath(fingerprint)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir shard: %w", err)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}

	// Atomic write: temp file in the same directory, then rename.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".gavel-cache-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op on successful rename

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

// entryPath returns the file path for a given fingerprint. We shard by the
// first two hex characters to avoid blowing up a single directory — same
// pattern Go's build cache uses (cmd/go/internal/cache/cache.go:117).
func (s *Store) entryPath(fingerprint string) string {
	shard := "00"
	if len(fingerprint) >= 2 {
		shard = fingerprint[:2]
	}
	return filepath.Join(s.Dir, shard, fingerprint+".json")
}
