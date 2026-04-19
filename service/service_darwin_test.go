//go:build darwin

package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPlist_ContainsExpectedDirectives(t *testing.T) {
	out, err := renderPlist("/usr/local/bin/gavel", "/tmp/pr-ui.log")
	require.NoError(t, err)

	// All three are load-bearing for a background agent that auto-starts,
	// survives logout of its controlling terminal, and restarts on crash.
	for _, want := range []string{
		"<string>" + launchdLabel + "</string>",
		"<string>/usr/local/bin/gavel</string>",
		"<string>pr</string>",
		"<string>--all</string>",
		"<string>--ui</string>",
		"<string>--menu-bar</string>",
		"<string>--port=0</string>",
		"<string>--persist-port</string>",
		"<key>RunAtLoad</key><true/>",
		"<key>KeepAlive</key>",
		"<key>SuccessfulExit</key><false/>",
		"<string>/tmp/pr-ui.log</string>",
	} {
		assert.True(t, strings.Contains(out, want), "plist missing %q\n---\n%s", want, out)
	}
}

func TestParseLaunchctlPrint(t *testing.T) {
	// Abbreviated samples from real `launchctl print gui/<uid>/<label>` output
	// — the parser only cares about the state = X and pid = N lines, so the
	// surrounding envelope is trimmed.
	runningSample := `gui/501/com.flanksource.gavel = {
	type = LaunchAgent
	program = /usr/local/bin/gavel
	state = running
	domain = gui/501
	pid = 54321
	default environment = {
	}
}`
	stoppedSample := `gui/501/com.flanksource.gavel = {
	type = LaunchAgent
	state = not running
}`
	tests := []struct {
		name        string
		out         string
		wantRunning bool
		wantPID     int
	}{
		{"running", runningSample, true, 54321},
		{"stopped", stoppedSample, false, 0},
		{"empty", "", false, 0},
		{"pid zero", "state = running\npid = 0\n", false, 0},
		{"state missing", "pid = 123\n", false, 123},
		{"malformed pid", "state = running\npid = not-a-number\n", false, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			running, pid := parseLaunchctlPrint(tc.out)
			assert.Equal(t, tc.wantRunning, running)
			assert.Equal(t, tc.wantPID, pid)
		})
	}
}
