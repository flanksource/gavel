package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDBConfig_MissingFileReturnsZero(t *testing.T) {
	withTempHome(t)
	cfg, err := LoadDBConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.Mode)
	assert.Empty(t, cfg.DSN)
}

func TestSaveLoadDBConfig_Roundtrip(t *testing.T) {
	withTempHome(t)
	in := DBConfig{Mode: DBModeDSN, DSN: "postgres://u:p@localhost:5432/gavel"}
	require.NoError(t, SaveDBConfig(in))
	out, err := LoadDBConfig()
	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestSaveDBConfig_WritesUserOnlyPerms(t *testing.T) {
	withTempHome(t)
	require.NoError(t, SaveDBConfig(DBConfig{Mode: DBModeEmbedded}))
	path, err := DBConfigPath()
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	// DSNs can contain credentials; 0600 prevents other users from reading.
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestSaveDBConfig_RejectsInvalidMode(t *testing.T) {
	withTempHome(t)
	assert.Error(t, SaveDBConfig(DBConfig{Mode: "sqlite"}))
	assert.Error(t, SaveDBConfig(DBConfig{Mode: DBModeDSN, DSN: ""}))
}

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
		assert.Equal(t, want, MaskDSN(in), "input=%q", in)
	}
}
