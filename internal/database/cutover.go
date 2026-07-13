package database

import (
	"context"
	"fmt"
	"strings"

	captainmigrations "github.com/flanksource/captain/migrations"
	captaindb "github.com/flanksource/captain/pkg/database"
)

const (
	CutoverModeNoLegacy = "no-legacy-session-cache"
	CutoverModeLegacy   = "legacy-session-cache-cutover"
)

// LegacySessionCutoverResult reports the complete explicit Captain-then-Gavel
// migration lifecycle. The configured DSN itself is deliberately omitted.
type LegacySessionCutoverResult struct {
	Mode                 string                                        `json:"mode"`
	DSNSource            string                                        `json:"dsnSource"`
	ArtifactDirectory    string                                        `json:"artifactDirectory,omitempty"`
	CaptainSchemaApplied bool                                          `json:"captainSchemaApplied"`
	GavelSchemaApplied   bool                                          `json:"gavelSchemaApplied"`
	Captain              *captainmigrations.LegacySessionCutoverReport `json:"captain,omitempty"`
}

// LegacySessionCutoverOptions controls the private filesystem evidence written
// for a real legacy cutover. ArtifactDir is a parent root; an empty value uses
// Gavel's private default migration directory.
type LegacySessionCutoverOptions struct {
	ArtifactDir string
}

type legacySessionCutoverDependencies struct {
	withLock        func(context.Context, string, func() error) error
	migrateCaptain  func(context.Context, string) (*captainmigrations.LegacySessionCutoverReport, error)
	captureArtifact func(context.Context, string, string, *captainmigrations.LegacySessionCutoverReport) (*CutoverArtifactResult, error)
	migrateGavel    func(context.Context, string) error
}

var defaultLegacySessionCutoverDependencies = legacySessionCutoverDependencies{
	withLock:        withMigrationAdvisoryLock,
	migrateCaptain:  captaindb.MigrateWithLegacySessionCutover,
	captureArtifact: CaptureLegacyCaptainCutoverArtifact,
	migrateGavel:    applyGavelSchema,
}

// CutoverLegacyCaptainSessions is the explicit opt-in migration for databases
// containing Captain's retired path-keyed session summary cache. Unlike Open,
// it asks Captain to archive and backfill that cache before applying Captain's
// authoritative schema, then applies Gavel's HCL bundle under the same outer
// Gavel advisory lock. Normal Open remains fail-closed for the legacy shape.
func CutoverLegacyCaptainSessions(ctx context.Context) (*LegacySessionCutoverResult, error) {
	return CutoverLegacyCaptainSessionsWithOptions(ctx, LegacySessionCutoverOptions{})
}

// CutoverLegacyCaptainSessionsWithOptions is CutoverLegacyCaptainSessions with
// an optional parent directory for the immutable rollback/validation artifact.
func CutoverLegacyCaptainSessionsWithOptions(ctx context.Context, options LegacySessionCutoverOptions) (*LegacySessionCutoverResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if source, disabled := disabledByEnvironment(); disabled {
		return nil, fmt.Errorf("Captain session cutover requires PostgreSQL: %w (%s=off)", ErrUnavailable, source)
	}
	dsn, source, err := resolveDSN()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("Captain session cutover requires PostgreSQL: %w; set %s or configure embedded PostgreSQL with gavel system install --embedded", ErrUnavailable, EnvDSN)
	}
	return runLegacySessionCutover(ctx, dsn, source, options.ArtifactDir, defaultLegacySessionCutoverDependencies)
}

func runLegacySessionCutover(
	ctx context.Context,
	dsn string,
	source string,
	artifactRoot string,
	deps legacySessionCutoverDependencies,
) (*LegacySessionCutoverResult, error) {
	result := &LegacySessionCutoverResult{DSNSource: source}
	err := deps.withLock(ctx, dsn, func() error {
		report, err := deps.migrateCaptain(ctx, dsn)
		if err != nil {
			return fmt.Errorf("cut over legacy Captain session cache: %w", err)
		}
		result.Captain = report
		result.CaptainSchemaApplied = true
		result.Mode = CutoverModeNoLegacy
		if report != nil {
			result.Mode = CutoverModeLegacy
			artifact, err := deps.captureArtifact(ctx, dsn, artifactRoot, report)
			if err != nil {
				return fmt.Errorf("capture Captain legacy session cutover artifact: %w", err)
			}
			if artifact == nil {
				return fmt.Errorf("capture Captain legacy session cutover artifact: empty result")
			}
			result.ArtifactDirectory = artifact.Directory
		}

		if err := deps.migrateGavel(ctx, dsn); err != nil {
			return err
		}
		result.GavelSchemaApplied = true
		return nil
	})
	return result, err
}
