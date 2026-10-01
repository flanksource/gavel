package ui

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

type fakeCaptainSessionStore struct {
	run        *captaindb.Session
	transcript *captaindb.Session
	overview   *captaindb.SessionOverview
	runs       []captaindb.PromptRun
}

func TestProjectThreadStatusKeepsAttemptFailureIndependent(t *testing.T) {
	status := projectThreadStatus([]captaindb.SessionOverview{{
		ID: uuid.New(), Source: "claude", LifecycleStatus: string(captaindb.SessionLifecycleCreated), MessageCount: 4,
	}})
	if status != "idle" {
		t.Fatalf("status = %q, want idle for a settled thread with messages", status)
	}
	status = projectThreadStatus([]captaindb.SessionOverview{{
		ID: uuid.New(), Source: "claude", LifecycleStatus: string(captaindb.SessionLifecycleCreated), ProcessActive: true,
	}})
	if status != "working" {
		t.Fatalf("status = %q, want working for a live process", status)
	}
}

func (f fakeCaptainSessionStore) ListPromptRuns(context.Context, captaindb.PromptRunFilter) ([]captaindb.PromptRun, error) {
	return f.runs, nil
}

func (f fakeCaptainSessionStore) GetSessionByIdentity(_ context.Context, _ string, source, _, _ string) (*captaindb.Session, error) {
	switch source {
	case "gavel":
		if f.run != nil {
			return f.run, nil
		}
	case "codex", "claude":
		if f.transcript != nil && f.transcript.Source == source {
			return f.transcript, nil
		}
	}
	return nil, captaindb.ErrSessionNotFound
}

// GetTranscriptSessionByIdentity stands in for Captain's sibling hop: the
// dashboard no longer chooses between the rows one provider id names.
func (f fakeCaptainSessionStore) GetTranscriptSessionByIdentity(context.Context, string) (*captaindb.Session, error) {
	if f.transcript == nil {
		return nil, captaindb.ErrSessionNotFound
	}
	return f.transcript, nil
}

func (f fakeCaptainSessionStore) GetSessionOverviewByIdentity(context.Context, string) (*captaindb.SessionOverview, error) {
	if f.overview == nil {
		return nil, captaindb.ErrSessionNotFound
	}
	return f.overview, nil
}

func TestCaptainSessionStatsUsesMonitoredSessionClock(t *testing.T) {
	now := time.Now().UTC()
	rootStarted := now.Add(-6 * time.Hour)
	agentStarted := now.Add(-35 * time.Second)
	lastActivity := now.Add(-2 * time.Second)
	providerSessionID := "session-clock"
	store := fakeCaptainSessionStore{
		run: &captaindb.Session{
			ID: uuid.New(), Source: "gavel", Provider: "headless-codex",
			ProviderSessionID: providerSessionID, CreatedAt: rootStarted,
		},
		transcript: &captaindb.Session{
			ID: uuid.New(), Source: "codex", AgentType: "codex",
			ProviderSessionID: providerSessionID, CreatedAt: agentStarted,
		},
		overview: &captaindb.SessionOverview{
			LifecycleStatus: string(captaindb.SessionLifecycleRunning),
			ActivityState:   string(captaindb.SessionActivityWorking),
			HealthState:     string(captaindb.SessionHealthHealthy),
			StartedAt:       &agentStarted, LastActivityAt: &lastActivity,
		},
	}
	resolved, err := resolveCaptainSession(t.Context(), store, providerSessionID)
	if err != nil {
		t.Fatal(err)
	}
	first, err := captainSessionStats(t.Context(), store, providerSessionID, resolved)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Found || !first.InProgress {
		t.Fatalf("stats = %#v, want found live monitored session", first.SessionStats)
	}
	if first.DurationMs < 34_000 || first.DurationMs > 37_000 {
		t.Fatalf("durationMs = %d, want monitored session age near 35s", first.DurationMs)
	}
	if first.DurationMs > int64(time.Hour/time.Millisecond) {
		t.Fatalf("durationMs = %d, admission-root clock leaked into agent stats", first.DurationMs)
	}

	time.Sleep(5 * time.Millisecond)
	second, err := captainSessionStats(t.Context(), store, providerSessionID, resolved)
	if err != nil {
		t.Fatal(err)
	}
	if second.DurationMs < first.DurationMs {
		t.Fatalf("live duration regressed from %d to %d", first.DurationMs, second.DurationMs)
	}
}

func TestCaptainSessionStatsHidesAdmissionWithoutMonitoredSession(t *testing.T) {
	resolved := captainSessionResolution{run: &captaindb.Session{
		ID: uuid.New(), Source: "gavel", ProviderSessionID: "admission-only",
	}, known: true}
	got, err := captainSessionStats(t.Context(), fakeCaptainSessionStore{}, "admission-only", resolved)
	if err != nil {
		t.Fatal(err)
	}
	if got.Found {
		t.Fatalf("found = true, admission-only sessions must not supply agent stats")
	}
}

func TestHandleTodoSessionStatsReportsUsage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	sessionID := "sess-stats"
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		t.Fatalf("SessionLogPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	log := strings.Join([]string{
		`{"type":"assistant","timestamp":"2026-06-23T10:00:00Z","message":{"model":"claude-opus-4-8","usage":{"input_tokens":120,"output_tokens":30,"cache_read_input_tokens":0,"cache_creation_input_tokens":0},"content":[{"type":"text","text":"hi"}]}}`,
		`{"type":"assistant","timestamp":"2026-06-23T10:00:20Z","message":{"model":"claude-opus-4-8","usage":{"input_tokens":80,"output_tokens":10,"cache_read_input_tokens":0,"cache_creation_input_tokens":0},"content":[{"type":"text","text":"done"}]}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(path, []byte(log), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	s := &Server{}
	target := "/api/todos/session/stats?sessionId=" + sessionID + "&dir=" + url.QueryEscape(dir)
	req := httptest.NewRequest("GET", target, nil)
	rec := httptest.NewRecorder()
	s.handleTodoSessionStats(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	var got cmuxprov.SessionStats
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v (body: %s)", err, rec.Body.String())
	}
	if !got.Found {
		t.Fatal("found = false, want true")
	}
	if got.InputTokens != 200 || got.OutputTokens != 40 {
		t.Fatalf("tokens = in:%d out:%d, want in:200 out:40", got.InputTokens, got.OutputTokens)
	}
	if got.Turns != 2 {
		t.Fatalf("turns = %d, want 2", got.Turns)
	}
	if got.DurationMs != 20_000 {
		t.Fatalf("durationMs = %d, want 20000", got.DurationMs)
	}
	if got.Model != "claude-opus-4-8" {
		t.Fatalf("model = %q, want claude-opus-4-8", got.Model)
	}
}

func TestHandleTodoSessionStatsRequiresSessionID(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest("GET", "/api/todos/session/stats", nil)
	rec := httptest.NewRecorder()
	s.handleTodoSessionStats(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
