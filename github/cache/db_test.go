package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaskDSN(t *testing.T) {
	tests := map[string]string{
		"":                                    "",
		"postgres://api.github.com/db":        "postgres://api.github.com/db",
		"postgres://user:secret@host:5432/db": "postgres://user:REDACTED@host:5432/db",
		"postgres://user@host/db":             "postgres://user@host/db",
		"host=localhost user=u password=s dbname=d": "host=localhost user=u password=REDACTED dbname=d",
		"password=topsecret":                        "password=REDACTED",
	}
	for in, want := range tests {
		assert.Equal(t, want, maskDSN(in), "input=%q", in)
	}
}

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
