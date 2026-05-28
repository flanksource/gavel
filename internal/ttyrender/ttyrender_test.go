package ttyrender

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestStateWriteSecondCallClearsPrevious(t *testing.T) {
	var buf bytes.Buffer
	state := State{}

	first := "line one\nline two\nline three\n"
	if err := state.Write(&buf, first); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if got := buf.String(); got != first {
		t.Fatalf("first write should pass through unchanged.\n got: %q\nwant: %q", got, first)
	}

	buf.Reset()
	second := "alpha\nbeta\n"
	if err := state.Write(&buf, second); err != nil {
		t.Fatalf("second write: %v", err)
	}
	wantPrefix := fmt.Sprintf("\x1b[%dA\x1b[J", 3)
	got := buf.String()
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("second write missing clear prefix.\n got: %q\nwant prefix: %q", got, wantPrefix)
	}
	if !strings.HasSuffix(got, second) {
		t.Fatalf("second write missing rendered payload.\n got: %q\nwant suffix: %q", got, second)
	}
}

func TestStateWriteEmptyRenderDoesNotEmitClear(t *testing.T) {
	var buf bytes.Buffer
	state := State{}
	if err := state.Write(&buf, ""); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := state.Write(&buf, "hello\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := buf.String(); got != "hello\n" {
		t.Fatalf("empty first render should not arm clear sequence.\n got: %q", got)
	}
}

func TestCountLines(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"a", 1},
		{"a\n", 1},
		{"a\nb", 2},
		{"a\nb\n", 2},
		{"a\nb\nc\n", 3},
	}
	for _, c := range cases {
		if got := CountLines(c.in); got != c.want {
			t.Errorf("CountLines(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
