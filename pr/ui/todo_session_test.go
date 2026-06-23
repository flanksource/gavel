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

	"github.com/flanksource/gavel/todos/cmux"
)

func TestHandleTodoSessionStreamEmitsEvents(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := t.TempDir()
	sessionID := "sess-test"
	path, err := cmux.SessionLogPath(dir, sessionID)
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

	body := rec.Body.String()
	for _, want := range []string{`"kind":"tool_use"`, `"tool":"Bash"`, `ls -la`, `"kind":"turn_end"`} {
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
	path, err := cmux.SessionLogPath(dir, sessionID)
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
	var got cmux.SessionStats
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
