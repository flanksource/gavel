// Package activity records HTTP requests made by the github package so the
// PR UI can surface live timing, size, and cache-hit-rate stats.
//
// The recorder is a process-wide singleton backed by a fixed-size ring buffer
// (capacity ringCapacity). It tracks aggregate counters incrementally so the
// /api/activity endpoint can return both recent entries and totals without
// rescanning.
package activity

import (
	"maps"
	"net/url"
	"sync"
	"time"
)

// ringCapacity bounds the number of recent entries kept in memory.
const ringCapacity = 500

// Kind labels the GitHub call type so the UI can break down stats per source.
const (
	KindREST    = "rest"
	KindGraphQL = "graphql"
	KindSearch  = "search"
	KindFavicon = "favicon"
)

// Entry is one recorded HTTP request.
type Entry struct {
	Timestamp  time.Time     `json:"timestamp"`
	Method     string        `json:"method"`
	URL        string        `json:"url"`
	Kind       string        `json:"kind"`
	StatusCode int           `json:"statusCode"`
	Duration   time.Duration `json:"durationNs"`
	SizeBytes  int           `json:"sizeBytes"`
	FromCache  bool          `json:"fromCache"`
	Error      string        `json:"error,omitempty"`
}

// KindStats are per-Kind aggregate counters.
type KindStats struct {
	Total      int64 `json:"total"`
	CacheHits  int64 `json:"cacheHits"`
	Errors     int64 `json:"errors"`
	TotalBytes int64 `json:"totalBytes"`
	TotalNs    int64 `json:"totalNs"`
}

// Stats are global aggregate counters across all kinds.
type Stats struct {
	Total      int64                `json:"total"`
	CacheHits  int64                `json:"cacheHits"`
	Errors     int64                `json:"errors"`
	TotalBytes int64                `json:"totalBytes"`
	TotalNs    int64                `json:"totalNs"`
	ByKind     map[string]KindStats `json:"byKind"`
}

// Recorder is a bounded in-memory log of HTTP requests with running totals.
// The zero value is not usable; construct via New() or use Shared().
type Recorder struct {
	mu    sync.RWMutex
	ring  [ringCapacity]Entry
	head  int  // next slot to write
	full  bool // ring has wrapped at least once
	stats Stats
}

// New returns a fresh Recorder.
func New() *Recorder {
	return &Recorder{stats: Stats{ByKind: map[string]KindStats{}}}
}

var (
	sharedOnce sync.Once
	shared     *Recorder
)

// Shared returns the process-wide Recorder singleton.
func Shared() *Recorder {
	sharedOnce.Do(func() { shared = New() })
	return shared
}

// Record appends an entry to the ring buffer and updates aggregate counters.
// The entry's URL is scrubbed of any sensitive query parameters before storage.
func (r *Recorder) Record(e Entry) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	e.URL = scrubURL(e.URL)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.ring[r.head] = e
	r.head++
	if r.head >= ringCapacity {
		r.head = 0
		r.full = true
	}

	r.stats.Total++
	r.stats.TotalBytes += int64(e.SizeBytes)
	r.stats.TotalNs += int64(e.Duration)
	if e.FromCache {
		r.stats.CacheHits++
	}
	if e.Error != "" || e.StatusCode >= 400 {
		r.stats.Errors++
	}

	if r.stats.ByKind == nil {
		r.stats.ByKind = map[string]KindStats{}
	}
	ks := r.stats.ByKind[e.Kind]
	ks.Total++
	ks.TotalBytes += int64(e.SizeBytes)
	ks.TotalNs += int64(e.Duration)
	if e.FromCache {
		ks.CacheHits++
	}
	if e.Error != "" || e.StatusCode >= 400 {
		ks.Errors++
	}
	r.stats.ByKind[e.Kind] = ks
}

// Snapshot returns up to limit most-recent entries (newest first) and a copy
// of the global stats. limit <= 0 returns all available entries.
func (r *Recorder) Snapshot(limit int) ([]Entry, Stats) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	available := r.head
	if r.full {
		available = ringCapacity
	}
	if limit <= 0 || limit > available {
		limit = available
	}

	out := make([]Entry, 0, limit)
	// Walk backwards from the most recently written slot.
	idx := r.head - 1
	for i := 0; i < limit; i++ {
		if idx < 0 {
			idx = ringCapacity - 1
		}
		out = append(out, r.ring[idx])
		idx--
	}

	statsCopy := r.stats
	statsCopy.ByKind = make(map[string]KindStats, len(r.stats.ByKind))
	maps.Copy(statsCopy.ByKind, r.stats.ByKind)
	return out, statsCopy
}

// Reset clears the ring buffer and zeroes all counters.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ring = [ringCapacity]Entry{}
	r.head = 0
	r.full = false
	r.stats = Stats{ByKind: map[string]KindStats{}}
}

// scrubURL removes Authorization-like query params (e.g. ?access_token=...)
// so the recorder never holds credentials. Path is preserved as-is.
func scrubURL(raw string) string {
	if raw == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	for _, k := range []string{"access_token", "token", "client_secret"} {
		if q.Has(k) {
			q.Set(k, "REDACTED")
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
