package cmux

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/ai/history"
)

func TestSessionLogPath(t *testing.T) {
	got, err := sessionLogPath("/tmp/work", "abc-123")
	if err != nil {
		t.Fatalf("sessionLogPath() error = %v", err)
	}
	wantSuffix := filepath.Join(history.NormalizePath("/tmp/work"), "abc-123.jsonl")
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("sessionLogPath() = %q, want suffix %q", got, wantSuffix)
	}
	if !strings.HasPrefix(got, history.GetProjectsDir()) {
		t.Fatalf("sessionLogPath() = %q, want prefix %q", got, history.GetProjectsDir())
	}
}

func TestSessionLogPathRequiresSessionID(t *testing.T) {
	if _, err := sessionLogPath("/tmp/work", ""); err == nil {
		t.Fatal("expected error for empty session id")
	}
}

func writeSessionLog(t *testing.T, path string, lines ...string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSessionTailerStreamsAndCompletesOnEndTurn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.jsonl")
	writeSessionLog(t, path,
		`{"type":"assistant","sessionId":"s","message":{"stop_reason":"tool_use","content":[{"type":"tool_use","name":"Bash","input":{"command":"ls"}}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result"}]}}`,
		`{"type":"assistant","sessionId":"s","message":{"stop_reason":"end_turn","content":[{"type":"text","text":"done"}]}}`,
	)

	tailer := sessionTailer{path: path, pollInterval: time.Millisecond, appearTimeout: 200 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var kinds []history.SessionEventKind
	completed, err := tailer.tail(ctx, func(ev history.SessionEvent) { kinds = append(kinds, ev.Kind) })
	if err != nil {
		t.Fatalf("tail() error = %v", err)
	}
	if !completed {
		t.Fatal("tail() completed = false, want true")
	}
	wantKinds := []history.SessionEventKind{history.EventToolUse, history.EventAssistantText, history.EventTurnEnd}
	if len(kinds) != len(wantKinds) {
		t.Fatalf("kinds = %v, want %v", kinds, wantKinds)
	}
	for i := range wantKinds {
		if kinds[i] != wantKinds[i] {
			t.Fatalf("kinds = %v, want %v", kinds, wantKinds)
		}
	}
}

func TestSessionTailerCompletesOnQuiescence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.jsonl")
	// No end_turn — only a mid-turn entry; completion must come from quiescence.
	writeSessionLog(t, path,
		`{"type":"assistant","message":{"stop_reason":"tool_use","content":[{"type":"text","text":"working"}]}}`,
	)

	tailer := sessionTailer{path: path, pollInterval: time.Millisecond, appearTimeout: 200 * time.Millisecond, quiescePeriod: 20 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	completed, err := tailer.tail(ctx, func(history.SessionEvent) {})
	if err != nil || !completed {
		t.Fatalf("tail() = (%v, %v), want (true, nil)", completed, err)
	}
}

func TestSessionTailerReturnsNotFound(t *testing.T) {
	tailer := sessionTailer{path: filepath.Join(t.TempDir(), "missing.jsonl"), pollInterval: time.Millisecond, appearTimeout: 5 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := tailer.tail(ctx, func(history.SessionEvent) {}); !errors.Is(err, errSessionLogNotFound) {
		t.Fatalf("tail() error = %v, want errSessionLogNotFound", err)
	}
}

func TestSessionTailerWaitsForAppearance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.jsonl")
	go func() {
		time.Sleep(20 * time.Millisecond)
		writeSessionLog(t, path,
			`{"type":"assistant","message":{"stop_reason":"end_turn","content":[{"type":"text","text":"hi"}]}}`,
		)
	}()

	tailer := sessionTailer{path: path, pollInterval: time.Millisecond, appearTimeout: 500 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	completed, err := tailer.tail(ctx, func(history.SessionEvent) {})
	if err != nil || !completed {
		t.Fatalf("tail() = (%v, %v), want (true, nil)", completed, err)
	}
}
