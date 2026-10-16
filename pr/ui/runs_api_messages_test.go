// Tests the run API's messages endpoint: an agent run's sandbox job sends its messages there,
// and serve hands each to the host waiting for the run.
package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/todos/remote"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunsAPIDeliversAgentRunMessages(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runs_api_messages"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	agentRun, unwatchedRun, testRun := uuid.New(), uuid.New(), uuid.New()
	for id, kind := range map[uuid.UUID]string{agentRun: "agent", unwatchedRun: "agent", testRun: "test"} {
		spec, err := json.Marshal(sandbox.RunSpec{Kind: kind})
		require.NoError(t, err)
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, spec) VALUES (?, ?, ?, 'running', 'memory', ?)`, id, org, kind, string(spec)).Error)
	}

	key := []byte(strings.Repeat("k", 32))
	relay := &remote.Relay{}
	server := httptest.NewServer((&RunsAPI{DB: db, TokenKey: key, Relay: relay}).Handler())
	t.Cleanup(server.Close)
	post := func(run uuid.UUID, body string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/runs/"+run.String()+"/messages", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+sandbox.SignRunToken(key, run, org, time.Now().Add(time.Hour)))
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(data)
	}

	messages, stop := relay.Listen(agentRun)
	defer stop()
	received := make(chan remote.Message, 1)
	go func() { received <- <-messages }()
	status, body := post(agentRun, `{"kind":"event","iteration":1,"event":{"Kind":"text","Text":"done"}}`)
	assert.Equal(t, http.StatusAccepted, status, body)
	message := <-received
	assert.Equal(t, remote.KindEvent, message.Kind)
	assert.Equal(t, "done", message.DecodedEvent().Text)

	status, body = post(unwatchedRun, `{"kind":"started"}`)
	assert.Equal(t, http.StatusConflict, status)
	assert.Contains(t, body, "no host is waiting for this run: run "+unwatchedRun.String())

	status, body = post(testRun, `{"kind":"started"}`)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "run "+testRun.String()+" is a test run; only agent runs send messages")

	status, body = post(agentRun, `not json`)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body, "decode the message of run "+agentRun.String())
}
