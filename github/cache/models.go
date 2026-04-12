package cache

import "time"

// HTTPCacheEntry stores a cached REST response keyed by (URL, Method).
// On a subsequent request the ETag is sent as If-None-Match; a 304 response
// from GitHub is served from this cache and does not count against the
// primary REST rate limit.
type HTTPCacheEntry struct {
	URL          string `gorm:"primaryKey;size:1024"`
	Method       string `gorm:"primaryKey;size:8"`
	ETag         string `gorm:"size:256"`
	LastModified string `gorm:"size:64"`
	StatusCode   int
	Body         []byte
	// Headers stores a JSON-encoded map[string]string of response headers we
	// care about (Content-Type, Link, etc.). Stored as bytes so it round-trips
	// cleanly across postgres/sqlite without driver-specific JSON columns.
	Headers   []byte
	FetchedAt time.Time `gorm:"index"`
	// ExpiresAt is optional. When nil the entry never expires on its own —
	// it is only revalidated via ETag. When set, the cache treats it as a
	// hard upper bound and will issue a fresh request after expiry even if
	// the ETag still matches.
	ExpiresAt *time.Time
}

// WorkflowRunCache stores completed GitHub Actions workflow runs. Once a run
// is completed its job/step structure is immutable, so we never need to
// re-fetch it from the API. In-progress runs are NOT stored here — those go
// through the ETag-aware HTTP cache only.
type WorkflowRunCache struct {
	RunID      int64  `gorm:"primaryKey"`
	Repo       string `gorm:"index;size:255"`
	Status     string `gorm:"size:32"`
	Conclusion string `gorm:"size:32"`
	// Payload is the JSON-encoded *github.WorkflowRun including its jobs and
	// steps. Stored as JSON so the github package can stay free of
	// cache-specific types.
	Payload   []byte
	FetchedAt time.Time
}

// JobLogCache stores raw GitHub Actions job logs once the job has reached a
// terminal conclusion. Logs can be 10–50MB so we store them gzipped; callers
// see plain strings.
type JobLogCache struct {
	JobID     int64  `gorm:"primaryKey"`
	Repo      string `gorm:"size:255"`
	LogsGz    []byte // gzip-compressed
	FetchedAt time.Time
}

// WorkflowDefCache stores the YAML definition of a workflow at a specific
// commit SHA. The definition at a given SHA is immutable forever.
type WorkflowDefCache struct {
	Repo       string `gorm:"primaryKey;size:255"`
	WorkflowID int64  `gorm:"primaryKey"`
	SHA        string `gorm:"primaryKey;size:64"`
	Path       string `gorm:"size:512"`
	YAML       string
	FetchedAt  time.Time
}
