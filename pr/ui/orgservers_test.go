// Tests OrgServers: one dashboard Server per organization, idle ones evicted and
// cancelled unless a request is still open, and no Server for a request without an org.
package ui

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orgServersFixture struct {
	servers *OrgServers
	now     time.Time
	mu      sync.Mutex
	built   map[uuid.UUID][]context.Context
	fail    error
}

func newOrgServersFixture(t *testing.T) *orgServersFixture {
	f := &orgServersFixture{now: time.Now(), built: map[uuid.UUID][]context.Context{}}
	f.servers = NewOrgServers(func(ctx context.Context, scope tenant.Scope) (*Server, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.fail != nil {
			return nil, f.fail
		}
		f.built[scope.OrgID] = append(f.built[scope.OrgID], ctx)
		noGitHub := github.Options{TokenSource: func() (string, error) { return "", errors.New("no GitHub in this test") }}
		srv := NewServer(gavelctx.New(ctx), time.Minute, noGitHub, SearchConfig{})
		srv.SetOrg(scope.OrgID, nil)
		srv.SetHostedMode(true)
		return srv, nil
	}, 30*time.Minute)
	f.servers.now = func() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.now }
	t.Cleanup(f.servers.Close)
	return f
}

func (f *orgServersFixture) advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

func (f *orgServersFixture) builds(org uuid.UUID) []context.Context {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]context.Context(nil), f.built[org]...)
}

func (f *orgServersFixture) get(org uuid.UUID, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	f.servers.Handler().ServeHTTP(rec, req.WithContext(tenant.WithScope(req.Context(), tenant.Scope{OrgID: org})))
	return rec
}

func TestOrgServersBuildsOneServerPerOrg(t *testing.T) {
	f := newOrgServersFixture(t)
	a, b := uuid.New(), uuid.New()
	for _, org := range []uuid.UUID{a, b, a} {
		require.Equal(t, http.StatusOK, f.get(org, "/api/config").Code)
	}
	assert.Len(t, f.builds(a), 1)
	assert.Len(t, f.builds(b), 1)
}

func TestOrgServersRefusesARequestWithoutAnOrg(t *testing.T) {
	f := newOrgServersFixture(t)
	rec := httptest.NewRecorder()
	f.servers.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/config", nil))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "the request has no organization")
	assert.Empty(t, f.builds(tenant.LocalOrgID), "the local org's data is never served in hosted mode")
}

func TestOrgServersReportsAFactoryError(t *testing.T) {
	f := newOrgServersFixture(t)
	f.fail = errors.New("org settings unavailable")
	org := uuid.New()
	rec := f.get(org, "/api/config")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "open the organization's dashboard: org settings unavailable")

	f.fail = nil
	assert.Equal(t, http.StatusOK, f.get(org, "/api/config").Code, "a failure isn't cached")
}

func TestOrgServersEvictsIdleServers(t *testing.T) {
	f := newOrgServersFixture(t)
	org := uuid.New()
	f.get(org, "/api/config")
	first := f.builds(org)[0]

	f.advance(29 * time.Minute)
	f.servers.evictIdle()
	require.NoError(t, first.Err(), "not idle long enough")

	f.advance(2 * time.Minute)
	f.servers.evictIdle()
	require.ErrorIs(t, first.Err(), context.Canceled, "an evicted Server's context ends")
	f.get(org, "/api/config")
	assert.Len(t, f.builds(org), 2, "the next request builds a fresh Server")
}

func TestOrgServersKeepsAServerWithAnOpenRequest(t *testing.T) {
	f := newOrgServersFixture(t)
	org := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.servers.Handler().ServeHTTP(w, r.WithContext(tenant.WithScope(r.Context(), tenant.Scope{OrgID: org})))
	}))
	t.Cleanup(server.Close)
	ctx, closeStream := context.WithCancel(context.Background())
	defer closeStream()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/prs/stream", nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.True(t, strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream"))
	built := f.builds(org)[0]

	f.advance(time.Hour)
	f.servers.evictIdle()
	require.NoError(t, built.Err(), "a Server with an open stream stays")

	closeStream()
	require.Eventually(t, func() bool {
		f.advance(31 * time.Minute)
		f.servers.evictIdle()
		return built.Err() != nil
	}, 5*time.Second, 10*time.Millisecond, "evicted once the stream closes and the Server sits idle")
}

func TestOrgServersCloseCancelsEveryServer(t *testing.T) {
	f := newOrgServersFixture(t)
	a, b := uuid.New(), uuid.New()
	f.get(a, "/api/config")
	f.get(b, "/api/config")
	f.servers.Close()
	assert.Error(t, f.builds(a)[0].Err())
	assert.Error(t, f.builds(b)[0].Err())
}
