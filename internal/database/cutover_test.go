package database

import (
	"context"
	"errors"
	"testing"

	captainmigrations "github.com/flanksource/captain/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunLegacySessionCutoverOrdersCaptainBeforeGavelInsideOuterLock(t *testing.T) {
	report := &captainmigrations.LegacySessionCutoverReport{CutoverKey: "captain-session-cache-v1"}
	var calls []string
	result, err := runLegacySessionCutover(t.Context(), "postgres://configured", EnvDSN, "/artifacts", legacySessionCutoverDependencies{
		withLock: func(_ context.Context, dsn string, callback func() error) error {
			calls = append(calls, "lock:"+dsn)
			err := callback()
			calls = append(calls, "unlock")
			return err
		},
		migrateCaptain: func(_ context.Context, dsn string) (*captainmigrations.LegacySessionCutoverReport, error) {
			calls = append(calls, "captain:"+dsn)
			return report, nil
		},
		captureArtifact: func(_ context.Context, dsn, root string, actual *captainmigrations.LegacySessionCutoverReport) (*CutoverArtifactResult, error) {
			calls = append(calls, "artifact:"+dsn+":"+root)
			assert.Same(t, report, actual)
			return &CutoverArtifactResult{Directory: "/artifacts/generation"}, nil
		},
		migrateGavel: func(_ context.Context, dsn string) error {
			calls = append(calls, "gavel:"+dsn)
			return nil
		},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"lock:postgres://configured",
		"captain:postgres://configured",
		"artifact:postgres://configured:/artifacts",
		"gavel:postgres://configured",
		"unlock",
	}, calls)
	assert.Equal(t, CutoverModeLegacy, result.Mode)
	assert.Equal(t, EnvDSN, result.DSNSource)
	assert.Equal(t, "/artifacts/generation", result.ArtifactDirectory)
	assert.True(t, result.CaptainSchemaApplied)
	assert.True(t, result.GavelSchemaApplied)
	assert.Same(t, report, result.Captain)
}

func TestRunLegacySessionCutoverCleanDatabaseStillAppliesBothSchemas(t *testing.T) {
	artifactCalled := false
	result, err := runLegacySessionCutover(t.Context(), "postgres://configured", "db.json (embedded)", "", legacySessionCutoverDependencies{
		withLock: func(_ context.Context, _ string, callback func() error) error { return callback() },
		migrateCaptain: func(context.Context, string) (*captainmigrations.LegacySessionCutoverReport, error) {
			return nil, nil
		},
		captureArtifact: func(context.Context, string, string, *captainmigrations.LegacySessionCutoverReport) (*CutoverArtifactResult, error) {
			artifactCalled = true
			return nil, nil
		},
		migrateGavel: func(context.Context, string) error { return nil },
	})
	require.NoError(t, err)
	assert.Equal(t, CutoverModeNoLegacy, result.Mode)
	assert.Nil(t, result.Captain)
	assert.False(t, artifactCalled)
	assert.True(t, result.CaptainSchemaApplied)
	assert.True(t, result.GavelSchemaApplied)
}

func TestRunLegacySessionCutoverStopsBeforeGavelWhenCaptainFails(t *testing.T) {
	wantErr := errors.New("captain failed")
	gavelCalled := false
	result, err := runLegacySessionCutover(t.Context(), "postgres://configured", EnvDSN, "", legacySessionCutoverDependencies{
		withLock: func(_ context.Context, _ string, callback func() error) error { return callback() },
		migrateCaptain: func(context.Context, string) (*captainmigrations.LegacySessionCutoverReport, error) {
			return nil, wantErr
		},
		captureArtifact: func(context.Context, string, string, *captainmigrations.LegacySessionCutoverReport) (*CutoverArtifactResult, error) {
			t.Fatal("artifact capture must not run after Captain failure")
			return nil, nil
		},
		migrateGavel: func(context.Context, string) error {
			gavelCalled = true
			return nil
		},
	})
	require.ErrorIs(t, err, wantErr)
	assert.False(t, gavelCalled)
	assert.False(t, result.CaptainSchemaApplied)
	assert.False(t, result.GavelSchemaApplied)
}

func TestRunLegacySessionCutoverReturnsPartialReportWhenGavelFails(t *testing.T) {
	wantErr := errors.New("gavel failed")
	report := &captainmigrations.LegacySessionCutoverReport{CutoverKey: "captain-session-cache-v1"}
	result, err := runLegacySessionCutover(t.Context(), "postgres://configured", EnvDSN, "", legacySessionCutoverDependencies{
		withLock: func(_ context.Context, _ string, callback func() error) error { return callback() },
		migrateCaptain: func(context.Context, string) (*captainmigrations.LegacySessionCutoverReport, error) {
			return report, nil
		},
		captureArtifact: func(context.Context, string, string, *captainmigrations.LegacySessionCutoverReport) (*CutoverArtifactResult, error) {
			return &CutoverArtifactResult{Directory: "/artifacts/generation"}, nil
		},
		migrateGavel: func(context.Context, string) error { return wantErr },
	})
	require.ErrorIs(t, err, wantErr)
	assert.True(t, result.CaptainSchemaApplied)
	assert.False(t, result.GavelSchemaApplied)
	assert.Same(t, report, result.Captain)
}

func TestRunLegacySessionCutoverStopsBeforeGavelWhenArtifactFails(t *testing.T) {
	wantErr := errors.New("artifact failed")
	gavelCalled := false
	result, err := runLegacySessionCutover(t.Context(), "postgres://configured", EnvDSN, "", legacySessionCutoverDependencies{
		withLock: func(_ context.Context, _ string, callback func() error) error { return callback() },
		migrateCaptain: func(context.Context, string) (*captainmigrations.LegacySessionCutoverReport, error) {
			return &captainmigrations.LegacySessionCutoverReport{CutoverKey: "captain-session-cache-v1"}, nil
		},
		captureArtifact: func(context.Context, string, string, *captainmigrations.LegacySessionCutoverReport) (*CutoverArtifactResult, error) {
			return nil, wantErr
		},
		migrateGavel: func(context.Context, string) error {
			gavelCalled = true
			return nil
		},
	})
	require.ErrorIs(t, err, wantErr)
	assert.True(t, result.CaptainSchemaApplied)
	assert.False(t, result.GavelSchemaApplied)
	assert.False(t, gavelCalled)
}

func TestCutoverLegacyCaptainSessionsRequiresConfiguredDatabase(t *testing.T) {
	clearDatabaseEnvironment(t)
	result, err := CutoverLegacyCaptainSessions(t.Context())
	require.ErrorIs(t, err, ErrUnavailable)
	assert.Nil(t, result)
}
