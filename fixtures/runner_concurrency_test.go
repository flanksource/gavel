package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// overlapFixture writes a fixture whose rows each append a start marker, sleep,
// then append an end marker to a shared log. Serial execution produces strictly
// alternating start/end pairs; any interleaving means two rows ran at once.
func overlapFixture(t *testing.T, rows int) (path, logPath, dir string) {
	t.Helper()
	dir = t.TempDir()
	logPath = filepath.Join(dir, "overlap.log")

	var b strings.Builder
	b.WriteString("---\nexec: bash\nargs: [\"-c\", \"{{.cmd}}\"]\n---\n\n# Overlap\n\n")
	b.WriteString("| Name | cmd | CEL Validation |\n|------|-----|----------------|\n")
	for i := 0; i < rows; i++ {
		// printf is a single write() per call, so the log records ordering
		// without the test racing on partial lines.
		cmd := "printf 'start\\n' >> " + logPath + " && sleep 0.3 && printf 'end\\n' >> " + logPath + " && echo ok"
		b.WriteString("| row " + string(rune('a'+i)) + " | " + cmd + " | stdout.trim() == \"ok\" |\n")
	}

	path = filepath.Join(dir, "overlap.fixture.md")
	require.NoError(t, os.WriteFile(path, []byte(b.String()), 0o600))
	return path, logPath, dir
}

// maxOverlap replays the marker log and returns the highest number of rows that
// were ever in flight simultaneously.
func maxOverlap(t *testing.T, logPath string) int {
	t.Helper()
	raw, err := os.ReadFile(logPath)
	require.NoError(t, err)

	inFlight, peak := 0, 0
	for _, line := range strings.Fields(string(raw)) {
		if line == "start" {
			inFlight++
			if inFlight > peak {
				peak = inFlight
			}
			continue
		}
		inFlight--
	}
	return peak
}

func runOverlapFixture(t *testing.T, maxWorkers, rows int) int {
	t.Helper()
	path, logPath, dir := overlapFixture(t, rows)

	runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: dir, MaxWorkers: maxWorkers})
	require.NoError(t, err)

	tree, err := runner.Run()
	require.NoError(t, err)
	require.NotNil(t, tree)

	return maxOverlap(t, logPath)
}

// MaxWorkers was declared but never read: every fixture row went onto one
// unbounded task group, so a run that asked for serial execution still had four
// rows in flight (clicky's default worker pool). Fixtures shell out to real
// binaries against shared databases and files, so that silently broke every
// caller that had pinned MaxWorkers: 1 to avoid exactly that.
func TestRunnerHonoursMaxWorkers(t *testing.T) {
	require.Equal(t, 1, runOverlapFixture(t, 1, 6), "MaxWorkers: 1 must run one fixture row at a time")
}

// Unset MaxWorkers is serial too: fixture rows share whatever the exec touches,
// so parallelism is opt-in rather than the default.
func TestRunnerDefaultsToSerialExecution(t *testing.T) {
	require.Equal(t, 1, runOverlapFixture(t, 0, 6), "an unset MaxWorkers must default to serial execution")
}

// The limit is a ceiling, not a fixed width: asking for more workers than there
// are rows must not deadlock or drop rows.
func TestRunnerAllowsExplicitParallelism(t *testing.T) {
	require.Greater(t, runOverlapFixture(t, 4, 6), 1, "MaxWorkers: 4 must let rows overlap")
}
