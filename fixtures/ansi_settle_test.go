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

	got := strings.TrimRight(settleANSI(raw, 0), "\n")
	if got != "line C" {
		t.Fatalf("settleANSI collapsed wrong: got %q", got)
	}
}

func TestSettleANSI_UnderClearLeavesStaleLines(t *testing.T) {
	// Under-clear bug: renderer writes "B", then tries to redraw "C" without
	// a preceding cursor-up. The stale "B" is left visible above "C".
	raw := "line A\nline B\n\x1b[2Kline C\n"

	got := settleANSI(raw, 0)
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
	got := strings.TrimRight(settleANSI(raw, 0), "\n")
	if got != "redhidden" {
		t.Fatalf("settleANSI should drop SGR/cursor-visibility: got %q", got)
	}
}

func TestDuplicateLines_TrueDuplicate(t *testing.T) {
	// Same line emitted twice with no intervening clear.
	raw := "line A\nline A\n"
	dups := duplicateLines(raw, 0)
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
	dups := duplicateLines(raw, 0)
	if len(dups) != 0 {
		t.Fatalf("spinner frames should not register as duplicates, got %+v", dups)
	}
}

func TestDuplicateLines_IgnoresEmptyLines(t *testing.T) {
	raw := "\n\n\nfoo\n\n\n\nfoo\n\n"
	dups := duplicateLines(raw, 0)
	if len(dups) != 1 || dups[0].Text != "foo" || dups[0].Count != 2 {
		t.Fatalf("expected foo×2, got %+v", dups)
	}
}

func TestSettleANSI_WrapsAtWidth(t *testing.T) {
	// A 10-char line settled at width 4 occupies 3 physical rows.
	got := settleANSI("abcdefghij", 4)
	want := "abcd\nefgh\nij"
	if got != want {
		t.Fatalf("width-4 wrap: got %q want %q", got, want)
	}
}

func TestSettleANSI_WidthZeroDoesNotWrap(t *testing.T) {
	// width 0 preserves the legacy unbounded behavior used by fixture CEL.
	got := settleANSI("abcdefghij", 0)
	if got != "abcdefghij" {
		t.Fatalf("width-0 should not wrap: got %q", got)
	}
}

func TestSettleANSI_CursorUpUnderCountsWrappedRows(t *testing.T) {
	// Reproduces the status-screen smear. A renderer prints a header, then a
	// 12-char row that wraps into 2 physical rows at width 8, then moves the
	// cursor up by the LOGICAL line count (1) to redraw the row. Because the
	// row actually occupied 2 physical rows, the cursor-up lands on the wrapped
	// tail instead of the row start, so the redraw overwrites mid-content and
	// the stale tail survives.
	const width = 8
	raw := "HEADER\n" + // physical row 0
		"ROWADDXXXXXX\n" + // physical rows 1-2 ("ROWADDXX" + "XXXX")
		"\x1b[1A\x1b[2KNEW" // up 1 physical row (lands on "XXXX"), erase, write NEW

	settled := settleANSI(raw, width)
	lines := strings.Split(settled, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected wrapped grid, got %q", settled)
	}
	// Header is untouched; the first wrapped physical row "ROWADDXX" survives
	// because the under-counted cursor-up never reached it — that is the smear.
	if lines[0] != "HEADER" {
		t.Fatalf("header clobbered: %q", settled)
	}
	if !strings.Contains(settled, "ROWADDXX") {
		t.Fatalf("expected stale wrapped row 'ROWADDXX' to survive under-count, got %q", settled)
	}
	if !strings.Contains(settled, "NEW") {
		t.Fatalf("expected redraw 'NEW' present, got %q", settled)
	}
}
