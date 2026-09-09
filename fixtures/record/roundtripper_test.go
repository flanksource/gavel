package record

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clientServer answers the two shapes a HAR has to distinguish: a success with
// a body and a secret header, and an error status.
func clientServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "session=super-secret")
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// recordingClient wires a client the way internal/httpx does and guarantees the
// sink is detached again, since it is process-global.
func recordingClient(t *testing.T, opts ClientOptions) (*http.Client, *ClientLog) {
	t.Helper()
	log := StartClients(opts)
	t.Cleanup(StopClients)
	return &http.Client{Transport: MaybeWrap(http.DefaultTransport)}, log
}

func TestTransportRecordsNothingUntilAFixtureAsks(t *testing.T) {
	server := clientServer(t)
	client := &http.Client{Transport: MaybeWrap(http.DefaultTransport)}

	resp, err := client.Get(server.URL + "/ok")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	log := StartClients(ClientOptions{})
	t.Cleanup(StopClients)
	assert.Empty(t, log.Entries(), "a request made before the recording started is not part of it")
}

func TestTransportRecordsRequestsWhileAttached(t *testing.T) {
	server := clientServer(t)
	client, log := recordingClient(t, ClientOptions{})

	for _, path := range []string{"/ok", "/missing"} {
		resp, err := client.Get(server.URL + path)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	}

	entries := log.Entries()
	require.Len(t, entries, 2)
	assert.Equal(t, http.MethodGet, entries[0].Request.Method)
	assert.Equal(t, server.URL+"/ok", entries[0].Request.URL)
	assert.Equal(t, http.StatusOK, entries[0].Response.Status)
	assert.Equal(t, http.StatusNotFound, entries[1].Response.Status)
}

func TestTransportRedactsSecretHeaders(t *testing.T) {
	server := clientServer(t)
	client, log := recordingClient(t, ClientOptions{})

	req, err := http.NewRequest(http.MethodGet, server.URL+"/ok", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer nobody-should-see-this")

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	entries := log.Entries()
	require.Len(t, entries, 1)

	headers := map[string]string{}
	for _, header := range entries[0].Request.Headers {
		headers[strings.ToLower(header.Name)] = header.Value
	}
	assert.Equal(t, redactedValue, headers["authorization"],
		"gavel's own clients carry the tokens the recorder exists to keep out of artifacts")

	for _, header := range entries[0].Response.Headers {
		if strings.EqualFold(header.Name, "Set-Cookie") {
			assert.Equal(t, redactedValue, header.Value)
		}
	}
}

func TestTransportKeepsBodiesOnlyWhenAsked(t *testing.T) {
	server := clientServer(t)

	t.Run("off by default", func(t *testing.T) {
		client, log := recordingClient(t, ClientOptions{})
		resp, err := client.Get(server.URL + "/ok")
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		require.Len(t, log.Entries(), 1)
		assert.Empty(t, log.Entries()[0].Response.Content.Text)
	})

	t.Run("captured and still readable", func(t *testing.T) {
		client, log := recordingClient(t, ClientOptions{Bodies: Size(1024)})
		resp, err := client.Post(server.URL+"/ok", "application/json", strings.NewReader(`{"send":true}`))
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		assert.JSONEq(t, `{"status":"ok"}`, string(body), "capturing a body must not consume it")

		require.Len(t, log.Entries(), 1)
		entry := log.Entries()[0]
		assert.Equal(t, `{"status":"ok"}`, entry.Response.Content.Text)
		require.NotNil(t, entry.Request.PostData)
		assert.Equal(t, `{"send":true}`, entry.Request.PostData.Text)
	})
}

// A request that never got an answer is the one a failing fixture most needs to
// see, so a transport error is an entry rather than a dropped exchange.
func TestTransportRecordsTransportErrors(t *testing.T) {
	_, log := recordingClient(t, ClientOptions{})
	client := &http.Client{Transport: MaybeWrap(failingTransport{})}

	_, err := client.Get("http://example.invalid/boom")
	require.Error(t, err)

	entries := log.Entries()
	require.Len(t, entries, 1)
	require.NotNil(t, entries[0].Gavel)
	assert.Contains(t, entries[0].Gavel.Error, "no route to anywhere")
	assert.Zero(t, entries[0].Response.Status, "there was no response, so there is no status to invent")
}

type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("no route to anywhere")
}

func TestBetweenNarrowsToOneFixturesWindow(t *testing.T) {
	server := clientServer(t)
	client, log := recordingClient(t, ClientOptions{})

	resp, err := client.Get(server.URL + "/ok")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	start := time.Now()
	resp, err = client.Get(server.URL + "/missing")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	window := log.Between(start, time.Now())
	require.Len(t, window, 1, "an earlier fixture's traffic must not land in this one's artifact")
	assert.Equal(t, server.URL+"/missing", window[0].Request.URL)
}

func TestMaybeWrapIsIdempotentAndNilSafe(t *testing.T) {
	wrapped := MaybeWrap(nil)
	require.IsType(t, &harTransport{}, wrapped)
	assert.Equal(t, http.DefaultTransport, wrapped.(*harTransport).base, "a nil base means the default transport")

	assert.Same(t, wrapped, MaybeWrap(wrapped), "wrapping twice would record every request twice")
}

func TestClientLogSavesAHARUnderItsOwnKind(t *testing.T) {
	server := clientServer(t)
	client, log := recordingClient(t, ClientOptions{})

	resp, err := client.Get(server.URL + "/ok")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	store := NewStore(t.TempDir(), time.Now())
	result, err := log.Save(store, "clients", log.Entries())
	require.NoError(t, err)

	assert.Equal(t, KindClients, result.Kind, "only the kind tells a client recording from a proxy one")
	assert.Equal(t, "har-1.2", result.Format)
	assert.Equal(t, 1, result.Count)
	assert.Positive(t, result.Bytes)

	har, err := os.ReadFile(result.Path)
	require.NoError(t, err)
	assert.Contains(t, string(har), "/ok")
	assert.NotContains(t, string(har), "super-secret")
}
