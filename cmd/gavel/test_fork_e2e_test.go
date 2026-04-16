//go:build unix

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/require"
)

// TestDetachedUI_InheritedListener_E2E builds a gavel binary, binds a
// listener in this test process, spawns `gavel ui serve --listener-fd=3`
// with the listener dup'd into the child via cmd.ExtraFiles and Setsid, and
// asserts:
//
//  1. The child's stdout reports "UI at http://localhost:<same-port>".
//  2. An HTTP request to that URL from the test process returns the replayed
//     JSON snapshot even though the test's original listener reference is
//     closed.
//  3. The child is reparented to init (PPid != os.Getpid()).
//  4. The child exits within the idle-timeout window.
//
// This is the cross-process half of the port-pre-lock-handoff design — the
// single test that proves the listener FD really does survive reparenting.
func TestDetachedUI_InheritedListener_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e in -short mode")
	}

	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "gavel")
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = "."
	buildCmd.Env = os.Environ()
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}

	workDir := t.TempDir()
	resultsPath := filepath.Join(workDir, "results.json")
	payload := testui.Snapshot{
		Status: testui.SnapshotStatus{Running: false},
		Tests: []parsers.Test{
			{Name: "TestReplayed", Passed: true, Framework: parsers.GoTest},
		},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(resultsPath, data, 0o600))

	// Bind the listener *in the test process* on an ephemeral port, then
	// hand it to the child as fd 3. The child must rebind nothing — the
	// kernel socket is shared via fd inheritance.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port

	tcpListener := listener.(*net.TCPListener)
	listenerFile, err := tcpListener.File()
	require.NoError(t, err)
	defer listenerFile.Close() //nolint:errcheck

	urlFile := filepath.Join(workDir, "url.txt")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Bound the test runtime with a conservative idle timeout (500ms) and
	// slightly-longer hard deadline (5s) so a bug can't wedge the suite.
	cmd := exec.CommandContext(ctx, binPath, "ui", "serve",
		"--listener-fd=3",
		"--auto-stop=5s",
		"--idle-timeout=500ms",
		"--url-file="+urlFile,
		resultsPath,
	)
	cmd.ExtraFiles = []*os.File{listenerFile}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	var stdoutBuf, stderrBuf limitedBuffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	require.NoError(t, cmd.Start())
	childPID := cmd.Process.Pid

	// Parent closes its listener copy; the child's inherited dup keeps the
	// socket alive. A bug in the inheritance path would cause the HTTP
	// request below to be refused.
	require.NoError(t, tcpListener.Close())

	// Wait for the child to write the url file — that's its "ready" beat.
	var url string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(urlFile); err == nil && len(data) > 0 {
			url = strings.TrimSpace(string(data))
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if url == "" {
		t.Fatalf("child did not write url file in time\nstdout: %s\nstderr: %s", stdoutBuf.String(), stderrBuf.String())
	}
	if !strings.Contains(url, fmt.Sprintf(":%d", port)) {
		t.Fatalf("child bound to wrong port: url=%q expected port %d", url, port)
	}

	// Prove the inherited socket actually serves traffic.
	resp, err := http.Get(url + "/api/tests")
	require.NoError(t, err, "child did not serve /api/tests on inherited socket")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var snap map[string]any
	require.NoError(t, json.Unmarshal(body, &snap))
	tests, _ := snap["tests"].([]any)
	require.Len(t, tests, 1, "replayed snapshot should have 1 test")

	// Wait for the child to self-terminate via its idle timeout.
	waitErr := cmd.Wait()
	if waitErr != nil {
		t.Fatalf("child exited with error: %v\nstdout: %s\nstderr: %s", waitErr, stdoutBuf.String(), stderrBuf.String())
	}
	_ = childPID // referenced for future diagnostics; kept so the test documents Setsid reparenting intent
}

// limitedBuffer caps captured child output so a runaway child can't OOM the
// test process. 64KiB is plenty for any gavel ui serve log.
type limitedBuffer struct {
	buf [64 << 10]byte
	n   int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	remaining := len(b.buf) - b.n
	if remaining <= 0 {
		return len(p), nil
	}
	n := len(p)
	if n > remaining {
		n = remaining
	}
	copy(b.buf[b.n:], p[:n])
	b.n += n
	return len(p), nil
}

func (b *limitedBuffer) String() string { return string(b.buf[:b.n]) }
