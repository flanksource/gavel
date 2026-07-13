package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	captainmigrations "github.com/flanksource/captain/migrations"
	"github.com/flanksource/gavel/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSystemDBCutoverReturnsTypedStructuredReport(t *testing.T) {
	original := cutoverLegacyCaptainSessions
	t.Cleanup(func() { cutoverLegacyCaptainSessions = original })
	promptArchive := "captain_session_prompts_legacy_v1"
	cutoverLegacyCaptainSessions = func(_ context.Context, options database.LegacySessionCutoverOptions) (*database.LegacySessionCutoverResult, error) {
		assert.Equal(t, "/custom/artifacts", options.ArtifactDir)
		return &database.LegacySessionCutoverResult{
			Mode: database.CutoverModeLegacy, DSNSource: database.EnvDSN,
			ArtifactDirectory:    "/custom/artifacts/generation",
			CaptainSchemaApplied: true, GavelSchemaApplied: true,
			Captain: &captainmigrations.LegacySessionCutoverReport{
				CutoverKey: "captain-session-cache-v1", LegacySessionsTable: "captain_sessions_legacy_v1",
				LegacyPromptsTable: &promptArchive, LegacySessionRows: 3, LegacyPromptRows: 2,
				ImportedSessionRows: 3, ImportedPromptRunRows: 2,
				LegacySessionsChecksum: "1234567890abcdef", NativeSessionsChecksum: "abcdef1234567890",
			},
		}, nil
	}

	value, err := runSystemDBCutover(SystemDBCutoverOptions{ArtifactDir: "/custom/artifacts"})
	require.NoError(t, err)
	report, ok := value.(systemDBCutoverReport)
	require.True(t, ok)
	pretty := report.Pretty().String()
	assert.Contains(t, pretty, "Captain legacy session cache cutover complete")
	assert.Contains(t, pretty, "3 archived / 3 imported")
	assert.Contains(t, pretty, "captain_sessions_legacy_v1")
	assert.Contains(t, pretty, "/custom/artifacts/generation")

	raw, err := json.Marshal(report)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"mode":"legacy-session-cache-cutover",
		"dsnSource":"GAVEL_DB_DSN",
		"artifactDirectory":"/custom/artifacts/generation",
		"captainSchemaApplied":true,
		"gavelSchemaApplied":true,
		"captain":{
			"cutoverKey":"captain-session-cache-v1",
			"legacySessionsTable":"captain_sessions_legacy_v1",
			"legacyPromptsTable":"captain_session_prompts_legacy_v1",
			"legacySessionRows":3,
			"legacyPromptRows":2,
			"importedSessionRows":3,
			"importedPromptRunRows":2,
			"legacySessionsChecksum":"1234567890abcdef",
			"nativeSessionsChecksum":"abcdef1234567890",
			"details":null,
			"startedAt":"0001-01-01T00:00:00Z",
			"completedAt":"0001-01-01T00:00:00Z",
			"updatedAt":"0001-01-01T00:00:00Z"
		}
	}`, string(raw))
}

func TestRunSystemDBCutoverReturnsPartialReportWithMigrationError(t *testing.T) {
	original := cutoverLegacyCaptainSessions
	t.Cleanup(func() { cutoverLegacyCaptainSessions = original })
	wantErr := errors.New("gavel migration failed")
	cutoverLegacyCaptainSessions = func(context.Context, database.LegacySessionCutoverOptions) (*database.LegacySessionCutoverResult, error) {
		return &database.LegacySessionCutoverResult{
			Mode: database.CutoverModeLegacy, CaptainSchemaApplied: true,
			Captain: &captainmigrations.LegacySessionCutoverReport{CutoverKey: "captain-session-cache-v1"},
		}, wantErr
	}

	value, err := runSystemDBCutover(SystemDBCutoverOptions{})
	require.ErrorIs(t, err, wantErr)
	report, ok := value.(systemDBCutoverReport)
	require.True(t, ok)
	assert.True(t, report.CaptainSchemaApplied)
	assert.False(t, report.GavelSchemaApplied)
	assert.Contains(t, report.Pretty().String(), "Database cutover incomplete")
}

func TestSystemDBCutoverCommandIsRegistered(t *testing.T) {
	command, _, err := systemCmd.Find([]string{"db-cutover"})
	require.NoError(t, err)
	assert.Equal(t, "db-cutover", command.Name())
}
