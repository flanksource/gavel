package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const viteUpstreamMarker = "VITE_DEV_UPSTREAM"

func newDevHandler(t *testing.T) http.Handler {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(viteUpstreamMarker))
	}))
	t.Cleanup(upstream.Close)

	s := &Server{}
	require.NoError(t, s.SetDevProxy(upstream.URL))
	return s.Handler()
}

func getBody(t *testing.T, h http.Handler, target string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec.Code, rec.Body.String()
}

func TestDevProxyRouting(t *testing.T) {
	h := newDevHandler(t)

	t.Run("root is proxied to vite", func(t *testing.T) {
		_, body := getBody(t, h, "/")
		assert.Equal(t, viteUpstreamMarker, body)
	})

	t.Run("client route is proxied to vite", func(t *testing.T) {
		_, body := getBody(t, h, "/prs")
		assert.Equal(t, viteUpstreamMarker, body)
	})

	t.Run("unknown asset path falls through to vite", func(t *testing.T) {
		_, body := getBody(t, h, "/@vite/client")
		assert.Equal(t, viteUpstreamMarker, body)
	})

	t.Run("favicon stays on the Go server", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))
		assert.Equal(t, "image/svg+xml", rec.Header().Get("Content-Type"))
		assert.NotContains(t, rec.Body.String(), viteUpstreamMarker)
	})

	t.Run("export request bypasses the proxy", func(t *testing.T) {
		// /prs?format=json is an export route — it must render server-side via
		// handleExport, not get forwarded to Vite.
		_, body := getBody(t, h, "/prs?format=json")
		assert.NotContains(t, body, viteUpstreamMarker)
	})
}

func TestSetDevProxyRejectsRelativeURL(t *testing.T) {
	s := &Server{}
	err := s.SetDevProxy("localhost:5173")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "absolute")
	assert.Nil(t, s.devProxy)
}
