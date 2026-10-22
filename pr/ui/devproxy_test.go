package ui

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
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

	target, err := url.Parse(upstream.URL)
	require.NoError(t, err)
	s := &Server{}
	s.SetDevProxy(httputil.NewSingleHostReverseProxy(target))
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

	t.Run("unregistered api route is a JSON 404, not vite's html fallback", func(t *testing.T) {
		const missingAPIRoute = "/api/not-a-route"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, missingAPIRoute, nil))

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"error":"no API route for GET `+missingAPIRoute+`"}`, rec.Body.String())
	})

	t.Run("export request bypasses the proxy", func(t *testing.T) {
		// /prs?format=json is an export route — it must render server-side via
		// handleExport, not get forwarded to Vite.
		_, body := getBody(t, h, "/prs?format=json")
		assert.NotContains(t, body, viteUpstreamMarker)
	})
}

func TestDevProxyServesEmbeddedProductionAssets(t *testing.T) {
	h := newDevHandler(t)

	t.Run("stable entry bypasses vite", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/_assets/prui.js", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "javascript")
		assert.NotContains(t, rec.Body.String(), viteUpstreamMarker)
	})

	t.Run("missing chunk does not fall back to html", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/_assets/chunks/missing.js", nil))

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.NotContains(t, rec.Header().Get("Content-Type"), "text/html")
		assert.NotContains(t, rec.Body.String(), viteUpstreamMarker)
	})
}
