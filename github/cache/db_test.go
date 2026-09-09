package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_DisabledWhenNoDSN(t *testing.T) {
	resetSharedStore(t)
	t.Setenv(EnvDatabaseDSN, "")
	t.Setenv(EnvDatabaseDisable, "")
	t.Setenv(EnvDSN, "")
	t.Setenv(EnvDisable, "")

	s := Shared()
	require.NotNil(t, s)
	st := s.Status()
	assert.False(t, st.Enabled)
	assert.Equal(t, "postgres", st.Driver)
	assert.Empty(t, st.DSNSource)
	assert.Contains(t, st.Error, EnvDatabaseDSN)
	assert.Empty(t, st.Counts)
}

func TestStatus_DisabledWhenEnvOff(t *testing.T) {
	resetSharedStore(t)
	t.Setenv(EnvDatabaseDSN, "")
	t.Setenv(EnvDatabaseDisable, "")
	t.Setenv(EnvDSN, "postgres://user:secret@host/db")
	t.Setenv(EnvDisable, "off")

	s := Shared()
	st := s.Status()
	assert.False(t, st.Enabled)
	assert.Equal(t, EnvDSN, st.DSNSource)
	assert.Equal(t, "postgres://user:REDACTED@host/db", st.DSNMasked, "DSN surfaced even when disabled, but password redacted")
	assert.Contains(t, st.Error, EnvDisable)
}

func TestStatus_PrefersGeneralEnvironment(t *testing.T) {
	resetSharedStore(t)
	t.Setenv(EnvDatabaseDSN, "postgres://user:secret@host/db")
	t.Setenv(EnvDatabaseDisable, "off")
	t.Setenv(EnvDSN, "postgres://legacy/ignored")
	t.Setenv(EnvDisable, "")

	s := Shared()
	st := s.Status()
	assert.False(t, st.Enabled)
	assert.Equal(t, EnvDatabaseDSN, st.DSNSource)
	assert.Equal(t, "postgres://user:REDACTED@host/db", st.DSNMasked)
	assert.Equal(t, EnvDatabaseDisable+"=off", st.Error)
}
