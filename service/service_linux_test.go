//go:build linux

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSystemctlShow(t *testing.T) {
	tests := []struct {
		name        string
		out         string
		wantRunning bool
		wantPID     int
	}{
		{
			name:        "active",
			out:         "ActiveState=active\nMainPID=4242\n",
			wantRunning: true,
			wantPID:     4242,
		},
		{
			name: "inactive",
			// systemd emits MainPID=0 when no process is associated with the
			// unit — treat as not running even if other fields are present.
			out:         "ActiveState=inactive\nMainPID=0\n",
			wantRunning: false,
			wantPID:     0,
		},
		{
			// activating / deactivating have transient pids we can't trust.
			name:        "activating",
			out:         "ActiveState=activating\nMainPID=1234\n",
			wantRunning: false,
			wantPID:     1234,
		},
		{
			name:        "failed",
			out:         "ActiveState=failed\nMainPID=0\n",
			wantRunning: false,
			wantPID:     0,
		},
		{
			name:        "empty",
			out:         "",
			wantRunning: false,
			wantPID:     0,
		},
		{
			name:        "malformed pid",
			out:         "ActiveState=active\nMainPID=nope\n",
			wantRunning: false,
			wantPID:     0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			running, pid := parseSystemctlShow(tc.out)
			assert.Equal(t, tc.wantRunning, running)
			assert.Equal(t, tc.wantPID, pid)
		})
	}
}
