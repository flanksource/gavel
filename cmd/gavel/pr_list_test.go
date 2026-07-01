package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBindUIListener_BindsRequestedHostAndPort verifies the listener is opened
// on the interface the caller asks for. 0.0.0.0 must produce a wildcard bind
// (the new default that exposes the dashboard on the LAN); 127.0.0.1 must stay
// loopback-only.
func TestBindUIListener_BindsRequestedHostAndPort(t *testing.T) {
	cases := []struct {
		name     string
		host     string
		wantWild bool
	}{
		{"wildcard", "0.0.0.0", true},
		{"loopback", "127.0.0.1", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// port 0 is not the auto-scan path here — we ask the OS for an
			// ephemeral fixed port so the test never races a real daemon on
			// 9092. Bind a probe socket first to learn a free port number.
			probe, err := net.Listen("tcp", net.JoinHostPort(tc.host, "0"))
			require.NoError(t, err)
			port := probe.Addr().(*net.TCPAddr).Port
			require.NoError(t, probe.Close())

			got, ln, err := bindUIListener(tc.host, port)
			require.NoError(t, err)
			t.Cleanup(func() { _ = ln.Close() })

			assert.Equal(t, port, got, "returned port must match the requested fixed port")
			bound := ln.Addr().(*net.TCPAddr)
			assert.Equal(t, port, bound.Port, "listener must be bound to the requested port")
			assert.Equal(t, tc.wantWild, bound.IP.IsUnspecified(),
				"0.0.0.0 must bind the wildcard address; 127.0.0.1 must not")
		})
	}
}

// TestBindUIListener_FixedPortConflictIsSurfaced confirms a fixed --port that
// is already taken fails loudly rather than silently scanning to another port.
func TestBindUIListener_FixedPortConflictIsSurfaced(t *testing.T) {
	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = blocker.Close() })
	port := blocker.Addr().(*net.TCPAddr).Port

	_, _, err = bindUIListener("127.0.0.1", port)
	assert.Error(t, err, "binding an occupied fixed port must error")
}
