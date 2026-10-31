// Tests FetchSecret against a gavel stand-in: URIs are parsed strictly, each gavel is asked with its
// own token, answers are cached per run and actor, and gavel's refusals map to gRPC codes.
package credentialprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flanksource/gavel/internal/logtest"
	"github.com/flanksource/gavel/sandbox/internal/substrate/credproviderpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	testRun   = "6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f"
	testActor = "spiffe://substrate-actor.local/actor/ate-env/gavel-run-" + testRun
	testURI   = "ate-secret://gavel/beta/" + testRun + "/anthropic"
)

// gavelStandIn answers /api/credentials/fetch with status and body, recording each request.
type gavelStandIn struct {
	server   *httptest.Server
	calls    atomic.Int32
	status   int
	value    string
	delay    time.Duration
	token    string
	requests chan map[string]string
}

func newGavel(t *testing.T) *gavelStandIn {
	t.Helper()
	g := &gavelStandIn{status: http.StatusOK, value: "sk-ant-oat01-provider-test", requests: make(chan map[string]string, 16)}
	g.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.calls.Add(1)
		time.Sleep(g.delay)
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		body["authorization"] = r.Header.Get("Authorization")
		body["path"] = r.URL.Path
		g.requests <- body
		w.WriteHeader(g.status)
		if g.status == http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]string{"value": g.value})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "refused by gavel"})
	}))
	t.Cleanup(g.server.Close)
	return g
}

func tokenFile(t *testing.T, token string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(path, []byte(token+"\n"), 0o600))
	return path
}

func newServer(t *testing.T, g *gavelStandIn) *Server {
	t.Helper()
	return New(Config{
		Gavels:       map[string]Gavel{"beta": {URL: g.server.URL, TokenFile: tokenFile(t, "beta-token")}},
		FetchTimeout: 2 * time.Second,
		CacheTTL:     time.Minute,
	})
}

func fetch(s *Server, uri, actor string) (*credproviderpb.FetchSecretResponse, error) {
	return s.FetchSecret(context.Background(), &credproviderpb.FetchSecretRequest{Uri: uri, ActorSpiffeId: actor})
}

func TestFetchSecretAsksTheNamedGavelAndCachesTheAnswer(t *testing.T) {
	g := newGavel(t)
	s := newServer(t, g)

	resp, err := fetch(s, testURI, testActor)
	require.NoError(t, err)
	assert.Equal(t, "sk-ant-oat01-provider-test", string(resp.GetOpaqueBytes()))
	assert.Equal(t, map[string]string{
		"run": testRun, "provider": "anthropic", "actor": testActor,
		"authorization": "Bearer beta-token", "path": "/api/credentials/fetch",
	}, <-g.requests)

	_, err = fetch(s, testURI, testActor)
	require.NoError(t, err)
	assert.EqualValues(t, 1, g.calls.Load(), "the second fetch in the TTL is answered from the cache")

	_, err = fetch(s, testURI, "spiffe://substrate-actor.local/actor/ate-env/other")
	require.NoError(t, err)
	assert.EqualValues(t, 2, g.calls.Load(), "another actor isn't answered from this actor's entry")
}

func TestFetchSecretCachesOnlyWithinTheTTL(t *testing.T) {
	g := newGavel(t)
	s := New(Config{
		Gavels:       map[string]Gavel{"beta": {URL: g.server.URL, TokenFile: tokenFile(t, "beta-token")}},
		FetchTimeout: 2 * time.Second, CacheTTL: time.Nanosecond,
	})
	for range 2 {
		_, err := fetch(s, testURI, testActor)
		require.NoError(t, err)
	}
	assert.EqualValues(t, 2, g.calls.Load())
}

func TestFetchSecretReadsTheTokenFileOnEveryCall(t *testing.T) {
	g := newGavel(t)
	path := tokenFile(t, "first-token")
	s := New(Config{Gavels: map[string]Gavel{"beta": {URL: g.server.URL, TokenFile: path}}, FetchTimeout: 2 * time.Second})

	_, err := fetch(s, testURI, testActor)
	require.NoError(t, err)
	assert.Equal(t, "Bearer first-token", (<-g.requests)["authorization"])
	require.NoError(t, os.WriteFile(path, []byte("rotated-token"), 0o600))
	_, err = fetch(s, testURI, testActor)
	require.NoError(t, err)
	assert.Equal(t, "Bearer rotated-token", (<-g.requests)["authorization"], "a rotated token is used without a restart")
}

