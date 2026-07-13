package database_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/migrations"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplicitLegacyCaptainCutoverThenGavelHCLIsRepeatable(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres cutover tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: "gavel_legacy_captain_cutover",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	legacy, err := commonsdb.NewGorm(dsn, commonsdb.DefaultGormConfig())
	require.NoError(t, err)
	legacySQL, err := legacy.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, legacySQL.Close()) })

	require.NoError(t, legacy.Exec(`CREATE TABLE captain_sessions (
		path text PRIMARY KEY,
		id text NOT NULL,
		parent_id text,
		source text NOT NULL,
		mod_unix bigint,
		size bigint,
		project text,
		cwd text,
		model text,
		title text,
		initial_prompt text,
		git jsonb,
		provider jsonb,
		started_at timestamptz,
		ended_at timestamptz,
		updated_at timestamptz
	)`).Error)
	require.NoError(t, legacy.Exec(`CREATE TABLE captain_session_prompts (
		session_id text PRIMARY KEY,
		run_id text,
		model text,
		backend text,
		realized jsonb,
		created_at timestamptz
	)`).Error)
	require.NoError(t, legacy.Exec(`INSERT INTO captain_sessions (
		path, id, source, mod_unix, size, project, cwd, model, title,
		initial_prompt, git, provider, started_at, ended_at, updated_at
	) VALUES (
		'/tmp/legacy-session.jsonl', 'legacy-provider-session', 'codex', 42, 100,
		'gavel', '/work/gavel', 'gpt-legacy', 'Legacy session', 'Implement the cutover',
		'{"branch":"main"}'::jsonb,
		'{"name":"openai","version":"legacy","backend":"codex"}'::jsonb,
		now() - interval '2 minutes', now() - interval '1 minute', now()
	)`).Error)
	require.NoError(t, legacy.Exec(`INSERT INTO captain_session_prompts (
		session_id, run_id, model, backend, realized, created_at
	) VALUES (
		'legacy-provider-session', 'legacy-prompt-run', 'gpt-legacy', 'codex',
		'{"input":{"prompt":{"user":"Implement the cutover"}}}'::jsonb, now()
	)`).Error)
	require.NoError(t, legacy.Exec(`CREATE TABLE grite_issue_caches (
		repo text NOT NULL, issue_id text NOT NULL, title text,
		PRIMARY KEY (repo, issue_id)
	)`).Error)
	require.NoError(t, legacy.Exec(`INSERT INTO grite_issue_caches (repo, issue_id, title)
		VALUES ('flanksource/gavel', 'e2a3b8c2d0f7c9a98b400dc78e8a94a5', 'preserve Grite cache')`).Error)
	require.NoError(t, legacy.Exec(`CREATE TABLE migration_unmanaged (id integer PRIMARY KEY, value text)`).Error)
	require.NoError(t, legacy.Exec(`INSERT INTO migration_unmanaged VALUES (1, 'preserve me')`).Error)

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())

	_, err = database.Open(t.Context())
	require.Error(t, err)
	assert.True(t, errors.Is(err, migrations.ErrLegacySessionSchema), "normal Open must remain fail-closed: %v", err)

	first, err := database.CutoverLegacyCaptainSessions(t.Context())
	require.NoError(t, err)
	require.NotNil(t, first)
	require.NotNil(t, first.Captain)
	assert.Equal(t, database.CutoverModeLegacy, first.Mode)
	assert.True(t, first.CaptainSchemaApplied)
	assert.True(t, first.GavelSchemaApplied)
	assert.NotEmpty(t, first.ArtifactDirectory)
	_, err = database.VerifyCutoverArtifact(first.ArtifactDirectory)
	require.NoError(t, err)
	archiveSchema, err := os.ReadFile(filepath.Join(first.ArtifactDirectory, "legacy-schema.json"))
	require.NoError(t, err)
	assert.Contains(t, string(archiveSchema), `"dataType": "text"`)
	assert.Contains(t, string(archiveSchema), `"definition": "PRIMARY KEY (path)"`)
	assert.EqualValues(t, 1, first.Captain.LegacySessionRows)
	assert.EqualValues(t, 1, first.Captain.LegacyPromptRows)
	assert.EqualValues(t, 1, first.Captain.ImportedSessionRows)
	assert.EqualValues(t, 1, first.Captain.ImportedPromptRunRows)

	for _, table := range []string{
		"captain_sessions", "captain_prompt_runs", "captain_sessions_legacy_v1",
		"captain_session_prompts_legacy_v1", "captain_legacy_session_cutovers",
		"todo_workspaces", "grite_issue_caches", "migration_unmanaged",
	} {
		assert.True(t, legacy.Migrator().HasTable(table), "%s should exist after cutover", table)
	}
	var legacyTitle, unmanagedValue, promptMarkdown string
	require.NoError(t, legacy.Raw(`SELECT title FROM grite_issue_caches
		WHERE repo = 'flanksource/gavel' AND issue_id = 'e2a3b8c2d0f7c9a98b400dc78e8a94a5'`).Scan(&legacyTitle).Error)
	require.NoError(t, legacy.Raw(`SELECT value FROM migration_unmanaged WHERE id = 1`).Scan(&unmanagedValue).Error)
	require.NoError(t, legacy.Raw(`SELECT prompt_markdown FROM captain_prompt_runs
		WHERE origin = 'legacy-session-cache'`).Scan(&promptMarkdown).Error)
	assert.Equal(t, "preserve Grite cache", legacyTitle)
	assert.Equal(t, "preserve me", unmanagedValue)
	assert.Equal(t, "Implement the cutover", promptMarkdown)

	second, err := database.CutoverLegacyCaptainSessions(t.Context())
	require.NoError(t, err)
	require.NotNil(t, second.Captain)
	assert.Equal(t, first.ArtifactDirectory, second.ArtifactDirectory)
	assert.Equal(t, first.Captain.LegacySessionsChecksum, second.Captain.LegacySessionsChecksum)
	assert.Equal(t, first.Captain.LegacyPromptsChecksum, second.Captain.LegacyPromptsChecksum)
	assert.Equal(t, first.Captain.NativeSessionsChecksum, second.Captain.NativeSessionsChecksum)
	assert.Equal(t, first.Captain.NativePromptRunsChecksum, second.Captain.NativePromptRunsChecksum)

	opened, err := database.Open(t.Context())
	require.NoError(t, err)
	require.False(t, opened.Disabled())
	require.NoError(t, opened.Close())
}
