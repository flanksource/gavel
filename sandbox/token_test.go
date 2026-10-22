// Tests run tokens: a signed token verifies to its run and organization, and a tampered,
// expired, malformed or differently keyed one doesn't.
package sandbox

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunTokens(t *testing.T) {
	key := []byte(strings.Repeat("k", 32))
	run, org := uuid.New(), uuid.New()
	now := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	token := SignRunToken(key, run, org, now.Add(time.Hour))

	gotRun, gotOrg, err := VerifyRunToken(key, token, now)
	require.NoError(t, err)
	assert.Equal(t, run, gotRun)
	assert.Equal(t, org, gotOrg)

	payload, signature, _ := strings.Cut(token, ".")
	other := SignRunToken(key, uuid.New(), org, now.Add(time.Hour))
	otherPayload, _, _ := strings.Cut(other, ".")
	for name, tc := range map[string]struct {
		token string
		key   []byte
		at    time.Time
		want  string
	}{
		"expired":           {token, key, now.Add(2 * time.Hour), "run token expired at 2026-10-15T10:00:00Z"},
		"another key":       {token, []byte(strings.Repeat("x", 32)), now, "run token signature doesn't match"},
		"payload swapped":   {otherPayload + "." + signature, key, now, "run token signature doesn't match"},
		"signature altered": {payload + "." + signature[:len(signature)-2] + "AA", key, now, "run token signature doesn't match"},
		"no signature":      {payload, key, now, "run token is malformed"},
		"not base64":        {"!!!.!!!", key, now, "run token is malformed"},
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := VerifyRunToken(tc.key, tc.token, tc.at)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidRunToken)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}
