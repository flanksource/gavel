// Tests `gavel credential-provider`'s startup checks and its health endpoints: it refuses to start
// without its gavels and certificates, and is ready only once it serves.
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialProviderRefusesToStartWithoutItsSettings(t *testing.T) {
	gavels := filepath.Join(t.TempDir(), "gavels.yaml")
	require.NoError(t, os.WriteFile(gavels, []byte("beta:\n  url: https://gavel-beta.example.com\n  tokenFile: /etc/token\n"), 0o600))
	for name, test := range map[string]struct {
		opts credentialProviderOptions
		want string
	}{
		"no gavels":      {credentialProviderOptions{}, "--gavels is required: the file naming each gavel's url and tokenFile"},
		"missing gavels": {credentialProviderOptions{Gavels: filepath.Join(t.TempDir(), "none.yaml")}, "read the gavels file"},
		"no certificate": {credentialProviderOptions{Gavels: gavels}, "--server-cred-bundle is required"},
		"timeout too long": {credentialProviderOptions{Gavels: gavels, FetchTimeout: 5 * time.Second},
			"--fetch-timeout must be under 5s, the egress gateway's limit, got 5s"},
	} {
		err := runCredentialProvider(t.Context(), test.opts, nil)
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), test.want, name)
	}
}

func TestCredentialProviderHealth(t *testing.T) {
	var ready atomic.Bool
	handler := credentialProviderHealth(&ready)
	get := func(path string) int {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Code
	}
	assert.Equal(t, http.StatusOK, get("/healthz"))
	assert.Equal(t, http.StatusServiceUnavailable, get("/readyz"), "not ready before it serves")
	ready.Store(true)
	assert.Equal(t, http.StatusOK, get("/readyz"))
}
