// Tests how serve keeps an agent run's transcript: stored under its data directory, ingested,
// and owned by the run's organization and user.
package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sandboxSessionID = "0195c1de-4ab8-7000-8000-0123456789ab"

const sandboxTranscript = `{"type":"user","uuid":"u-1","sessionId":"` + sandboxSessionID + `","timestamp":"2026-07-05T10:00:00Z","cwd":"/workspace/src","message":{"role":"user","content":"fix the build"}}
{"type":"assistant","uuid":"a-1","sessionId":"` + sandboxSessionID + `","timestamp":"2026-07-05T10:00:05Z","message":{"role":"assistant","model":"claude-sonnet-5","usage":{"input_tokens":100,"output_tokens":20},"content":[{"type":"text","text":"fixed it"}]}}
`

func TestRunTranscriptsKeepsAndIngestsARunsTranscript(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_run_transcripts"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org, user := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, user).Error)
	run := sandbox.RunRecord{ID: uuid.New(), Org: org, User: &user, Spec: sandbox.RunSpec{Kind: "agent"}}
	dataDir := t.TempDir()
	transcripts := RunTranscripts{DB: db, DataDir: dataDir}

	name := "-workspace-src/" + sandboxSessionID + ".jsonl"
	require.NoError(t, transcripts.Keep(t.Context(), run, "claude", name, strings.NewReader(sandboxTranscript)))
	stored, err := os.ReadFile(filepath.Join(dataDir, "transcripts", org.String(), run.ID.String(), "claude", name))
	require.NoError(t, err)
	assert.Equal(t, sandboxTranscript, string(stored))

	var session struct {
		ID           uuid.UUID
		MessageCount int
		OrgID        uuid.UUID
		UserID       uuid.UUID
	}
	require.NoError(t, db.Raw(`SELECT s.id, (SELECT count(*) FROM captain_messages m WHERE m.session_id = s.id) AS message_count, o.org_id, o.user_id
		FROM captain_sessions s JOIN session_owners o ON o.session_id = s.id WHERE s.provider_session_id = ?`, sandboxSessionID).Scan(&session).Error)
	assert.Equal(t, 2, session.MessageCount, "the transcript was ingested")
	assert.Equal(t, org, session.OrgID)
	assert.Equal(t, user, session.UserID)

	for _, refused := range []struct{ source, name, want string }{
		{"gemini", "a.jsonl", `transcript source "gemini" is not claude or codex`},
		{"claude", "../../escape.jsonl", `transcript name "../../escape.jsonl" must be a relative .jsonl path inside the run's transcripts`},
		{"claude", "/etc/passwd", `transcript name "/etc/passwd" must be a relative .jsonl path inside the run's transcripts`},
		{"claude", "sub/../a.jsonl", `transcript name "sub/../a.jsonl" must be a relative .jsonl path inside the run's transcripts`},
		{"claude", "notes.txt", `transcript name "notes.txt" must be a relative .jsonl path inside the run's transcripts`},
	} {
		err := transcripts.Keep(t.Context(), run, refused.source, refused.name, strings.NewReader(sandboxTranscript))
		require.ErrorAs(t, err, new(*TranscriptRefusedError), refused.name)
		assert.EqualError(t, err, refused.want)
	}
}
