package github

import (
	nethttp "net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeGitHub wires up a minimal httptest server that speaks just the two
// endpoints probeToken calls. Handler funcs can be overridden per test to
// simulate different auth states.
type fakeGitHub struct {
	server      *httptest.Server
	rateLimitFn nethttp.HandlerFunc
	userFn      nethttp.HandlerFunc
}

func newFakeGitHub(t *testing.T) *fakeGitHub {
	t.Helper()
	f := &fakeGitHub{}
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/rate_limit", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if f.rateLimitFn != nil {
			f.rateLimitFn(w, r)
			return
		}
		nethttp.Error(w, "not configured", nethttp.StatusInternalServerError)
	})
	mux.HandleFunc("/user", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if f.userFn != nil {
			f.userFn(w, r)
			return
		}
		nethttp.Error(w, "not configured", nethttp.StatusNotFound)
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func TestProbeToken_NoToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("HOME", t.TempDir()) // isolate auth.json

	f := newFakeGitHub(t)
	r := probeToken(Options{}, f.server.URL)
	assert.Equal(t, AuthStateNoToken, r.State)
}

func TestProbeToken_OK(t *testing.T) {
	f := newFakeGitHub(t)
	f.rateLimitFn = func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			nethttp.Error(w, "unauth", nethttp.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"resources":{"core":{"limit":5000,"remaining":4999,"used":1,"reset":1234567890}}}`))
	}
	f.userFn = func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`{"login":"octocat"}`))
	}

	r := probeToken(Options{Token: "valid-token"}, f.server.URL)
	assert.Equal(t, AuthStateOK, r.State)
	assert.Equal(t, "octocat", r.Login)
	assert.Contains(t, r.Message, "octocat")
	if assert.NotNil(t, r.RateLimit) {
		assert.Equal(t, 5000, r.RateLimit.Limit)
		assert.Equal(t, 4999, r.RateLimit.Remaining)
	}
}

func TestProbeToken_Invalid401(t *testing.T) {
	f := newFakeGitHub(t)
	f.rateLimitFn = func(w nethttp.ResponseWriter, r *nethttp.Request) {
		nethttp.Error(w, "Bad credentials", nethttp.StatusUnauthorized)
	}
	r := probeToken(Options{Token: "expired"}, f.server.URL)
	assert.Equal(t, AuthStateInvalid, r.State)
	assert.Contains(t, r.Message, "401")
}

func TestProbeToken_Forbidden403_RateLimit(t *testing.T) {
	f := newFakeGitHub(t)
	f.rateLimitFn = func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Limit", "5000")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
		nethttp.Error(w, "rate limit", nethttp.StatusForbidden)
	}
	r := probeToken(Options{Token: "tok"}, f.server.URL)
	assert.Equal(t, AuthStateRateLimited, r.State)
}

func TestProbeToken_Forbidden403_SSOOrScopes(t *testing.T) {
	f := newFakeGitHub(t)
	f.rateLimitFn = func(w nethttp.ResponseWriter, r *nethttp.Request) {
		nethttp.Error(w, "sso required", nethttp.StatusForbidden)
	}
	r := probeToken(Options{Token: "tok"}, f.server.URL)
	assert.Equal(t, AuthStateInvalid, r.State)
	assert.Contains(t, r.Message, "403")
}

func TestProbeToken_OK_RateLimited(t *testing.T) {
	f := newFakeGitHub(t)
	f.rateLimitFn = func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`{"resources":{"core":{"limit":5000,"remaining":0,"used":5000,"reset":1234567890}}}`))
	}
	r := probeToken(Options{Token: "tok"}, f.server.URL)
	assert.Equal(t, AuthStateRateLimited, r.State)
}

func TestProbeToken_Unreachable(t *testing.T) {
	// Point at a never-listening URL so the dial fails fast. Port 1 is
	// reserved and almost always refused.
	r := probeToken(Options{Token: "tok"}, "http://127.0.0.1:1")
	assert.Equal(t, AuthStateUnreachable, r.State)
}
