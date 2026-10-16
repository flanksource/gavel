// Tests the signed state that carries an install from gavel to GitHub and back: it names
// the org and user, expires after 15 minutes, and rejects any change.
package installations_test

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/github/installations"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallState(t *testing.T) {
	key := []byte(strings.Repeat("k", 32))
	org, user := uuid.New(), uuid.New()
	now := time.Now()
	state := installations.SignState(key, org, user, now)

	got, err := installations.VerifyState(key, state, now.Add(14*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, installations.State{Org: org, User: user, Expires: now.Add(15 * time.Minute).Unix()}, got)

	for name, test := range map[string]struct {
		key   []byte
		state string
		at    time.Time
		want  string
	}{
		"expired":     {key, state, now.Add(16 * time.Minute), "the GitHub install link expired"},
		"wrong key":   {[]byte(strings.Repeat("x", 32)), state, now, "isn't valid"},
		"tampered":    {key, tamper(t, state), now, "isn't valid"},
		"no dot":      {key, "abc", now, "isn't valid"},
		"empty":       {key, "", now, "isn't valid"},
		"bad payload": {key, "!!!." + strings.Split(state, ".")[1], now, "isn't valid"},
	} {
		_, err := installations.VerifyState(test.key, test.state, test.at)
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), test.want, name)
	}
}

// tamper changes the signed payload to name another org, keeping the signature.
func tamper(t *testing.T, state string) string {
	t.Helper()
	payload, signature, _ := strings.Cut(state, ".")
	data, err := base64.RawURLEncoding.DecodeString(payload)
	require.NoError(t, err)
	changed := strings.Replace(string(data), `"org":"`, `"org":"0`, 1)
	return base64.RawURLEncoding.EncodeToString([]byte(changed)) + "." + signature
}
