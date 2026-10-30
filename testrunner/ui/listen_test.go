package testui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIURLs(t *testing.T) {
	const port = 9092
	cases := []struct {
		name, host, local string
		wildcard          bool
	}{
		{"empty", "", "http://localhost:9092", false},
		{"localhost", "localhost", "http://localhost:9092", false},
		{"loopback v4", "127.0.0.1", "http://localhost:9092", false},
		{"loopback v6", "::1", "http://localhost:9092", false},
		{"explicit ip", "192.168.42.7", "http://192.168.42.7:9092", false},
		{"explicit hostname", "buildhost.lan", "http://buildhost.lan:9092", false},
		{"wildcard v4", "0.0.0.0", "http://localhost:9092", true},
		{"wildcard v6", "::", "http://localhost:9092", true},
	}
	private := firstNonLoopbackIPv4()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			local, network := URLs(tc.host, port)
			assert.Equal(t, tc.local, local)
			// Only a wildcard bind answers on the private IP; a concrete or
			// loopback bind has no second reachable address to announce.
			wantNetwork := ""
			if tc.wildcard && private != "" {
				wantNetwork = "http://" + private + ":9092"
			}
			assert.Equal(t, wantNetwork, network)
		})
	}
}

func TestUIBanner(t *testing.T) {
	assert.Equal(t, "UI at http://localhost:1 (network: http://10.0.0.5:1)", Banner("UI", "http://localhost:1", "http://10.0.0.5:1"))
	assert.Equal(t, "UI at http://localhost:1", Banner("UI", "http://localhost:1", ""))
}

func TestLocalHost(t *testing.T) {
	// The browser and the native menu-bar webview run on this machine, so they
	// dial loopback for every loopback/wildcard bind: macOS WKWebView blocks
	// cleartext HTTP to non-loopback hosts under App Transport Security, and a
	// private IP is fragile (VPN/offline).
	cases := []struct {
		name      string
		requested string
		want      string
	}{
		{"empty", "", "localhost"},
		{"localhost", "localhost", "localhost"},
		{"loopback v4", "127.0.0.1", "localhost"},
		{"loopback v6", "::1", "localhost"},
		{"wildcard v4", "0.0.0.0", "localhost"},
		{"wildcard v6", "::", "localhost"},
		{"explicit ip", "192.168.42.7", "192.168.42.7"},
		{"explicit hostname", "buildhost.lan", "buildhost.lan"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, LocalHost(tc.requested))
		})
	}
}