func TestFetchSecretMapsGavelsAnswersToCodes(t *testing.T) {
	logged := logtest.Capture(t)
	for name, test := range map[string]struct {
		status int
		code   codes.Code
	}{
		"refused":       {http.StatusForbidden, codes.PermissionDenied},
		"no key":        {http.StatusNotFound, codes.NotFound},
		"wrong token":   {http.StatusUnauthorized, codes.PermissionDenied},
		"gavel failing": {http.StatusInternalServerError, codes.Unavailable},
		"gavel down":    {http.StatusBadGateway, codes.Unavailable},
	} {
		g := newGavel(t)
		g.status = test.status
		s := newServer(t, g)
		_, err := fetch(s, testURI, testActor)
		assert.Equal(t, test.code, status.Code(err), name)
		assert.Contains(t, status.Convert(err).Message(), "refused by gavel", name)
		_, _ = fetch(s, testURI, testActor)
		assert.EqualValues(t, 2, g.calls.Load(), "%s: a failure isn't cached", name)
	}
	assert.Contains(t, logged.String(), "credential fetch "+testURI+" for "+testActor+": rpc error: code = PermissionDenied desc = gavel beta refused this provider's token, so its tokenFile is wrong")
}

func TestFetchSecretWhenGavelIsUnreachableOrSlow(t *testing.T) {
	logged := logtest.Capture(t)
	g := newGavel(t)
	s := newServer(t, g)
	g.server.Close()
	_, err := fetch(s, testURI, testActor)
	assert.Equal(t, codes.Unavailable, status.Code(err))

	slow := newGavel(t)
	slow.delay = 200 * time.Millisecond
	s = New(Config{Gavels: map[string]Gavel{"beta": {URL: slow.server.URL, TokenFile: tokenFile(t, "t")}}, FetchTimeout: 50 * time.Millisecond})
	_, err = fetch(s, testURI, testActor)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
	assert.Contains(t, logged.String(), "gavel beta didn't answer within 50ms")
}

func TestFetchSecretRefusesURIsItDoesNotServe(t *testing.T) {
	g := newGavel(t)
	s := newServer(t, g)
	for uri, want := range map[string]string{
		"ate-secret://gavel/prod/" + testRun + "/anthropic":         `no gavel named "prod"`,
		"ate-secret://k8s.io/default/gavel/model/key":               `names provider "k8s.io", not gavel`,
		"ate-secret://gavel/beta/not-a-run/anthropic":               `"not-a-run" is not a run ID`,
		"ate-secret://gavel/beta/" + testRun:                        "must be ate-secret://gavel/<gavel>/<run id>/<provider>",
		"ate-secret://gavel/beta/" + testRun + "/anthropic/extra":   "must be ate-secret://gavel/<gavel>/<run id>/<provider>",
		"ate-secret://gavel/beta/" + testRun + "/anthropic?x=1":     "must be ate-secret://gavel/<gavel>/<run id>/<provider>",
		"ate-secret://gavel/beta/" + testRun + "/Anthropic%20Key":   "must be ate-secret://gavel/<gavel>/<run id>/<provider>",
		"https://gavel/beta/" + testRun + "/anthropic":              "must be ate-secret://gavel/<gavel>/<run id>/<provider>",
		"ate-secret://user@gavel/beta/" + testRun + "/anthropic":    "must be ate-secret://gavel/<gavel>/<run id>/<provider>",
		"ate-secret://gavel/beta/" + testRun + "/anthropic#section": "must be ate-secret://gavel/<gavel>/<run id>/<provider>",
	} {
		_, err := fetch(s, uri, testActor)
		assert.Equal(t, codes.InvalidArgument, status.Code(err), uri)
		assert.Contains(t, status.Convert(err).Message(), want, uri)
	}
	assert.Zero(t, g.calls.Load(), "gavel is never asked about a URI the provider refuses")
}

func TestLoadGavelsReadsEachGavelsURLAndTokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gavels.yaml")
	require.NoError(t, os.WriteFile(path, []byte("beta:\n  url: https://gavel-beta.flanksource.com\n  tokenFile: /etc/gavel/beta-token\n"), 0o600))
	gavels, err := LoadGavels(path)
	require.NoError(t, err)
	assert.Equal(t, map[string]Gavel{"beta": {URL: "https://gavel-beta.flanksource.com", TokenFile: "/etc/gavel/beta-token"}}, gavels)

	for content, want := range map[string]string{
		"":                          "names no gavel",
		"beta:\n  url: https://x\n": `gavel "beta" needs url and tokenFile`,
		"Beta:\n  url: https://x\n  tokenFile: t": `gavel name "Beta" must be a DNS label`,
		"beta:\n  url: ftp://x\n  tokenFile: t":   `gavel "beta" url "ftp://x" must be http or https`,
	} {
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		_, err := LoadGavels(path)
		require.Error(t, err, content)
		assert.Contains(t, err.Error(), want, content)
	}
}
