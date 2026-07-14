package database

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearDatabaseEnvironment(t *testing.T) {
	t.Helper()
	databaseURL = ""
	t.Cleanup(func() { databaseURL = "" })
	t.Setenv(EnvDSN, "")
	t.Setenv(EnvDisable, "")
	t.Setenv(LegacyEnvDSN, "")
	t.Setenv(LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
}

func TestDatabaseURLFlagPrecedence(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv(EnvDSN, "postgres://preferred")
	t.Setenv(LegacyEnvDSN, "postgres://legacy")

	flags := pflag.NewFlagSet("gavel", pflag.ContinueOnError)
	BindDatabaseURLFlag(flags)
	require.NoError(t, flags.Parse([]string{"--db-url=postgres://flag/gavel"}))

	dsn, source, err := resolveDSN()
	require.NoError(t, err)
	assert.Equal(t, "postgres://flag/gavel", dsn)
	assert.Equal(t, "--db-url", source)
}

func TestOpenWithoutConfigurationIsDisabled(t *testing.T) {
	clearDatabaseEnvironment(t)
	db, err := Open(t.Context())
	require.NoError(t, err)
	require.True(t, db.Disabled())
	assert.Empty(t, db.DSN())
}

func TestEnvironmentPrecedence(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv(LegacyEnvDSN, "postgres://legacy")
	t.Setenv(EnvDSN, "postgres://preferred")

	dsn, source, err := resolveDSN()
	require.NoError(t, err)
	assert.Equal(t, "postgres://preferred", dsn)
	assert.Equal(t, EnvDSN, source)

	t.Setenv(EnvDisable, "on")
	t.Setenv(LegacyEnvDisable, "off")
	source, disabled := disabledByEnvironment()
	assert.False(t, disabled)
	assert.Equal(t, EnvDisable, source)

	require.NoError(t, os.Unsetenv(EnvDisable))
	source, disabled = disabledByEnvironment()
	assert.True(t, disabled)
	assert.Equal(t, LegacyEnvDisable, source)
}
