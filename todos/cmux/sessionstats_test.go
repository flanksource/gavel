package cmux

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// assistantLine builds a session-log assistant entry with the given usage so the
// stats parser has realistic input without embedding opaque literals.
func assistantLine(ts, model string, in, out, cacheRead, cacheCreate int) string {
	return fmt.Sprintf(
		`{"type":"assistant","timestamp":%q,"message":{"model":%q,"usage":{"input_tokens":%d,"output_tokens":%d,"cache_read_input_tokens":%d,"cache_creation_input_tokens":%d},"content":[{"type":"text","text":"hi"}]}}`,
		ts, model, in, out, cacheRead, cacheCreate,
	)
}

func TestComputeSessionStatsAggregatesUsage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.jsonl")
	// Two assistant turns 30s apart plus a non-assistant line that must be ignored
	// for token accounting but still counts toward the time span.
	writeSessionLog(t, path,
		assistantLine("2026-06-23T10:00:00Z", "claude-opus-4-8", 100, 20, 5, 50),
		`{"type":"user","timestamp":"2026-06-23T10:00:10Z","message":{"content":[{"type":"tool_result"}]}}`,
		assistantLine("2026-06-23T10:00:30Z", "claude-opus-4-8", 200, 40, 7, 0),
	)

	stats, err := computeSessionStats(path)
	if err != nil {
		t.Fatalf("computeSessionStats() error = %v", err)
	}
	if !stats.Found {
		t.Fatal("Found = false, want true")
	}
	if stats.InputTokens != 300 || stats.OutputTokens != 60 {
		t.Fatalf("tokens = in:%d out:%d, want in:300 out:60", stats.InputTokens, stats.OutputTokens)
	}
	if stats.CacheReadTokens != 12 || stats.CacheCreationTokens != 50 {
		t.Fatalf("cache tokens = read:%d create:%d, want read:12 create:50", stats.CacheReadTokens, stats.CacheCreationTokens)
	}
	if stats.TotalTokens != 300+60+12+50 {
		t.Fatalf("TotalTokens = %d, want %d", stats.TotalTokens, 300+60+12+50)
	}
	if stats.Turns != 2 {
		t.Fatalf("Turns = %d, want 2", stats.Turns)
	}
	if stats.Model != "claude-opus-4-8" {
		t.Fatalf("Model = %q, want claude-opus-4-8", stats.Model)
	}
	if stats.DurationMs != 30_000 {
		t.Fatalf("DurationMs = %d, want 30000", stats.DurationMs)
	}
	if stats.InProgress {
		t.Fatal("cold stats must not be in progress")
	}
}

func TestSessionStatsCacheMissingLogIsNotFound(t *testing.T) {
	c := NewSessionStatsCache()
	stats, err := c.Get("sess", filepath.Join(t.TempDir(), "missing.jsonl"))
	if err != nil {
		t.Fatalf("Get() error = %v, want nil for a missing log", err)
	}
	if stats.Found {
		t.Fatal("Found = true for a missing log, want false")
	}
}

func TestSessionStatsCacheColdCachesByMtime(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.jsonl")
	writeSessionLog(t, path, assistantLine("2026-06-23T10:00:00Z", "claude-opus-4-8", 100, 20, 0, 0))

	c := NewSessionStatsCache()
	first, err := c.Get("sess", path)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if first.OutputTokens != 20 {
		t.Fatalf("OutputTokens = %d, want 20", first.OutputTokens)
	}

	// Rewriting the log with a new mtime must invalidate the cold entry.
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	writeSessionLog(t, path,
		assistantLine("2026-06-23T10:00:00Z", "claude-opus-4-8", 100, 20, 0, 0),
		assistantLine("2026-06-23T10:00:05Z", "claude-opus-4-8", 100, 80, 0, 0),
	)
	second, err := c.Get("sess", path)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if second.OutputTokens != 100 {
		t.Fatalf("OutputTokens after rewrite = %d, want 100", second.OutputTokens)
	}
}

func TestSessionStatsCacheLivePreferredOverDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.jsonl")
	writeSessionLog(t, path, assistantLine("2026-06-23T10:00:00Z", "claude-opus-4-8", 1, 1, 0, 0))

	c := NewSessionStatsCache()
	acc := c.Begin("sess", "claude", "opus", "high", time.Now())
	acc.AddLine([]byte(assistantLine("2026-06-23T10:00:00Z", "claude-opus-4-8", 100, 20, 0, 0)))

	live, err := c.Get("sess", path)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !live.InProgress {
		t.Fatal("live session InProgress = false, want true")
	}
	if live.Effort != "high" || live.Agent != "claude" {
		t.Fatalf("identity = agent:%q effort:%q, want claude/high", live.Agent, live.Effort)
	}
	// The live accumulator (out:20) wins over the on-disk line (out:1).
	if live.OutputTokens != 20 {
		t.Fatalf("OutputTokens = %d, want 20 (live, not disk)", live.OutputTokens)
	}

	acc.Finish()
	done, err := c.Get("sess", path)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if done.InProgress {
		t.Fatal("finished session InProgress = true, want false")
	}
}
