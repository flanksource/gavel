package aifix

import (
	"bytes"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
)

func TestRenderEvent_Text(t *testing.T) {
	out := renderEvent("[m 0%] ", captainai.Event{Kind: captainai.EventText, Text: "hello world"})
	if !strings.Contains(out, "hello world") {
		t.Errorf("missing text payload: %q", out)
	}
	if !strings.Contains(out, "m 0%") {
		t.Errorf("missing prefix: %q", out)
	}
}

func TestRenderEvent_ToolUse_PicksFilePathFromInput(t *testing.T) {
	out := renderEvent("[m 0%] ", captainai.Event{
		Kind:  captainai.EventToolUse,
		Tool:  "Edit",
		Input: map[string]any{"file_path": "/repo/foo.go"},
	})
	if !strings.Contains(out, "Edit") || !strings.Contains(out, "/repo/foo.go") {
		t.Errorf("tool/file_path missing from output: %q", out)
	}
}

func TestRenderEvent_ResultIncludesCostAndStatus(t *testing.T) {
	out := renderEvent("[m 0%] ", captainai.Event{
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
	if got := renderEvent("[m 0%] ", captainai.Event{Kind: captainai.EventSystem}); got != "" {
		t.Errorf("expected empty render for system without session_id; got %q", got)
	}
}

func TestEventPrefix_ShowsContextPercent(t *testing.T) {
	// 50000 of 200000 tokens = 25%.
	out := eventPrefix("claude-x", 50000, 200000)
	if !strings.Contains(out, "claude-x 25%") {
		t.Errorf("expected `claude-x 25%%` in prefix, got %q", out)
	}
}

func TestEventPrefix_OmitsPercentWhenWindowUnknown(t *testing.T) {
	out := eventPrefix("claude-x", 50000, 0)
	if !strings.Contains(out, "claude-x") || strings.Contains(out, "%") {
		t.Errorf("expected model-only prefix without %%, got %q", out)
	}
}

func TestNewStderrRenderer_TracksUsageAcrossEvents(t *testing.T) {
	var buf bytes.Buffer
	render := NewStderrRenderer(&buf, "claude-x", 200000)

	// Before any result the percentage is 0%.
	render(0, captainai.Event{Kind: captainai.EventText, Text: "starting"})
	// A result reports usage that fills 25% of the window (40k in + 10k out).
	render(0, captainai.Event{
		Kind:  captainai.EventResult,
		Usage: &captainai.Usage{InputTokens: 40000, OutputTokens: 10000},
	})
	// A subsequent line reflects the last-known usage.
	render(0, captainai.Event{Kind: captainai.EventText, Text: "continuing"})

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 rendered lines, got %d: %q", len(lines), buf.String())
	}
	if !strings.Contains(lines[0], "claude-x 0%") {
		t.Errorf("first line should show 0%%, got %q", lines[0])
	}
	if !strings.Contains(lines[2], "claude-x 25%") {
		t.Errorf("line after result should show 25%%, got %q", lines[2])
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
