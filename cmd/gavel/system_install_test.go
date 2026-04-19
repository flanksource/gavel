package main

import (
	"testing"

	"github.com/flanksource/gavel/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveInstallDBConfig(t *testing.T) {
	tests := []struct {
		name    string
		opts    SystemInstallOptions
		want    service.DBConfig
		wantErr bool
	}{
		{
			name: "dsn",
			opts: SystemInstallOptions{DSN: "postgres://u:p@host/db"},
			want: service.DBConfig{Mode: service.DBModeDSN, DSN: "postgres://u:p@host/db"},
		},
		{
			name: "embedded",
			opts: SystemInstallOptions{Embedded: true},
			want: service.DBConfig{Mode: service.DBModeEmbedded},
		},
		{
			// No flags → embedded is the zero-config default so
			// `gavel system install` Just Works on a fresh machine.
			name: "neither defaults to embedded",
			opts: SystemInstallOptions{},
			want: service.DBConfig{Mode: service.DBModeEmbedded},
		},
		{
			name:    "both",
			opts:    SystemInstallOptions{DSN: "postgres://x", Embedded: true},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveInstallDBConfig(tc.opts)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestVerifyDBConfig_DSNRejectsBogusHost(t *testing.T) {
	// Point at a definitely-unreachable port so pgx.Connect returns quickly.
	err := verifyDBConfig(service.DBConfig{
		Mode: service.DBModeDSN,
		DSN:  "postgres://u:p@127.0.0.1:1/nope?connect_timeout=2",
	})
	assert.Error(t, err)
}

func TestVerifyGitHubToken_SkipBypassesAllProbes(t *testing.T) {
	// --skip-verify-token must succeed without touching the network. We
	// can't easily assert "no network" here, but we can at least assert
	// the function returns nil for a token we know the live probe would
	// reject (too short / obviously fake).
	assert.NoError(t, verifyGitHubToken("not-a-real-token", true))
}

func TestVerifyGitHubToken_UnreachableFailsUnlessSkipped(t *testing.T) {
	// Point the probe at a known-closed port by overriding GITHUB_API_URL
	// — ProbeToken honors it via githubAPIBase(). Port 1 is reserved and
	// almost always refused, so the probe returns Unreachable quickly.
	t.Setenv("GITHUB_API_URL", "http://127.0.0.1:1")
	err := verifyGitHubToken("any-token", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unreachable")

	// --skip bypass should still pass in the same environment.
	assert.NoError(t, verifyGitHubToken("any-token", true))
}
