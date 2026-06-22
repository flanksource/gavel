package cmux

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseHookSessionsSupportsArrayAndMap(t *testing.T) {
	data := []byte(`{
		"sessions": [
			{"sessionId":"s1","workspaceId":"w1","surfaceId":"f1","cwd":"/repo","pid":42,"lifecycle":"running","updatedAt":"2026-06-22T00:00:00Z"}
		],
		"byId": {
			"s2": {"workspace_id":"w2","surface_id":"f2","cwd":"/repo","pid":"43","state":"idle","updated_at":"2026-06-22T00:00:01Z"}
		}
	}`)

	sessions, err := parseHookSessions(data)
	if err != nil {
		t.Fatalf("parseHookSessions() error = %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("sessions = %d, want 2: %#v", len(sessions), sessions)
	}
	byID := map[string]HookSession{}
	for _, sess := range sessions {
		byID[sess.SessionID] = sess
	}
	if byID["s1"].WorkspaceID != "w1" || byID["s1"].PID != 42 {
		t.Fatalf("unexpected s1 session: %#v", byID["s1"])
	}
	if byID["s2"].Lifecycle != "idle" || byID["s2"].PID != 43 {
		t.Fatalf("unexpected s2 session: %#v", byID["s2"])
	}
}

func TestSessionStoreWaitForIdle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "claude-hook-sessions.json")
	if err := os.WriteFile(path, []byte(`[{"sessionId":"s1","cwd":"/repo","lifecycle":"idle"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	store := &SessionStore{Dir: dir, PollInterval: time.Millisecond}

	sess, err := store.WaitForIdle(context.Background(), "claude", "/repo", 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForIdle() error = %v", err)
	}
	if sess.SessionID != "s1" {
		t.Fatalf("session = %#v, want s1", sess)
	}
}
