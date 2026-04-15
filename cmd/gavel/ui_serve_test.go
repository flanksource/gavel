package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeSnapshot writes a minimal JSON payload that runUIServe can load.
func writeSnapshot(t *testing.T, dir string) string {
	t.Helper()
	payload := snapshotPayload{
		Tests: []parsers.Test{
			{Name: "TestA", Package: "pkg/a", Passed: true, Framework: parsers.GoTest},
			{Name: "TestB", Package: "pkg/b", Failed: true, Framework: parsers.GoTest, Message: "boom"},
		},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	path := filepath.Join(dir, "snapshot.json")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

// freePort binds and closes a listener to discover an ephemeral port number.
// Only used so the test can predict the URL before starting the server.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return port
}

func TestLoadResults_PopulatesServer(t *testing.T) {
	// loadResults is the replay path used by both standalone and
	// detached-child modes: read a JSON file written by the fork parent,
	// rehydrate a fresh testui.Server, and serve it to the /api/tests
	// endpoint as if the run had just finished.
	dir := t.TempDir()
	path := writeSnapshot(t, dir)

	srv := testui.NewServer()
	require.NoError(t, loadResults(srv, path))

	handler := srv.Handler()
	req, _ := http.NewRequest("GET", "/api/tests", nil)
	rw := &recordingResponse{header: http.Header{}}
	handler.ServeHTTP(rw, req)

	var got map[string]any
	require.NoError(t, json.Unmarshal(rw.body, &got))
	tests, ok := got["tests"].([]any)
	require.True(t, ok, "tests field missing from snapshot: %s", rw.body)
	require.Len(t, tests, 2)
}

// recordingResponse is a minimal http.ResponseWriter for in-process tests
// that don't need a full httptest.Server.
type recordingResponse struct {
	header http.Header
	status int
	body   []byte
}

func (r *recordingResponse) Header() http.Header { return r.header }
func (r *recordingResponse) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
func (r *recordingResponse) WriteHeader(code int) { r.status = code }

func TestServeUntilTimeout_IdleFiresFirst(t *testing.T) {
	idleCh := make(chan struct{}, 1)
	start := time.Now()
	err := serveUntilTimeout(5*time.Second, 50*time.Millisecond, idleCh)
	elapsed := time.Since(start)
	require.NoError(t, err)
	// Should exit roughly at idle-timeout, not hard-timeout.
	assert.Less(t, elapsed, 500*time.Millisecond, "expected idle timer to win, took %s", elapsed)
}

func TestServeUntilTimeout_HardFiresFirst(t *testing.T) {
	idleCh := make(chan struct{}, 1)
	start := time.Now()
	err := serveUntilTimeout(50*time.Millisecond, 5*time.Second, idleCh)
	elapsed := time.Since(start)
	require.NoError(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond, "expected hard timer to win, took %s", elapsed)
}

func TestServeUntilTimeout_ActivityResetsIdle(t *testing.T) {
	idleCh := make(chan struct{}, 4)
	done := make(chan error, 1)

	go func() {
		done <- serveUntilTimeout(0, 150*time.Millisecond, idleCh)
	}()

	// Send 3 pokes spaced 80ms apart. Each one resets the 150ms idle timer,
	// so the server must stay alive for at least 3*80ms = 240ms even though
	// the idle timeout is 150ms.
	for range 3 {
		time.Sleep(80 * time.Millisecond)
		idleCh <- struct{}{}
	}

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("serveUntilTimeout did not exit after activity window")
	}
}

func TestWriteURLFile_Atomic(t *testing.T) {
	// writeURLFile must write atomically because a wrapping shell script may
	// be polling the file concurrently. Verify the final file contains the
	// URL and that no stray tempfiles are left behind.
	dir := t.TempDir()
	path := filepath.Join(dir, "url.txt")
	require.NoError(t, writeURLFile(path, "http://localhost:43123"))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:43123\n", string(data))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "expected only the final file, found: %v", entries)
}

func TestRunUIServe_ReplaysSnapshotAndExits(t *testing.T) {
	// Full integration: drive runUIServe with a snapshot, short idle, hit the
	// endpoint once, and verify the handler returned the loaded tests and
	// the function exited cleanly within the expected window.
	dir := t.TempDir()
	snapshot := writeSnapshot(t, dir)
	urlFile := filepath.Join(dir, "url.txt")

	port := freePort(t)
	opts := UIServeOptions{
		Port:        port,
		ResultsFile: snapshot,
		AutoStop:    5 * time.Second,
		IdleTimeout: 200 * time.Millisecond,
		URLFile:     urlFile,
	}

	done := make(chan error, 1)
	go func() {
		_, err := runUIServe(opts)
		done <- err
	}()

	// Wait for the server to write the URL file — that's the "ready" signal.
	deadline := time.Now().Add(2 * time.Second)
	var url string
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(urlFile); err == nil && len(data) > 0 {
			url = string(data[:len(data)-1]) // strip trailing newline
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NotEmpty(t, url, "server did not write URL file in time")

	// Hit the snapshot endpoint and verify both loaded tests come back.
	resp, err := http.Get(url + "/api/tests")
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)

	var snap map[string]any
	require.NoError(t, json.Unmarshal(body, &snap))
	tests, _ := snap["tests"].([]any)
	assert.Len(t, tests, 2)
	assert.Equal(t, true, snap["done"], "server should be marked done after snapshot load")

	// Don't poke the server again — let the idle timer fire.
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("runUIServe did not auto-stop within 3s")
	}
}

func TestAnnounceHost(t *testing.T) {
	external := firstNonLoopbackIPv4()

	cases := []struct {
		name      string
		requested string
		want      string
	}{
		{"empty", "", "localhost"},
		{"localhost", "localhost", "localhost"},
		{"loopback v4", "127.0.0.1", "localhost"},
		{"loopback v6", "::1", "localhost"},
		{"explicit ip", "192.168.42.7", "192.168.42.7"},
		{"explicit hostname", "buildhost.lan", "buildhost.lan"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, announceHost(tc.requested))
		})
	}

	t.Run("wildcard 0.0.0.0", func(t *testing.T) {
		got := announceHost("0.0.0.0")
		if external == "" {
			assert.Equal(t, "localhost", got, "no non-loopback iface; expect fallback")
		} else {
			assert.Equal(t, external, got, "expected first non-loopback ipv4")
		}
	})

	t.Run("wildcard ::", func(t *testing.T) {
		got := announceHost("::")
		if external == "" {
			assert.Equal(t, "localhost", got)
		} else {
			assert.Equal(t, external, got)
		}
	})
}
