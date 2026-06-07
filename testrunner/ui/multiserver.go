package testui

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// DefaultMaxRuns caps how many per-run Servers a MultiServer holds at once.
// Beyond it, the oldest idle finished run is evicted on the next BeginRun.
const DefaultMaxRuns = 32

// DefaultRunIdleTTL is how long a finished run's Server is retained after its
// last access before it is reclaimed. Aligned with the clicky task GC window so
// a viewer has time to reconnect; afterwards the run replays from history.
const DefaultRunIdleTTL = 10 * time.Minute

// MultiServer hosts many concurrent test runs, one ordinary *Server per run id.
// It is the multi-run front for Server: each run keeps its own isolated state
// and SSE stream, and Handler() routes /{runID}/... to that run's Server. Server
// itself is unchanged and single-run; the single-run paths (apply suite --ui,
// snapshot replay) keep using a bare *Server directly.
type MultiServer struct {
	mu      sync.Mutex
	runs    map[string]*runEntry
	max     int
	idleTTL time.Duration
	timeNow func() time.Time
}

type runEntry struct {
	server     *Server
	lastAccess time.Time
}

// NewMultiServer returns a MultiServer with default capacity and idle TTL.
func NewMultiServer() *MultiServer {
	return &MultiServer{
		runs:    make(map[string]*runEntry),
		max:     DefaultMaxRuns,
		idleTTL: DefaultRunIdleTTL,
		timeNow: time.Now,
	}
}

// SetLimits overrides the run cap and idle TTL. A non-positive value leaves the
// corresponding default in place.
func (m *MultiServer) SetLimits(max int, idleTTL time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if max > 0 {
		m.max = max
	}
	if idleTTL > 0 {
		m.idleTTL = idleTTL
	}
}

// BeginRun returns the Server for runID, creating it on first use, and starts a
// run on it via Server.BeginRun(kind). Re-running an existing id reuses its
// Server (a fresh BeginRun resets that run's state, matching single-run
// semantics). Reclaims idle finished runs first so a long-lived serve process
// does not accumulate Servers without bound.
func (m *MultiServer) BeginRun(runID, kind string) *Server {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.runs[runID]
	if !ok {
		entry = &runEntry{server: NewServer()}
		m.runs[runID] = entry
	}
	entry.lastAccess = m.timeNow()
	entry.server.BeginRun(kind)
	// Reclaim after inserting so the cap is enforced against the final set and
	// this just-started run is the most-recently-accessed (never evicted here).
	m.reclaimLocked()
	return entry.server
}

// Get returns the Server for runID, if it exists, and bumps its last-access time.
func (m *MultiServer) Get(runID string) (*Server, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.runs[runID]
	if !ok {
		return nil, false
	}
	entry.lastAccess = m.timeNow()
	return entry.server, true
}

// reclaimLocked drops finished runs that have been idle past idleTTL, then, if
// still over capacity, evicts the least-recently-accessed finished runs until at
// or under max. A still-running run is never evicted: if every run is running,
// the registry is allowed to exceed max rather than dropping a live stream.
func (m *MultiServer) reclaimLocked() {
	now := m.timeNow()
	for id, entry := range m.runs {
		if entry.server.Done() && now.Sub(entry.lastAccess) > m.idleTTL {
			delete(m.runs, id)
		}
	}
	for len(m.runs) > m.max {
		oldestID, oldest := "", time.Time{}
		for id, entry := range m.runs {
			if !entry.server.Done() {
				continue
			}
			if oldestID == "" || entry.lastAccess.Before(oldest) {
				oldestID, oldest = id, entry.lastAccess
			}
		}
		if oldestID == "" {
			return // all remaining runs are live; keep them
		}
		delete(m.runs, oldestID)
	}
}

// Handler routes /{runID}/<rest> to that run's Server.Handler() serving
// <rest>. An unknown run id is a 404 — there is no shared fallback stream, so a
// stale client surfaces the miss (and the UI falls back to the history
// snapshot) rather than silently reading another run's state.
func (m *MultiServer) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runID, rest := splitRunID(r.URL.Path)
		if runID == "" {
			http.NotFound(w, r)
			return
		}
		srv, ok := m.Get(runID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		srv.Handler().ServeHTTP(w, withPath(r, rest))
	})
}

// splitRunID peels the first path segment as the run id and returns the
// remaining path (leading-slash preserved, "/" when empty).
func splitRunID(path string) (runID, rest string) {
	trimmed := strings.TrimPrefix(path, "/")
	if trimmed == "" {
		return "", "/"
	}
	if i := strings.IndexByte(trimmed, '/'); i >= 0 {
		return trimmed[:i], trimmed[i:]
	}
	return trimmed, "/"
}

// withPath returns a shallow copy of r whose URL path is set to p. StripPrefix
// keys off URL.Path, so this normalises the remainder before delegation.
func withPath(r *http.Request, p string) *http.Request {
	clone := r.Clone(r.Context())
	clone.URL.Path = "/" + strings.TrimPrefix(p, "/")
	return clone
}
