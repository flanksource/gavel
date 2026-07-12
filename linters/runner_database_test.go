package linters

import (
	"testing"

	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/require"
)

func TestNewRunnerWithoutDatabaseConfigurationDisablesCaches(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(database.EnvDSN, "")
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")

	runner, err := NewRunnerWithOptions(&models.Config{}, t.TempDir(), RunnerOptions{})
	require.NoError(t, err)
	require.Nil(t, runner.violationCache)
	require.Nil(t, runner.linterStats)
	require.NoError(t, runner.Close())
}
