package record

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// proxyClient returns a client routed through the recorder, the way an
// HTTP_PROXY-respecting child process would be.
func proxyClient(t *testing.T, recorder *HTTPRecorder) *http.Client {
	t.Helper()
	proxyURL, err := url.Parse("http://" + recorder.Addr())
	require.NoError(t, err)
	return &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		Timeout:   10 * time.Second,
	}
}

// mitmClient trusts both the recorder's generated CA — for hosts it decrypts —
// and the upstream's own certificate, for hosts it tunnels. A real child trusts
// the CA through the environment WriteCA/ProxyEnv hand it.
func mitmClient(t *testing.T, recorder *HTTPRecorder, upstream *httptest.Server) *http.Client {
	t.Helper()
	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(recorder.ca.certPEM))
	pool.AddCert(upstream.Certificate())

	client := proxyClient(t, recorder)
	client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	return client
}

func startRecorder(t *testing.T, opts HTTPOptions) *HTTPRecorder {
	t.Helper()
	recorder, err := StartHTTP(opts)
	require.NoError(t, err)
	t.Cleanup(func() { _ = recorder.Close() })
	return recorder
}

func TestHTTPRecorderRecordsPlainExchange(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "session=supersecret")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer upstream.Close()

	recorder := startRecorder(t, HTTPOptions{Bodies: 1024})

	resp, err := proxyClient(t, recorder).Post(upstream.URL+"/v1/things?page=2",
		"application/json", strings.NewReader(`{"name":"widget"}`))
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.JSONEq(t, `{"ok":true}`, string(body), "capturing the body must not consume it")

	entries := recorder.Entries()
	require.Len(t, entries, 1)
	entry := entries[0]

	assert.Equal(t, http.MethodPost, entry.Request.Method)
	assert.Equal(t, upstream.URL+"/v1/things?page=2", entry.Request.URL)
	assert.Equal(t, []NameValue{{Name: "page", Value: "2"}}, entry.Request.QueryString)
	require.NotNil(t, entry.Request.PostData)
	assert.JSONEq(t, `{"name":"widget"}`, entry.Request.PostData.Text)

	assert.Equal(t, http.StatusOK, entry.Response.Status)
	assert.JSONEq(t, `{"ok":true}`, entry.Response.Content.Text)
	assert.Equal(t, "application/json", entry.Response.Content.MimeType)

	for _, header := range entry.Response.Headers {
		if strings.EqualFold(header.Name, "Set-Cookie") {
			assert.Equal(t, redactedValue, header.Value, "a Set-Cookie must never reach the artifact")
		}
	}
}

func TestHTTPRecorderCapsBodies(t *testing.T) {
	payload := strings.Repeat("x", 500)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, payload)
	}))
	defer upstream.Close()

	recorder := startRecorder(t, HTTPOptions{Bodies: 100})

	resp, err := proxyClient(t, recorder).Get(upstream.URL)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.Len(t, body, len(payload), "the child must still receive the whole response")

	entries := recorder.Entries()
	require.Len(t, entries, 1)
	assert.Len(t, entries[0].Response.Content.Text, 100)
	assert.Contains(t, entries[0].Response.Content.Comment, "truncated")
}

func TestHTTPRecorderWithoutBodiesRecordsMetadataOnly(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "hello")
	}))
	defer upstream.Close()

	recorder := startRecorder(t, HTTPOptions{})

	resp, err := proxyClient(t, recorder).Get(upstream.URL)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	entries := recorder.Entries()
	require.Len(t, entries, 1)
	assert.Empty(t, entries[0].Response.Content.Text, "bodies are opt-in")
	assert.Equal(t, http.StatusOK, entries[0].Response.Status)
}

func TestHTTPRecorderHostsFilter(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer upstream.Close()

	recorder := startRecorder(t, HTTPOptions{Hosts: []string{"*.github.com"}})

	resp, err := proxyClient(t, recorder).Get(upstream.URL)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.Empty(t, recorder.Entries(), "an out-of-scope host must still be proxied but not recorded")
}

func TestHTTPRecorderRecordsConnectTunnel(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "over tls")
	}))
	defer upstream.Close()

	recorder := startRecorder(t, HTTPOptions{Mode: HTTPConnect})

	client := proxyClient(t, recorder)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	transport.TLSClientConfig = upstream.Client().Transport.(*http.Transport).TLSClientConfig

	resp, err := client.Get(upstream.URL)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	// The entry is written when the tunnel closes, which happens after the
	// response has been read.
	transport.CloseIdleConnections()
	require.Eventually(t, func() bool { return len(recorder.Entries()) == 1 }, 5*time.Second, 20*time.Millisecond)

	entry := recorder.Entries()[0]
	assert.Equal(t, http.MethodConnect, entry.Request.Method)
	require.NotNil(t, entry.Gavel)
	assert.True(t, entry.Gavel.Tunnelled)
	assert.Positive(t, entry.Gavel.BytesIn, "a tunnel that carried a response must report bytes in")
	assert.Positive(t, entry.Gavel.BytesOut)
	assert.Empty(t, entry.Response.Content.Text, "connect mode must never claim to have seen the payload")
}

