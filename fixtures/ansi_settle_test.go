package fixtures

import (
	"strings"
	"testing"
)

func TestSettleANSI_CleanRedrawCollapsesFrames(t *testing.T) {
	// Emit three "frames" where each redraws the line below. A correct
	// renderer produces settled text equal to the final frame only.
	raw := "line A\n" +
		"\x1b[1A\x1b[2Kline B\n" +
		"\x1b[1A\x1b[2Kline C\n"

	got := strings.TrimRight(settleANSI(raw), "\n")
	if got != "line C" {
		t.Fatalf("settleANSI collapsed wrong: got %q", got)
	}
}

func TestSettleANSI_UnderClearLeavesStaleLines(t *testing.T) {
	// Under-clear bug: renderer writes "B", then tries to redraw "C" without
	// a preceding cursor-up. The stale "B" is left visible above "C".
	raw := "line A\nline B\n\x1b[2Kline C\n"

	got := settleANSI(raw)
	// We expect both "line B" and "line C" to be present.
	if !strings.Contains(got, "line B") {
		t.Fatalf("expected 'line B' to survive missing cursor-up, got %q", got)
	}
	if !strings.Contains(got, "line C") {
		t.Fatalf("expected 'line C' to be present, got %q", got)
	}
}

func TestSettleANSI_StripsSGRAndControlCodes(t *testing.T) {
	raw := "\x1b[31mred\x1b[0m\x1b[?25lhidden\x1b[?25h\n"
	got := strings.TrimRight(settleANSI(raw), "\n")
	if got != "redhidden" {
		t.Fatalf("settleANSI should drop SGR/cursor-visibility: got %q", got)
	}
}

func TestDuplicateLines_TrueDuplicate(t *testing.T) {
	// Same line emitted twice with no intervening clear.
	raw := "line A\nline A\n"
	dups := duplicateLines(raw)
	if len(dups) != 1 || dups[0].Text != "line A" || dups[0].Count != 2 {
		t.Fatalf("expected single duplicate of 'line A' count=2, got %+v", dups)
	}
}

func TestDuplicateLines_SpinnerFramesNotDuplicate(t *testing.T) {
	// Spinner frames that overwrite the same line via cursor-up + clear must
	// NOT be flagged as duplicates — that's the false-positive the settler
	// exists to prevent.
	raw := "⠋ task\n" +
		"\x1b[1A\x1b[2K⠙ task\n" +
		"\x1b[1A\x1b[2K⠹ task\n"
	dups := duplicateLines(raw)
	if len(dups) != 0 {
		t.Fatalf("spinner frames should not register as duplicates, got %+v", dups)
	}
}

func TestDuplicateLines_IgnoresEmptyLines(t *testing.T) {
	raw := "\n\n\nfoo\n\n\n\nfoo\n\n"
	dups := duplicateLines(raw)
	if len(dups) != 1 || dups[0].Text != "foo" || dups[0].Count != 2 {
		t.Fatalf("expected foo×2, got %+v", dups)
	}
}

func TestHasDuplicateLines(t *testing.T) {
	if hasDuplicateLines("a\nb\nc\n") {
		t.Fatal("no duplicates expected")
	}
	if !hasDuplicateLines("a\na\n") {
		t.Fatal("duplicate expected")
	}
}
