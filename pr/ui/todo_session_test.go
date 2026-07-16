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
	"github.com/flanksource/captain/pkg/session"
	"github.com/google/uuid"
)

type fakeCaptainSessionStore struct {
	run        *captaindb.Session
	transcript *captaindb.Session
	overview   *captaindb.SessionOverview
	messages   []captaindb.TranscriptMessage
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

func (f fakeCaptainSessionStore) GetSessionOverviewByIdentity(context.Context, string) (*captaindb.SessionOverview, error) {
	if f.overview == nil {
		return nil, captaindb.ErrSessionNotFound
	}
	return f.overview, nil
}

func (f fakeCaptainSessionStore) ListTranscriptMessages(context.Context, captaindb.TranscriptPage) ([]captaindb.TranscriptMessage, error) {
	return f.messages, nil
}

func TestStreamCaptainSessionEmitsUnifiedMessages(t *testing.T) {
	sessionID := "019f5b08-dfee-7b80-aef3-15a58eb6371b"
	transcriptID := uuid.New()
	messageID := uuid.New()
	now := time.Now().UTC()
	parts, err := json.Marshal([]session.Part{{
		Type: session.PartTool, ToolName: "Bash", ToolCallID: "call-1",
		State:  session.ToolStateOutputAvailable,
		Input:  json.RawMessage(`{"command":"go test ./..."}`),
		Output: json.RawMessage(`"ok"`),
	}})
	if err != nil {
		t.Fatal(err)
	}
	store := fakeCaptainSessionStore{
		run: &captaindb.Session{
			ID: uuid.New(), Source: "gavel", Provider: "headless-codex",
			ProviderSessionID: sessionID, LifecycleStatus: captaindb.SessionLifecycleRunning,
		},
		transcript: &captaindb.Session{
			ID: transcriptID, Source: "codex", CWD: "/work/captain", ProviderSessionID: sessionID,
		},
		messages: []captaindb.TranscriptMessage{{
			ID: messageID, SessionID: transcriptID, Sequence: 1, Role: "assistant",
			Parts: parts, OccurredAt: &now,
		}},
	}
	resolved, err := resolveCaptainSession(t.Context(), store, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if !resolved.known || resolved.run == nil || resolved.transcript == nil {
		t.Fatalf("resolution = %#v, want run and transcript", resolved)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest("GET", "/api/todos/session/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	streamCaptainSession(rec, req, store, sessionID, resolved)

	body := rec.Body.String()
	for _, want := range []string{
		`event: entry`, `"id":"` + messageID.String() + `"`, `"role":"assistant"`,
		`"toolName":"Bash"`, `go test ./...`, `"source":"codex"`, `"cwd":"/work/captain"`,
		`"agentId":"` + transcriptID.String() + `"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("Captain session stream missing %q in:\n%s", want, body)
		}
	}
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

func TestHandleTodoSessionStreamEmitsEvents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	sessionID := "sess-test"
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		t.Fatalf("SessionLogPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	log := strings.Join([]string{
		`{"type":"assistant","sessionId":"sess-test","message":{"stop_reason":"tool_use","content":[{"type":"tool_use","name":"Bash","input":{"command":"ls -la"}}]}}`,
		`{"type":"assistant","sessionId":"sess-test","message":{"stop_reason":"end_turn","content":[{"type":"text","text":"done"}]}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(path, []byte(log), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	s := &Server{}
	target := "/api/todos/session/stream?sessionId=" + sessionID + "&dir=" + url.QueryEscape(dir)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest("GET", target, nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	s.handleTodoSessionStream(rec, req)

	// The stream now emits raw captain SessionEntry records (the schema the
	// clicky-ui SessionViewer consumes), so assert on the entry/block fields.
	body := rec.Body.String()
	for _, want := range []string{`"type":"assistant"`, `"name":"Bash"`, `ls -la`, `"text":"done"`, `"stop_reason":"end_turn"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("session stream missing %q in:\n%s", want, body)
		}
	}
}

func TestHandleTodoSessionStreamSurfacesSubagent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	sessionID := "sess-sub"
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		t.Fatalf("SessionLogPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// A Task tool call dispatching an Explore subagent must carry the
	// subagent_type through to the streamed entry so the viewer can filter it.
	log := `{"type":"assistant","sessionId":"sess-sub","message":{"stop_reason":"tool_use","content":[{"type":"tool_use","name":"Task","input":{"description":"find the runner","subagent_type":"Explore"}}]}}` + "\n"
	if err := os.WriteFile(path, []byte(log), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	s := &Server{}
	target := "/api/todos/session/stream?sessionId=" + sessionID + "&dir=" + url.QueryEscape(dir)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest("GET", target, nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	s.handleTodoSessionStream(rec, req)

	body := rec.Body.String()
	for _, want := range []string{`"name":"Task"`, `"subagent_type":"Explore"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("session stream missing %q in:\n%s", want, body)
		}
	}
}

func TestHandleTodoSessionStreamEmitsErrorEvent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	sessionID := "sess-err"
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		t.Fatalf("SessionLogPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// A synthetic API error (stop_sequence) must stream as an entry flagged
	// isApiErrorMessage with the HTTP status, so the viewer renders it as an
	// error rather than a normal completion.
	log := `{"type":"assistant","sessionId":"sess-err","message":{"model":"<synthetic>","stop_reason":"stop_sequence","content":[{"type":"text","text":"API Error: 529 Overloaded"}]},"error":"server_error","isApiErrorMessage":true,"apiErrorStatus":529}` + "\n"
	if err := os.WriteFile(path, []byte(log), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}

	s := &Server{}
	target := "/api/todos/session/stream?sessionId=" + sessionID + "&dir=" + url.QueryEscape(dir)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req := httptest.NewRequest("GET", target, nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	s.handleTodoSessionStream(rec, req)

	body := rec.Body.String()
	for _, want := range []string{`"isApiErrorMessage":true`, `"error":"server_error"`, `"apiErrorStatus":529`, `529 Overloaded`} {
		if !strings.Contains(body, want) {
			t.Fatalf("session stream missing %q in:\n%s", want, body)
		}
	}
}

func TestHandleTodoSessionStreamRequiresSessionID(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest("GET", "/api/todos/session/stream", nil)
	rec := httptest.NewRecorder()
	s.handleTodoSessionStream(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
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