// The whole point of mitm: the payload a connect-mode recording can only report
// the size of is readable, and the request is a full HAR entry rather than a
// tunnel.
func TestHTTPRecorderMITMRecordsDecryptedExchange(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"path":"`+r.URL.Path+`"}`)
	}))
	defer upstream.Close()

	recorder := startRecorder(t, HTTPOptions{Mode: HTTPMITM, Bodies: 1024})
	client := mitmClient(t, recorder, upstream)

	resp, err := client.Get(upstream.URL + "/v1/secrets")
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.JSONEq(t, `{"path":"/v1/secrets"}`, string(body), "decrypting must not disturb the payload")

	entries := recorder.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, http.MethodGet, entries[0].Request.Method)
	assert.Contains(t, entries[0].Request.URL, "/v1/secrets")
	assert.Equal(t, http.StatusOK, entries[0].Response.Status)
	assert.JSONEq(t, `{"path":"/v1/secrets"}`, entries[0].Response.Content.Text)
}

// An out-of-scope host stays encrypted even in mitm mode, so `hosts:` is a
// decryption boundary and not merely a recording filter.
func TestHTTPRecorderMITMLeavesOutOfScopeHostsEncrypted(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "untouched")
	}))
	defer upstream.Close()

	recorder := startRecorder(t, HTTPOptions{Mode: HTTPMITM, Hosts: []string{"*.github.com"}})
	client := mitmClient(t, recorder, upstream)

	resp, err := client.Get(upstream.URL)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	assert.Equal(t, "untouched", string(body))
	assert.Empty(t, recorder.Entries())
}

func TestWriteCAMaterialisesOnlyTheCertificate(t *testing.T) {
	recorder := startRecorder(t, HTTPOptions{Mode: HTTPMITM})
	store := NewStore(t.TempDir(), time.Now())

	require.NoError(t, recorder.WriteCA(store, "example.fixture.md"))
	require.NotEmpty(t, recorder.CAPath())

	pemBytes, err := os.ReadFile(recorder.CAPath())
	require.NoError(t, err)

	block, rest := pem.Decode(pemBytes)
	require.NotNil(t, block)
	assert.Equal(t, "CERTIFICATE", block.Type)
	assert.Empty(t, rest, "the private key must never leave the process")

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	assert.True(t, cert.IsCA)
	assert.WithinDuration(t, time.Now().Add(caValidity), cert.NotAfter, time.Minute,
		"an authority that outlives the run is a liability, not a convenience")

	env := recorder.ProxyEnv(nil)
	for _, key := range []string{"SSL_CERT_FILE", "CURL_CA_BUNDLE", "NODE_EXTRA_CA_CERTS", "REQUESTS_CA_BUNDLE"} {
		assert.Equal(t, recorder.CAPath(), env[key], "%s must point a child at the generated CA", key)
	}
}

func TestWriteCAIsANoOpWithoutMITM(t *testing.T) {
	recorder := startRecorder(t, HTTPOptions{})

	require.NoError(t, recorder.WriteCA(NewStore(t.TempDir(), time.Now()), "example.fixture.md"))
	assert.Empty(t, recorder.CAPath())
	assert.NotContains(t, recorder.ProxyEnv(nil), "SSL_CERT_FILE",
		"a connect-mode recorder has nothing for a child to trust")
}

func TestProxyEnvCoversBothCases(t *testing.T) {
	recorder := startRecorder(t, HTTPOptions{})

	env := recorder.ProxyEnv([]string{"127.0.0.1:8080"})
	url := "http://" + recorder.Addr()

	for _, key := range []string{"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy"} {
		assert.Equal(t, url, env[key], "%s must be set — curl reads the lowercase names, Node the uppercase", key)
	}
	assert.Equal(t, env["NO_PROXY"], env["no_proxy"])
	assert.Contains(t, env["NO_PROXY"], "127.0.0.1:8080", "a daemon port must bypass the proxy or the HAR fills with self-traffic")
	assert.Contains(t, env["NO_PROXY"], recorder.Addr(), "the proxy must not be reachable through itself")
	// Loopback at large stays proxied: a fixture hitting a server it started on
	// 127.0.0.1 is the case the recorder is most useful for.
	assert.NotContains(t, strings.Split(env["NO_PROXY"], ","), "127.0.0.1")
	assert.NotContains(t, strings.Split(env["NO_PROXY"], ","), "localhost")
}

func TestEntriesBetweenNarrowsToWindow(t *testing.T) {
	recorder := &HTTPRecorder{}
	base := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	for _, offset := range []time.Duration{0, time.Second, 2 * time.Second} {
		recorder.add(connectEntry("example.com:443", base.Add(offset), time.Millisecond, 1, 1, nil))
	}

	window := recorder.EntriesBetween(base.Add(500*time.Millisecond), base.Add(1500*time.Millisecond))
	require.Len(t, window, 1)
	assert.Equal(t, base.Add(time.Second).Format(httpTimeFormat), window[0].StartedDateTime)
}

func TestCloseIsIdempotent(t *testing.T) {
	recorder, err := StartHTTP(HTTPOptions{})
	require.NoError(t, err)
	require.NoError(t, recorder.Close())
	require.NoError(t, recorder.Close(), "cleanup runs from both a defer and a shutdown hook")
}
