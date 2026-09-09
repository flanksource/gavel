package verify

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestConfigTimeouts(t *testing.T) {
	global, test, lint, err := TestConfig{Timeout: "45m", TestTimeout: "20m", LintTimeout: "8m"}.Timeouts()
	require.NoError(t, err)
	assert.Equal(t, 45*time.Minute, global)
	assert.Equal(t, 20*time.Minute, test)
	assert.Equal(t, 8*time.Minute, lint)
}

// An unset field reads as zero so the caller keeps its own default rather than
// collapsing the deadline to nothing.
func TestTestConfigTimeoutsTreatsEmptyAsUnset(t *testing.T) {
	global, test, lint, err := TestConfig{TestTimeout: "20m"}.Timeouts()
	require.NoError(t, err)
	assert.Zero(t, global)
	assert.Equal(t, 20*time.Minute, test)
	assert.Zero(t, lint)
}

func TestTestConfigTimeoutsRejectsMalformedValues(t *testing.T) {
	for _, tc := range []struct{ name, value, wants string }{
		{name: "not a duration", value: "twenty-minutes", wants: "is not a duration"},
		{name: "bare number", value: "20", wants: "is not a duration"},
		{name: "zero", value: "0s", wants: "must be positive"},
		{name: "negative", value: "-5m", wants: "must be positive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := TestConfig{TestTimeout: tc.value}.Timeouts()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "test.testTimeout")
			assert.Contains(t, err.Error(), tc.wants)
		})
	}
}

// The field name in the error has to be the one the reader would edit, so a
// malformed global timeout must not be reported as the per-package one.
func TestTestConfigTimeoutsNamesTheOffendingField(t *testing.T) {
	_, _, _, err := TestConfig{Timeout: "soon"}.Timeouts()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test.timeout")
	assert.NotContains(t, err.Error(), "testTimeout")
}
