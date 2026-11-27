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
		"signature altered": {payload + "." + replaceAt(signature, 0), key, now, "run token signature doesn't match"},
		// The last of a signature's 43 characters holds 4 of its bits and 2 that must be zero.
		"signature padding set": {payload + "." + setPaddingBit(signature), key, now, "run token is malformed"},
		"no signature":          {payload, key, now, "run token is malformed"},
		"not base64":            {"!!!.!!!", key, now, "run token is malformed"},
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := VerifyRunToken(tc.key, tc.token, tc.at)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidRunToken)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

const base64URLAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// replaceAt is encoded with its i'th character changed to another, so it decodes to other bytes.
func replaceAt(encoded string, i int) string {
	replacement := byte('A')
	if encoded[i] == replacement {
		replacement = 'B'
	}
	return encoded[:i] + string(replacement) + encoded[i+1:]
}

// setPaddingBit is encoded with the lowest bit of its last character set: an encoding of a 32-byte
// signature that a decoder ignoring padding bits reads as the same bytes.
func setPaddingBit(encoded string) string {
	last := strings.IndexByte(base64URLAlphabet, encoded[len(encoded)-1])
	return encoded[:len(encoded)-1] + string(base64URLAlphabet[last|1])
}
