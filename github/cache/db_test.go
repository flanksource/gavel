package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_DisabledWhenNoDSN(t *testing.T) {
	resetSharedStore(t)
	t.Setenv(EnvDSN, "")
	t.Setenv(EnvDisable, "")

	s := Shared()
	require.NotNil(t, s)
	st := s.Status()
	assert.False(t, st.Enabled)
	assert.Equal(t, "postgres", st.Driver)
	assert.Empty(t, st.DSNSource)
	assert.Contains(t, st.Error, EnvDSN)
	assert.Empty(t, st.Counts)
}

func TestStatus_DisabledWhenEnvOff(t *testing.T) {
	resetSharedStore(t)
	t.Setenv(EnvDSN, "postgres://user:secret@host/db")
	t.Setenv(EnvDisable, "off")

	s := Shared()
	st := s.Status()
	assert.False(t, st.Enabled)
	assert.Equal(t, EnvDSN, st.DSNSource)
	assert.Equal(t, "postgres://user:REDACTED@host/db", st.DSNMasked, "DSN surfaced even when disabled, but password redacted")
	assert.Contains(t, st.Error, EnvDisable)
}
