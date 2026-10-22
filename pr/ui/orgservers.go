// Serves one dashboard Server per organization in authenticated mode. The request's
// tenant.Scope selects the Server; idle organizations are evicted and their Servers stopped.
package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// orgServersEvictInterval is how often OrgServers looks for idle organizations.
const orgServersEvictInterval = time.Minute

// OrgServerFactory builds an organization's Server. Everything the Server starts must
// stop when ctx ends, which is when the organization is evicted.
type OrgServerFactory func(ctx context.Context, scope tenant.Scope) (*Server, error)

// OrgServers keeps one Server per organization. Server itself stays single-org.
type OrgServers struct {
	factory OrgServerFactory
	idleTTL time.Duration
	now     func() time.Time
	stop    context.CancelFunc

	mu      sync.Mutex
	servers map[uuid.UUID]*orgEntry
}

type orgEntry struct {
	handler  http.Handler
	cancel   context.CancelFunc
	lastUsed time.Time
	open     int
}

// NewOrgServers builds Servers with factory and evicts those idle for idleTTL.
func NewOrgServers(factory OrgServerFactory, idleTTL time.Duration) *OrgServers {
	ctx, stop := context.WithCancel(context.Background())
	o := &OrgServers{factory: factory, idleTTL: idleTTL, now: time.Now, stop: stop, servers: map[uuid.UUID]*orgEntry{}}
	go o.evictLoop(ctx)
	return o
}

// Handler serves each request with its organization's Server.
func (o *OrgServers) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scope := tenant.FromContext(r.Context())
		if scope.IsLocal() {
			// The auth middleware always sets a scope; without one the request
			// would be served the host's own data.
			writeJSONError(w, http.StatusInternalServerError, errors.New("the request has no organization"))
			return
		}
		entry, err := o.acquire(scope)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Errorf("open the organization's dashboard: %w", err))
			return
		}
		defer o.release(entry)
		entry.handler.ServeHTTP(w, r)
	})
}

// acquire returns scope's entry, building its Server on first use, and counts the request open.
func (o *OrgServers) acquire(scope tenant.Scope) (*orgEntry, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	entry, ok := o.servers[scope.OrgID]
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		srv, err := o.factory(ctx, scope)
		if err != nil {
			cancel()
			return nil, err
		}
		entry = &orgEntry{handler: srv.Handler(), cancel: cancel}
		o.servers[scope.OrgID] = entry
	}
	entry.open++
	entry.lastUsed = o.now()
	return entry, nil
}

func (o *OrgServers) release(entry *orgEntry) {
	o.mu.Lock()
	defer o.mu.Unlock()
	entry.open--
	entry.lastUsed = o.now()
}

// evictIdle stops the Servers of organizations with no open request that have been
// idle longer than idleTTL.
func (o *OrgServers) evictIdle() {
	o.mu.Lock()
	defer o.mu.Unlock()
	now := o.now()
	for org, entry := range o.servers {
		if entry.open == 0 && now.Sub(entry.lastUsed) > o.idleTTL {
			entry.cancel()
			delete(o.servers, org)
		}
	}
}

func (o *OrgServers) evictLoop(ctx context.Context) {
	t := time.NewTicker(orgServersEvictInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			o.evictIdle()
		}
	}
}

// Close stops every organization's Server and the eviction loop.
func (o *OrgServers) Close() {
	o.stop()
	o.mu.Lock()
	defer o.mu.Unlock()
	for org, entry := range o.servers {
		entry.cancel()
		delete(o.servers, org)
	}
}
