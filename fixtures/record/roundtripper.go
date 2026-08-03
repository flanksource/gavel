package record

import (
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func init() { Implemented[KindClients] = true }

// maxClientEntries bounds what one recording keeps in memory, mirroring the
// statement cap. Reaching it marks the artifact truncated rather than growing
// without limit.
const maxClientEntries = 20000

// clients is where wrapped transports send what they see. Like the gorm sink it
// is a process-global: gavel's HTTP clients are built all over the process, long
// before any fixture declares `record: clients`, and threading a sink to each of
// them would mean rebuilding every client per fixture.
//
// Nil until a fixture asks — one atomic load per request, and the transport
// behaves exactly as it did before.
var clients atomic.Pointer[ClientLog]

// ClientLog collects the exchanges gavel's own HTTP clients made. It is the
// counterpart to the HTTP proxy: the proxy watches a child process, this watches
// gavel. Neither sees the other's traffic.
type ClientLog struct {
	policy capturePolicy

	mu        sync.Mutex
	entries   []Entry
	truncated bool
}

// StartClients begins recording gavel's own HTTP calls. Only one recording is
// active at a time — the sink is global — so a second call replaces the first.
func StartClients(opts ClientOptions) *ClientLog {
	log := &ClientLog{policy: capturePolicy{Bodies: opts.Bodies, Redact: opts.Redact}}
	clients.Store(log)
	return log
}

// StopClients detaches the sink. The log keeps what it collected.
func StopClients() { clients.Store(nil) }

func (l *ClientLog) add(entry Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) >= maxClientEntries {
		l.truncated = true
		return
	}
	l.entries = append(l.entries, entry)
}

// Entries returns everything recorded so far, oldest first. gavel makes requests
// concurrently, so the sort is what makes the artifact reproducible.
func (l *ClientLog) Entries() []Entry {
	l.mu.Lock()
	entries := append([]Entry(nil), l.entries...)
	l.mu.Unlock()

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].started.Before(entries[j].started) })
	return entries
}

// Between narrows to one fixture's window, with the same time-slice caveat the
// proxy documents.
func (l *ClientLog) Between(from, to time.Time) []Entry {
	var window []Entry
	for _, entry := range l.Entries() {
		if entry.started.Before(from) || entry.started.After(to) {
			continue
		}
		window = append(window, entry)
	}
	return window
}

// Truncated reports whether the cap was reached.
func (l *ClientLog) Truncated() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.truncated
}

// Save writes the log as a HAR artifact.
func (l *ClientLog) Save(store *Store, label string, entries []Entry) (Result, error) {
	result, err := SaveHAR(store, label, KindClients, entries)
	result.Truncated = l.Truncated()
	return result, err
}

// MaybeWrap returns a transport that records into whatever recording is active
// when a request is made. It is always safe to call and always cheap, so gavel's
// clients can be wired once at construction: whether anything is recorded is
// decided per request, not per client.
func MaybeWrap(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if _, already := base.(*harTransport); already {
		return base
	}
	return &harTransport{base: base}
}

type harTransport struct{ base http.RoundTripper }

func (t *harTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	log := clients.Load()
	if log == nil {
		return t.base.RoundTrip(req)
	}

	state := exchangeState{started: time.Now()}
	if log.policy.Bodies > 0 && req.Body != nil {
		state.post = capturePostData(req, log.policy.Bodies.Or(defaultBodies))
	}

	resp, err := t.base.RoundTrip(req)
	// Recorded even on error: a request that never got an answer is the one a
	// failing fixture most needs to see.
	log.add(exchangeEntry(log.policy, req, resp, state, err))
	return resp, err
}
