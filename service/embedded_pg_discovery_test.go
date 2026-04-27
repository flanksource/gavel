package service

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writePostmasterPID writes a postmaster.pid file with the given pid and port
// at <dataDir>/data/postmaster.pid, mirroring the layout commons-db creates.
// Lines beyond pid+port aren't required for parsing — we still pad them so
// the file looks plausible.
func writePostmasterPID(t *testing.T, dataDir string, pid, port int) {
	t.Helper()
	dir := filepath.Join(dataDir, "data")
	require.NoError(t, os.MkdirAll(dir, 0o750))
	contents := strings.Join([]string{
		strconv.Itoa(pid),
		dir,
		"1700000000",
		strconv.Itoa(port),
		"/tmp",
		"localhost",
		"541005839 2490370",
		"ready",
		"",
	}, "\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "postmaster.pid"), []byte(contents), 0o600))
}

// listenLoopback binds an ephemeral loopback port and returns the port plus a
// closer; the listener stays open for the test's lifetime so tcpReachable can
// confirm the port is alive.
func listenLoopback(t *testing.T) (int, func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	return port, func() { _ = l.Close() }
}

func TestFindRunningEmbeddedPostgres_ReturnsNilWhenNoPidfile(t *testing.T) {
	withTempHome(t)
	got, err := FindRunningEmbeddedPostgres()
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestFindRunningEmbeddedPostgres_ReturnsRunningWhenPortReachable(t *testing.T) {
	withTempHome(t)
	dataDir, err := EmbeddedDataDir()
	require.NoError(t, err)

	port, stop := listenLoopback(t)
	defer stop()
	writePostmasterPID(t, dataDir, os.Getpid(), port)

	got, err := FindRunningEmbeddedPostgres()
	require.NoError(t, err)
	require.NotNil(t, got, "expected reuse: pid alive + port reachable")
	assert.Equal(t, os.Getpid(), got.PID)
	assert.Equal(t, port, got.Port)
}

func TestFindRunningEmbeddedPostgres_NilWhenPortNotReachable(t *testing.T) {
	withTempHome(t)
	dataDir, err := EmbeddedDataDir()
	require.NoError(t, err)

	// Bind+release a port so we know nothing's listening on it.
	port, stop := listenLoopback(t)
	stop()

	writePostmasterPID(t, dataDir, os.Getpid(), port)
	got, err := FindRunningEmbeddedPostgres()
	require.NoError(t, err)
	assert.Nil(t, got, "stale postmaster.pid pointing at a dead port should not be reused")
}

func TestFindRunningEmbeddedPostgres_NilWhenPidIsDead(t *testing.T) {
	withTempHome(t)
	dataDir, err := EmbeddedDataDir()
	require.NoError(t, err)

	port, stop := listenLoopback(t)
	defer stop()

	// pid=1 on macOS/Linux is launchd/init, but signal(0) can hit EPERM
	// rather than success — to make this deterministic, spawn a short-lived
	// child, capture its pid, wait for it to exit, then write that pid.
	cmd := exec.Command("true")
	require.NoError(t, cmd.Start())
	deadPID := cmd.Process.Pid
	require.NoError(t, cmd.Wait())

	writePostmasterPID(t, dataDir, deadPID, port)
	got, err := FindRunningEmbeddedPostgres()
	require.NoError(t, err)
	assert.Nil(t, got, "dead pid in postmaster.pid should not be reused")
}

func TestFindRunningEmbeddedPostgres_ErrorOnTruncatedFile(t *testing.T) {
	withTempHome(t)
	dataDir, err := EmbeddedDataDir()
	require.NoError(t, err)

	dir := filepath.Join(dataDir, "data")
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "postmaster.pid"), []byte("12345\n"), 0o600))

	_, err = FindRunningEmbeddedPostgres()
	require.Error(t, err, "truncated postmaster.pid (no port line) should surface as error")
}

func TestEmbeddedDSN_Format(t *testing.T) {
	dsn := EmbeddedDSN(54321)
	assert.Equal(t,
		fmt.Sprintf("postgres://%s:%s@localhost:54321/%s?sslmode=disable",
			embeddedPGUser, embeddedPGPassword, embeddedPGDatabase),
		dsn)
}
