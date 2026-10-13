// Tests that a configured TokenSource is the only credential an Options uses, so a
// hosted organization's client never falls back to the host's GitHub token.
package github

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenSourceNeverFallsBackToHostCredentials(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "host-token")
	t.Setenv("GH_TOKEN", "host-token")

	token, err := Options{TokenSource: func() (string, error) { return "org-token", nil }}.token()
	require.NoError(t, err)
	assert.Equal(t, "org-token", token)

	_, err = Options{TokenSource: func() (string, error) { return "", errors.New("connect GitHub first") }}.token()
	require.EqualError(t, err, "connect GitHub first")

	_, err = Options{TokenSource: func() (string, error) { return "", nil }}.token()
	require.ErrorIs(t, err, ErrNoToken)
	assert.EqualError(t, err, "no GitHub token: the token source returned an empty token")

	token, err = Options{}.token()
	require.NoError(t, err)
	assert.Equal(t, "host-token", token, "without a TokenSource the env fallback still applies")
}
