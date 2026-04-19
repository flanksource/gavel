package cache

import (
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasValidator(t *testing.T) {
	tests := []struct {
		name string
		r    *CachedHTTPResponse
		want bool
	}{
		{"nil", nil, false},
		{"empty", &CachedHTTPResponse{}, false},
		{"etag only", &CachedHTTPResponse{ETag: `"abc"`}, true},
		{"last-modified only", &CachedHTTPResponse{LastModified: "Wed, 21 Oct 2026 07:28:00 GMT"}, true},
		{"both", &CachedHTTPResponse{ETag: `"abc"`, LastModified: "Wed, 21 Oct 2026 07:28:00 GMT"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.r.HasValidator())
		})
	}
}

func TestEncodeDecodeHeaders(t *testing.T) {
	in := nethttp.Header{}
	in.Set("Content-Type", "application/json")
	in.Set("ETag", `"v1"`)
	in.Set("Set-Cookie", "session=secret") // not in relevantHeaderKeys, must be dropped
	in.Set("Last-Modified", "Wed, 21 Oct 2026 07:28:00 GMT")

	encoded := encodeHeaders(in)
	require.NotNil(t, encoded)

	decoded := decodeHeaders(encoded)
	require.NotNil(t, decoded)

	// Whitelisted headers round-trip.
	assert.Equal(t, "application/json", decoded.Get("Content-Type"))
	assert.Equal(t, `"v1"`, decoded.Get("ETag"))
	assert.Equal(t, "Wed, 21 Oct 2026 07:28:00 GMT", decoded.Get("Last-Modified"))

	// Non-whitelisted header was stripped.
	assert.Empty(t, decoded.Get("Set-Cookie"), "Set-Cookie should not be persisted")
}

func TestEncodeHeadersEmpty(t *testing.T) {
	assert.Nil(t, encodeHeaders(nil))
	assert.Nil(t, encodeHeaders(nethttp.Header{}))
	// Headers present but none in the whitelist.
	h := nethttp.Header{}
	h.Set("X-Custom", "value")
	assert.Nil(t, encodeHeaders(h))
}

func TestDecodeHeadersEmpty(t *testing.T) {
	assert.Nil(t, decodeHeaders(nil))
	assert.Nil(t, decodeHeaders([]byte{}))
	assert.Nil(t, decodeHeaders([]byte("not json")))
}

func TestSharedReturnsDisabledWhenNoDSN(t *testing.T) {
	resetSharedStore(t)
	t.Setenv(EnvDSN, "")
	t.Setenv(EnvDisable, "")

	s := Shared()
	require.NotNil(t, s)
	assert.True(t, s.Disabled(), "no DSN means store is disabled")

	// Disabled stores must answer cache lookups with nil and ignore writes
	// — this is the contract that lets callers use Shared() unconditionally.
	assert.Nil(t, s.LookupHTTP("https://api.github.com/foo", "GET"))
	s.StoreHTTP("https://api.github.com/foo", "GET", 200, []byte("body"), nil, nil)
	s.TouchHTTP("https://api.github.com/foo", "GET")
	assert.Nil(t, s.LookupHTTP("https://api.github.com/foo", "GET"))
}

func TestSharedRespectsDisableEnv(t *testing.T) {
	resetSharedStore(t)
	t.Setenv(EnvDSN, "postgres://invalid")
	t.Setenv(EnvDisable, "off")

	s := Shared()
	require.NotNil(t, s)
	assert.True(t, s.Disabled(), "GAVEL_GITHUB_CACHE=off must override DSN")
}

// resetSharedStore clears the package-level singleton so each test gets a
// fresh Open() pass, and points $HOME at a temp dir so Open's fallback to
// ~/.config/gavel/db.json can't pick up a real install on the developer's
// machine (which would launch embedded postgres and flip the "disabled"
// assertions). sync.Once has no Reset method so we replace the value.
func resetSharedStore(t *testing.T) {
	t.Helper()
	sharedStore = nil
	sharedStoreOnce = sync.Once{}
	t.Setenv("HOME", t.TempDir())
}

// Integration tests below require GAVEL_GITHUB_CACHE_DSN to point at a real
// postgres database. They are skipped automatically when not set so the
// suite stays green in environments without postgres.

func TestHTTPCacheRoundtripIntegration(t *testing.T) {
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run integration tests", EnvDSN)
	}
	t.Setenv(EnvDSN, dsn)
	t.Setenv(EnvDisable, "")

	store, err := Open()
	require.NoError(t, err)
	require.False(t, store.Disabled())
	t.Cleanup(func() { _ = store.Close() })
	t.Cleanup(func() { _ = store.gorm().Where("1=1").Delete(&HTTPCacheEntry{}) })

	// Simulate a server that returns ETag on first hit, 304 on subsequent
	// hits when If-None-Match matches.
	const etag = `"v1"`
	hits := 0
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		hits++
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(nethttp.StatusNotModified)
			return
		}
		w.Header().Set("ETag", etag)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(nethttp.StatusOK)
		_, _ = w.Write([]byte(`{"value":42}`))
	}))
	defer srv.Close()

	url := srv.URL + "/repos/owner/name/actions/runs/1"

	// Cold call.
	require.Nil(t, store.LookupHTTP(url, "GET"))
	resp, err := nethttp.Get(url)
	require.NoError(t, err)
	body := []byte(`{"value":42}`)
	store.StoreHTTP(url, "GET", resp.StatusCode, body, resp.Header, nil)
	resp.Body.Close()

	// Warm lookup returns the body and the ETag.
	cached := store.LookupHTTP(url, "GET")
	require.NotNil(t, cached)
	assert.Equal(t, etag, cached.ETag)
	assert.Equal(t, body, cached.Body)
	assert.True(t, cached.HasValidator())

	// A subsequent conditional request hits 304 and we touch the entry.
	req, _ := nethttp.NewRequest("GET", url, nil)
	req.Header.Set("If-None-Match", cached.ETag)
	resp, err = nethttp.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, nethttp.StatusNotModified, resp.StatusCode, "server should return 304 on matching ETag")
	store.TouchHTTP(url, "GET")

	// FetchedAt was bumped.
	cachedAfter := store.LookupHTTP(url, "GET")
	require.NotNil(t, cachedAfter)
	assert.True(t, !cachedAfter.FetchedAt.Before(cached.FetchedAt))

	// We made exactly 2 server hits — the second was a 304.
	assert.Equal(t, 2, hits)
}

func TestHTTPCacheExpiresAtIntegration(t *testing.T) {
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run integration tests", EnvDSN)
	}
	t.Setenv(EnvDSN, dsn)
	t.Setenv(EnvDisable, "")

	store, err := Open()
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	t.Cleanup(func() { _ = store.gorm().Where("1=1").Delete(&HTTPCacheEntry{}) })

	url := "https://example.test/foo"
	past := time.Now().Add(-time.Hour)
	headers := nethttp.Header{}
	headers.Set("ETag", `"x"`)
	store.StoreHTTP(url, "GET", 200, []byte("payload"), headers, &past)

	// Already-expired entry yields nil.
	assert.Nil(t, store.LookupHTTP(url, "GET"))
}
