// Tests the CLI against a hosted gavel: `login --server` discovers Clerk from the server,
// `org list/use` picks the organization, and `ssh-key` manages the user's keys.
package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

// runCommand runs the command at path with args and flags, returning its stdout.
func runCommand(t *testing.T, path []string, args []string, flags map[string]string) (string, error) {
	t.Helper()
	command, _, err := rootCmd.Find(path)
	require.NoError(t, err)
	for name, value := range flags {
		require.NoError(t, command.Flags().Set(name, value))
	}
	t.Cleanup(func() {
		for name := range flags {
			flag := command.Flags().Lookup(name)
			_ = flag.Value.Set(flag.DefValue)
			flag.Changed = false
		}
	})
	command.SetContext(context.Background())
	var out, errOut bytes.Buffer
	command.SetOut(&out)
	command.SetErr(&errOut)
	defer command.SetOut(nil)
	defer command.SetErr(nil)
	err = command.RunE(command, args)
	return out.String() + errOut.String(), err
}

func savedSession(t *testing.T) *api.Session {
	t.Helper()
	path, err := api.CredentialPath()
	require.NoError(t, err)
	session, err := api.LoadSession(path)
	require.NoError(t, err)
	return session
}

func TestServerCommands(t *testing.T) {
	f := newHostedFixture(t)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv(database.EnvDisable, "off")
	original := gavelServerHTTPClient
	gavelServerHTTPClient = f.server.Client()
	t.Cleanup(func() { gavelServerHTTPClient = original })

	_, err := runCommand(t, []string{"org", "list"}, nil, nil)
	require.ErrorContains(t, err, "gavel login --server")

	out, err := runCommand(t, []string{"login"}, nil, map[string]string{"server": f.server.URL, "no-browser": "true"})
	require.NoError(t, err)
	assert.Contains(t, out, "gavel org use", "two organizations: the user is told to pick one")
	session := savedSession(t)
	assert.Equal(t, f.server.URL, session.Server)
	assert.Equal(t, f.clerk.URL, session.Config.Issuer, "the issuer comes from the server")
	assert.Empty(t, session.Org)

	out, err = runCommand(t, []string{"org", "list"}, nil, nil)
	require.NoError(t, err)
	assert.Contains(t, out, "alpha")
	assert.Contains(t, out, "beta")

	_, err = runCommand(t, []string{"org", "use"}, []string{"gamma"}, nil)
	require.ErrorContains(t, err, `not a member of "gamma"; your organizations: alpha, beta`)
	_, err = runCommand(t, []string{"org", "use"}, []string{"beta"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "beta", savedSession(t).Org)

	public, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	key, err := gossh.NewPublicKey(public)
	require.NoError(t, err)
	keyPath := filepath.Join(t.TempDir(), "id_ed25519.pub")
	require.NoError(t, os.WriteFile(keyPath, gossh.MarshalAuthorizedKey(key), 0o600))

	out, err = runCommand(t, []string{"ssh-key", "add"}, []string{keyPath}, map[string]string{"name": "laptop"})
	require.NoError(t, err)
	assert.Contains(t, out, gossh.FingerprintSHA256(key))
	_, err = runCommand(t, []string{"ssh-key", "add"}, []string{keyPath}, map[string]string{"name": "again"})
	require.ErrorContains(t, err, "already registered")

	out, err = runCommand(t, []string{"ssh-key", "list"}, nil, nil)
	require.NoError(t, err)
	assert.Contains(t, out, "laptop")
	var id string
	for _, line := range strings.Split(out, "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && strings.Contains(line, "laptop") {
			for _, field := range fields {
				if len(field) == 36 && strings.Count(field, "-") == 4 {
					id = field
				}
			}
		}
	}
	require.NotEmpty(t, id, "list shows each key's ID: %s", out)
	_, err = runCommand(t, []string{"ssh-key", "rm"}, []string{id}, nil)
	require.NoError(t, err)
	out, err = runCommand(t, []string{"ssh-key", "list"}, nil, nil)
	require.NoError(t, err)
	assert.NotContains(t, out, "laptop")
}
