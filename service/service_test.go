package service

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withTempHome isolates StateDir()/PidFile()/LogFile() from the real user
// environment by pointing $HOME at a temp directory for the duration of the
// test.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestStateDir_CreatesUnderHome(t *testing.T) {
	home := withTempHome(t)
	dir, err := StateDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "gavel"), dir)
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestReadStatus_NoPidFile(t *testing.T) {
	withTempHome(t)
	st, err := ReadStatus()
	require.NoError(t, err)
	assert.False(t, st.Running)
	assert.False(t, st.Stale)
	assert.Zero(t, st.PID)
}

func TestReadStatus_StalePidFile(t *testing.T) {
	withTempHome(t)
	pf, err := PidFile()
	require.NoError(t, err)
	// PID 1 is init and (likely) alive, so pick a PID that is very unlikely
	// to exist — the highest 31-bit value. processAlive() returns false
	// because Signal(0) fails with ESRCH.
	require.NoError(t, os.WriteFile(pf, []byte(strconv.Itoa(0x7fffffff)), 0o644))

	st, err := ReadStatus()
	require.NoError(t, err)
	assert.False(t, st.Running)
	assert.True(t, st.Stale, "expected stale pidfile to be flagged")
}

func TestReadStatus_LiveProcess(t *testing.T) {
	withTempHome(t)
	// Use `sleep` as a cheap live process that stays alive during the test.
	cmd := exec.Command("sleep", "5")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	pf, err := PidFile()
	require.NoError(t, err)
	require.NoError(t, writePid(pf, cmd.Process.Pid))

	st, err := ReadStatus()
	require.NoError(t, err)
	assert.True(t, st.Running)
	assert.False(t, st.Stale)
	assert.Equal(t, cmd.Process.Pid, st.PID)
}

func TestStop_NoProcessIsNoOp(t *testing.T) {
	withTempHome(t)
	// No pidfile -> Stop returns nil without error.
	assert.NoError(t, Stop(time.Second))
}

func TestStop_RemovesStalePidFile(t *testing.T) {
	withTempHome(t)
	pf, err := PidFile()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(pf, []byte(strconv.Itoa(0x7fffffff)), 0o644))

	require.NoError(t, Stop(time.Second))
	_, err = os.Stat(pf)
	assert.True(t, os.IsNotExist(err), "stale pidfile should have been removed")
}

func TestTailLog_MissingFileReturnsEmpty(t *testing.T) {
	withTempHome(t)
	lines, err := TailLog(25)
	require.NoError(t, err)
	assert.Empty(t, lines)
}

func TestTailLog_FewerLinesThanRequested(t *testing.T) {
	withTempHome(t)
	path, err := LogFile()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte("a\nb\nc\n"), 0o644))

	lines, err := TailLog(25)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, lines)
}

func TestTailLog_ReturnsOnlyLastN(t *testing.T) {
	withTempHome(t)
	path, err := LogFile()
	require.NoError(t, err)
	// 50 lines — larger than 25 and larger than one 8KiB chunk if verbose,
	// but still tiny so the test stays fast. The key invariant: exactly the
	// last 25 come back, oldest-first.
	var body []byte
	for i := range 50 {
		body = fmt.Appendf(body, "line-%02d\n", i)
	}
	require.NoError(t, os.WriteFile(path, body, 0o644))

	lines, err := TailLog(25)
	require.NoError(t, err)
	require.Len(t, lines, 25)
	assert.Equal(t, "line-25", lines[0])
	assert.Equal(t, "line-49", lines[24])
}

func TestTailLog_LargeFileReadsInChunks(t *testing.T) {
	withTempHome(t)
	path, err := LogFile()
	require.NoError(t, err)
	// Write >8KiB so the tail loop has to read multiple chunks. Each line is
	// ~20 bytes; 1000 lines ≈ 20KiB.
	var body []byte
	for i := range 1000 {
		body = fmt.Appendf(body, "chunk-line-%04d\n", i)
	}
	require.NoError(t, os.WriteFile(path, body, 0o644))

	lines, err := TailLog(10)
	require.NoError(t, err)
	require.Len(t, lines, 10)
	assert.Equal(t, "chunk-line-0990", lines[0])
	assert.Equal(t, "chunk-line-0999", lines[9])
}

func TestWaitForReady_NoPidfileEventuallyReturnsCrashed(t *testing.T) {
	withTempHome(t)
	// No pidfile means ReadStatus().Running == false. WaitForReady waits
	// out a 3s grace window (to tolerate launchctl kickstart churn) before
	// declaring that a genuine crash — pass a timeout larger than that so
	// the grace expires first.
	got, err := WaitForReady(4 * time.Second)
	require.NoError(t, err)
	assert.Equal(t, ReadinessCrashed, got)
}

