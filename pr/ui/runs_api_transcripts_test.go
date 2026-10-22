// Tests the run API's transcripts endpoint: an agent run's job uploads its agent's transcript,
// and serve stores and ingests it, owned by the run's organization and user.
package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestRunsAPIIngestsAnAgentRunsTranscripts(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runs_api_transcripts"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org, user := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, user).Error)
	agentRun, testRun := uuid.New(), uuid.New()
	for id, kind := range map[uuid.UUID]string{agentRun: "agent", testRun: "test"} {
		spec, err := json.Marshal(sandbox.RunSpec{Kind: kind})
		require.NoError(t, err)
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, user_id, kind, status, backend, spec) VALUES (?, ?, ?, ?, 'running', 'memory', ?)`, id, org, user, kind, string(spec)).Error)
	}

	key := []byte(strings.Repeat("k", 32))
	dataDir := t.TempDir()
	server := httptest.NewServer((&RunsAPI{DB: db, TokenKey: key, Sources: &sandbox.Sources{DataDir: dataDir}}).Handler())
	t.Cleanup(server.Close)
	post := func(run uuid.UUID, query, body string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/runs/"+run.String()+"/transcripts?"+query, bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+sandbox.SignRunToken(key, run, org, time.Now().Add(time.Hour)))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}

	name := "-workspace-src/" + sandboxSessionID + ".jsonl"
	status, body := post(agentRun, "source=claude&name="+name, sandboxTranscript)
	require.Equal(t, http.StatusAccepted, status, body)
	stored, err := os.ReadFile(filepath.Join(dataDir, "transcripts", org.String(), agentRun.String(), "claude", name))
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

	for query, want := range map[string]string{
		"source=gemini&name=a.jsonl":            `transcript source "gemini" is not claude or codex`,
		"source=claude&name=../../escape.jsonl": `transcript name "../../escape.jsonl" must be a relative .jsonl path inside the run's transcripts`,
		"source=claude&name=/etc/passwd":        `transcript name "/etc/passwd" must be a relative .jsonl path inside the run's transcripts`,
		"source=claude&name=sub/../a.jsonl":     `transcript name "sub/../a.jsonl" must be a relative .jsonl path inside the run's transcripts`,
		"source=claude&name=notes.txt":          `transcript name "notes.txt" must be a relative .jsonl path inside the run's transcripts`,
	} {
		status, body := post(agentRun, query, sandboxTranscript)
		assert.Equal(t, http.StatusBadRequest, status, query)
		var refused struct{ Error string }
		require.NoError(t, json.Unmarshal([]byte(body), &refused), body)
		assert.Equal(t, want, refused.Error, query)
	}
	status, body = post(testRun, "source=claude&name="+name, sandboxTranscript)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "only agent runs upload transcripts")
}
