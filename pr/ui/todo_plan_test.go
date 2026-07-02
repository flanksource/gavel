package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	"github.com/flanksource/captain/pkg/claude"
)

// writePlanSession seeds a Claude session log whose assistant turn exits plan mode
// against planPath, carrying inline as the transcript's inline plan copy, so
// captaincli.RunPlan can recover it by session id.
func writePlanSession(t *testing.T, dir, sessionID, slug, planPath, inline string) {
	t.Helper()
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		t.Fatalf("SessionLogPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	lines := []map[string]any{
		{"type": "user", "sessionId": sessionID, "uuid": "u1", "timestamp": "2026-06-01T10:00:00Z", "cwd": dir, "slug": slug,
			"message": map[string]any{"role": "user", "content": []any{map[string]any{"type": "text", "text": "make a plan"}}}},
		{"type": "assistant", "sessionId": sessionID, "uuid": "a1", "timestamp": "2026-06-01T10:00:01Z", "cwd": dir, "slug": slug,
			"message": map[string]any{"role": "assistant", "content": []any{
				map[string]any{"type": "tool_use", "id": "tu1", "name": "ExitPlanMode",
					"input": map[string]any{"planFilePath": planPath, "plan": inline}}}}},
	}
	var b strings.Builder
	for _, line := range lines {
		data, err := json.Marshal(line)
		if err != nil {
			t.Fatalf("marshal session line: %v", err)
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write session log: %v", err)
	}
}

func TestHandleTodoSessionPlanReturnsAndSaves(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := t.TempDir()
	const sessionID, slug = "sess-plan", "tidy-otter"

	planPath := filepath.Join(claude.GetPlansDir(), slug+".md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, []byte("# Tidy otter\n\non-disk body"), 0o644); err != nil {
		t.Fatal(err)
	}
	writePlanSession(t, dir, sessionID, slug, planPath, "# inline body")

	s := &Server{}

	// GET recovers the plan, preferring the on-disk file over the inline copy.
	rec := httptest.NewRecorder()
	s.handleTodoSessionPlan(rec, httptest.NewRequest("GET", "/api/todos/session/plan?sessionId="+sessionID, nil))
	var got todoPlanResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode GET: %v (body=%s)", err, rec.Body.String())
	}
	if !got.Found {
		t.Fatalf("found=false, want true; body=%s", rec.Body.String())
	}
	if got.Content != "# Tidy otter\n\non-disk body" {
		t.Fatalf("content=%q, want on-disk body", got.Content)
	}
	if got.Path != planPath {
		t.Fatalf("path=%q, want %q", got.Path, planPath)
	}
	if !got.OnDisk {
		t.Fatalf("onDisk=false, want true")
	}

	// POST rewrites the plan file with the edited markdown.
	body, _ := json.Marshal(map[string]any{"sessionId": sessionID, "path": planPath, "content": "# edited plan"})
	rec = httptest.NewRecorder()
	s.handleTodoSessionPlanSave(rec, httptest.NewRequest("POST", "/api/todos/session/plan", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("save status=%d, body=%s", rec.Code, rec.Body.String())
	}
	saved, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(saved) != "# edited plan" {
		t.Fatalf("plan file=%q, want edited content", string(saved))
	}
}

func TestHandleTodoSessionPlanSaveRejectsOutsidePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	s := &Server{}
	outside := filepath.Join(home, "evil.md")

	body, _ := json.Marshal(map[string]any{"path": outside, "content": "pwned"})
	rec := httptest.NewRecorder()
	s.handleTodoSessionPlanSave(rec, httptest.NewRequest("POST", "/api/todos/session/plan", bytes.NewReader(body)))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want 403; body=%s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("a path outside the plans dir must not be written")
	}
}

func TestHandleTodoSessionPlanNotFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	s := &Server{}

	rec := httptest.NewRecorder()
	s.handleTodoSessionPlan(rec, httptest.NewRequest("GET", "/api/todos/session/plan?sessionId=unknown", nil))

	var got todoPlanResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v (body=%s)", err, rec.Body.String())
	}
	if got.Found {
		t.Fatalf("found=true for an unknown session, want false")
	}
}