func TestWaitForReady_LiveProcessButPortClosedTimesOut(t *testing.T) {
	withTempHome(t)
	cmd := exec.Command("sleep", "5")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	pf, err := PidFile()
	require.NoError(t, err)
	require.NoError(t, writePid(pf, cmd.Process.Pid))

	// Use a freshly-closed TCP port so we know nothing is listening — the
	// dev machine's real :9092 may be occupied by a prior daemon which
	// would make this test observe ReadinessReady instead of TimedOut.
	port := freeTCPPort(t)

	got, err := waitForReady(600*time.Millisecond, port)
	require.NoError(t, err)
	assert.Equal(t, ReadinessTimedOut, got)
}

// freeTCPPort reserves an ephemeral port and immediately releases it —
// the returned port is guaranteed-closed for the brief window of the test.
func freeTCPPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return port
}

func TestTruncateLogFile_EmptiesExistingLog(t *testing.T) {
	withTempHome(t)
	path, err := LogFile()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte("stale content\n"), 0o644))

	require.NoError(t, truncateLogFile())

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.Size(), "log should be zero bytes after truncate")
}

func TestTruncateLogFile_MissingFileIsNoError(t *testing.T) {
	withTempHome(t)
	// No logfile present — truncate should create an empty one without
	// complaining. (Start() may run before any write has ever happened.)
	require.NoError(t, truncateLogFile())

	path, err := LogFile()
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.Size())
}

func TestScanFreePort_FindsAvailablePort(t *testing.T) {
	// Occupy a port so the scanner has to skip past it. Using port 0 gets
	// us an OS-assigned free port whose number we can feed back in as
	// "start" — the scanner should return start+1.
	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = blocker.Close() })
	start := blocker.Addr().(*net.TCPAddr).Port

	port, listener, err := ScanFreePort(start, 10)
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	assert.Greater(t, port, start, "scanner should skip past the occupied port")
	assert.Equal(t, port, listener.Addr().(*net.TCPAddr).Port, "returned port must match the bound listener")
}

func TestScanFreePort_ExhaustedReturnsError(t *testing.T) {
	// Occupy N consecutive ports, then ask the scanner to try only N — it
	// should fail. We grab the ports by listening on :0 repeatedly; the
	// assigned ports are usually sequential enough for this to be tight,
	// but we defensively scan from the smallest up to tighten the window.
	var occupied []net.Listener
	t.Cleanup(func() {
		for _, l := range occupied {
			_ = l.Close()
		}
	})
	for range 3 {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		occupied = append(occupied, l)
	}
	minPort := occupied[0].Addr().(*net.TCPAddr).Port
	for _, l := range occupied[1:] {
		if p := l.Addr().(*net.TCPAddr).Port; p < minPort {
			minPort = p
		}
	}
	// tries=1 guarantees exhaustion — start is occupied, we don't try any
	// neighbors.
	_, _, err := ScanFreePort(minPort, 1)
	assert.Error(t, err)
}

func TestWriteAndReadUIPort_Roundtrip(t *testing.T) {
	withTempHome(t)
	require.NoError(t, WriteUIPort(12345))
	assert.Equal(t, 12345, ReadUIPort())
}

func TestReadUIPort_MissingFileFallsBackToDefault(t *testing.T) {
	withTempHome(t)
	// No port file on disk → fallback to DefaultUIPort so upgraders
	// without a stored port (fresh install / pre-rename) still poll the
	// expected address.
	assert.Equal(t, DefaultUIPort, ReadUIPort())
}

func TestReadUIPort_MalformedFileFallsBackToDefault(t *testing.T) {
	withTempHome(t)
	path, err := PortFile()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte("not-a-port\n"), 0o644))
	assert.Equal(t, DefaultUIPort, ReadUIPort())
}

func TestStop_TerminatesLiveProcess(t *testing.T) {
	withTempHome(t)
	// `sleep` on macOS doesn't always honor SIGTERM promptly, so we accept
	// the SIGKILL-escalation path here. The contract being verified is:
	// Stop() returns with the process dead and the pidfile removed.
	cmd := exec.Command("sleep", "60")
	require.NoError(t, cmd.Start())

	pf, err := PidFile()
	require.NoError(t, err)
	require.NoError(t, writePid(pf, cmd.Process.Pid))

	require.NoError(t, Stop(500*time.Millisecond))
	_, _ = cmd.Process.Wait()
	assert.False(t, processAlive(cmd.Process.Pid), "process should have exited")
	_, err = os.Stat(pf)
	assert.True(t, os.IsNotExist(err), "pidfile should have been removed")
}
