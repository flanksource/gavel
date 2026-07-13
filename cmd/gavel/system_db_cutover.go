package main

import (
	"context"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/internal/database"
)

type SystemDBCutoverOptions struct {
	ArtifactDir string `flag:"artifact-dir" help:"Parent directory for the private immutable rollback/validation generation (default: ~/.gavel/migrations/captain-sessions)"`
}

func (SystemDBCutoverOptions) Help() string {
	return `Explicitly archive and backfill Captain's retired path-keyed session
cache, then apply Captain and Gavel's authoritative HCL schemas.

Normal Gavel startup deliberately refuses the incompatible legacy table shape.
Run this command once after stopping older Gavel/Captain processes. The original
session and prompt rows remain in versioned Captain archive tables, and the
returned report includes row counts and checksums for validation and rollback.
The command is idempotent and validates the archived/native rows on retries.`
}

var cutoverLegacyCaptainSessions = database.CutoverLegacyCaptainSessionsWithOptions

func init() {
	clicky.AddNamedCommand("db-cutover", systemCmd, SystemDBCutoverOptions{}, runSystemDBCutover)
}

// systemDBCutoverReport keeps the database package free of presentation
// dependencies while preserving its typed JSON fields for Clicky's formatters.
type systemDBCutoverReport database.LegacySessionCutoverResult

func runSystemDBCutover(options SystemDBCutoverOptions) (any, error) {
	result, err := cutoverLegacyCaptainSessions(context.Background(), database.LegacySessionCutoverOptions{
		ArtifactDir: options.ArtifactDir,
	})
	if result == nil {
		return nil, err
	}
	return systemDBCutoverReport(*result), err
}

func (r systemDBCutoverReport) Pretty() api.Text {
	t := api.Text{}
	complete := r.CaptainSchemaApplied && r.GavelSchemaApplied
	if !complete {
		t = t.Add(icons.Warning).Space().Append("Database cutover incomplete", "warning").NewLine()
	} else if r.Captain == nil {
		t = t.Add(icons.Success).Space()
		t = t.Append("Database schemas applied; no legacy Captain session cache found", "text-green-600").NewLine()
	} else {
		t = t.Add(icons.Success).Space()
		t = t.Append("Captain legacy session cache cutover complete", "text-green-600").NewLine()
	}
	t = t.Append(kv("source", r.DSNSource)).NewLine()
	t = t.Append(kv("captain schema", schemaAppliedLabel(r.CaptainSchemaApplied))).NewLine()
	t = t.Append(kv("gavel schema", schemaAppliedLabel(r.GavelSchemaApplied))).NewLine()
	if r.Captain == nil {
		return t
	}

	t = t.Append(kv("session archive", r.Captain.LegacySessionsTable)).NewLine()
	if r.Captain.LegacyPromptsTable != nil {
		t = t.Append(kv("prompt archive", *r.Captain.LegacyPromptsTable)).NewLine()
	}
	if r.ArtifactDirectory != "" {
		t = t.Append(kv("artifact", r.ArtifactDirectory)).NewLine()
	}
	t = t.Append(kv("sessions", "")).
		Append(r.Captain.LegacySessionRows).Append(" archived / ").
		Append(r.Captain.ImportedSessionRows).Append(" imported").NewLine()
	t = t.Append(kv("prompt runs", "")).
		Append(r.Captain.LegacyPromptRows).Append(" archived / ").
		Append(r.Captain.ImportedPromptRunRows).Append(" imported").NewLine()
	t = t.Append(kv("legacy checksum", shortHash(r.Captain.LegacySessionsChecksum))).NewLine()
	t = t.Append(kv("native checksum", shortHash(r.Captain.NativeSessionsChecksum))).NewLine()
	return t
}

func schemaAppliedLabel(applied bool) string {
	if applied {
		return "applied"
	}
	return "not applied"
}
