package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestANSICommandRegistered(t *testing.T) {
	cmd, _, err := testCmd.Find([]string{"ansi"})
	require.NoError(t, err)
	require.NotNil(t, cmd)
	assert.Equal(t, "ansi", cmd.Name())
}

func TestRunTestANSIRequiresCommandAfterDash(t *testing.T) {
	// out.json given but no `--` and no command → ArgsLenAtDash() is -1 and the
	// command must refuse rather than treat the path as the command.
	require.NoError(t, testANSICmd.Flags().Parse([]string{"out.json"}))
	_, err := runTestANSI(testANSIOptions{Interval: "50ms", Args: testANSICmd.Flags().Args()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "after `--`")
}

func TestRunTestANSIWritesCaptureJSON(t *testing.T) {
	out := filepath.Join(t.TempDir(), "cap.json")
	require.NoError(t, testANSICmd.Flags().Parse([]string{out, "--", "/bin/sh", "-c", "printf 'hi\\n'"}))

	res, err := runTestANSI(testANSIOptions{Interval: "50ms", Args: testANSICmd.Flags().Args()})
	require.NoError(t, err)

	summary, ok := res.(ansiSummary)
	require.True(t, ok)
	assert.Equal(t, out, summary.Out)
	assert.Equal(t, 0, summary.ExitCode)
	assert.Greater(t, summary.Events, 0)

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	var capture map[string]any
	require.NoError(t, json.Unmarshal(data, &capture))
	for _, key := range []string{"version", "width", "height", "command", "exit_code", "duration_ms", "events", "snapshots", "final"} {
		assert.Contains(t, capture, key, "out.json should contain %q", key)
	}
	final, ok := capture["final"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, final, "screen")
}
