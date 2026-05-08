package aifix

import (
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
)

func TestRenderEvent_Text(t *testing.T) {
	out := renderEvent(0, captainai.Event{Kind: captainai.EventText, Text: "hello world"})
	if !strings.Contains(out, "hello world") {
		t.Errorf("missing text payload: %q", out)
	}
	if !strings.Contains(out, "iter 0") {
		t.Errorf("missing iter prefix: %q", out)
	}
}

func TestRenderEvent_ToolUse_PicksFilePathFromInput(t *testing.T) {
	out := renderEvent(2, captainai.Event{
		Kind:  captainai.EventToolUse,
		Tool:  "Edit",
		Input: map[string]any{"file_path": "/repo/foo.go"},
	})
	if !strings.Contains(out, "Edit") || !strings.Contains(out, "/repo/foo.go") {
		t.Errorf("tool/file_path missing from output: %q", out)
	}
}

func TestRenderEvent_ResultIncludesCostAndStatus(t *testing.T) {
	out := renderEvent(1, captainai.Event{
		Kind:    captainai.EventResult,
		Success: true,
		CostUSD: 0.0123,
	})
	if !strings.Contains(out, "ok") {
		t.Errorf("missing ok status: %q", out)
	}
	if !strings.Contains(out, "$0.0123") {
		t.Errorf("missing cost: %q", out)
	}
}

func TestRenderEvent_SystemWithoutSessionIDIsEmpty(t *testing.T) {
	if got := renderEvent(0, captainai.Event{Kind: captainai.EventSystem}); got != "" {
		t.Errorf("expected empty render for system without session_id; got %q", got)
	}
}

func TestTruncateOneLine_CollapsesNewlinesAndCaps(t *testing.T) {
	got := truncateOneLine("a\nb\nc", 100)
	if got != "a b c" {
		t.Errorf("got %q, want \"a b c\"", got)
	}
	long := strings.Repeat("x", 50)
	got = truncateOneLine(long, 10)
	// `…` is 3 bytes in UTF-8 — cap is on byte length per the implementation.
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected truncation ending in …, got %q", got)
	}
	xCount := strings.Count(got, "x")
	if xCount != 9 {
		t.Errorf("expected 9 'x' chars before …, got %d (out=%q)", xCount, got)
	}
}
