package ui

import (
	"sort"
	"sync"
	"time"
)

type SyncState string

const (
	SyncQueued    SyncState = "queued"
	SyncSyncing   SyncState = "syncing"
	SyncUpToDate  SyncState = "up-to-date"
	SyncOutOfDate SyncState = "out-of-date"
	SyncError     SyncState = "error"
)

type PRSyncStatus struct {
	State      SyncState `json:"state"`
	LastSynced time.Time `json:"lastSynced,omitempty"`
	Error      string    `json:"error,omitempty"`
	Phase      string    `json:"phase,omitempty"`
}

type detailCacheEntry struct {
	Detail      prDetail
	FetchedAt   time.Time
	PRUpdatedAt time.Time
}

type DetailCache struct {
	mu      sync.RWMutex
	entries map[string]*detailCacheEntry
	status  map[string]*PRSyncStatus
	// started counts fetches begun; settled is, per PR, the newest fetch whose
	// detail was cached or that an invalidation overtook.
	started uint64
	settled map[string]uint64
	maxSize int
}

func NewDetailCache() *DetailCache {
	return &DetailCache{
		entries: make(map[string]*detailCacheEntry),
		status:  make(map[string]*PRSyncStatus),
		settled: make(map[string]uint64),
		maxSize: 100,
	}
}

func (c *DetailCache) Get(key string) (*detailCacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	return e, ok
}

// Begin starts a fetch and returns the ticket its Put must present.
func (c *DetailCache) Begin() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.started++
	return c.started
}

// Put records the outcome of the fetch that ticket began. It reports false, and
// changes nothing, when a fetch begun later was already cached or the PR was
// invalidated meanwhile. A failed fetch keeps the last good detail.
func (c *DetailCache) Put(key string, ticket uint64, detail prDetail) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ticket <= c.settled[key] {
		return false
	}
	status := &PRSyncStatus{State: SyncUpToDate, LastSynced: time.Now()}
	if problem := detail.problem(); problem != "" {
		status.State, status.Error = SyncError, problem
	}
	c.status[key] = status
	if detail.Error != "" {
		return true
	}
	c.settled[key] = ticket
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}
	c.entries[key] = &detailCacheEntry{
		Detail:      detail,
		FetchedAt:   time.Now(),
		PRUpdatedAt: detail.PR.UpdatedAt,
	}
	return true
}

func (c *DetailCache) IsStale(key string, currentUpdatedAt time.Time) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok {
		return true
	}
	return currentUpdatedAt.After(e.PRUpdatedAt)
}

func (c *DetailCache) GetStatus(key string) (PRSyncStatus, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.status[key]
	if !ok {
		return PRSyncStatus{}, false
	}
	return *s, true
}

func (c *DetailCache) SetStatus(key string, status PRSyncStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status[key] = &status
}

func (c *DetailCache) AllStatuses() map[string]PRSyncStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]PRSyncStatus, len(c.status))
	for k, v := range c.status {
		out[k] = *v
	}
	return out
}

// Invalidate drops the cached detail and status for a single PR so the next
// request re-fetches it from GitHub. Used after a mutating action (merge,
// approve, auto-merge) where the cached state is now stale.
func (c *DetailCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.settled[key] = c.started
	delete(c.entries, key)
	delete(c.status, key)
}

// EvictStale removes entries and statuses for PRs not in the given set of keys.
func (c *DetailCache) EvictStale(activeKeys map[string]bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.entries {
		if !activeKeys[k] {
			delete(c.entries, k)
		}
	}
	for k := range c.status {
		if !activeKeys[k] {
			delete(c.status, k)
		}
	}
}

// evictOldest removes the oldest entry by FetchedAt. Must be called with mu held.
func (c *DetailCache) evictOldest() {
	if len(c.entries) == 0 {
		return
	}
	type kv struct {
		key string
		at  time.Time
	}
	items := make([]kv, 0, len(c.entries))
	for k, e := range c.entries {
		items = append(items, kv{k, e.FetchedAt})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].at.Before(items[j].at) })
	// Remove oldest 10% to avoid evicting on every insert
	n := len(items) / 10
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		delete(c.entries, items[i].key)
	}
}
