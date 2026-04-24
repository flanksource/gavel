package parsers

import (
	"strings"
	"testing"
)

type capturingSink struct {
	updates []Test
}

func (c *capturingSink) UpdateGinkgoProgress(progress Test) {
	c.updates = append(c.updates, progress)
}

func TestGinkgoStreamWriterDetectsSpecCompletions(t *testing.T) {
	sink := &capturingSink{}
	w := NewGinkgoStreamWriter("pkg/foo", sink)

	// Synthetic ginkgo -v output: delimiter, hierarchy, leaf, then a spec
	// completion line. This mirrors the minimal shape the parser must
	// handle to surface per-spec events to the UI.
	stream := strings.Join([]string{
		"Running Suite: Foo Suite",
		"------------------------------",
		"Describe Foo",
		"  when X happens",
		"    should Y",
		"/home/user/foo_test.go:42",
		"• [0.123 seconds]",
		"------------------------------",
		"Describe Bar",
		"    boom",
		"/home/user/bar_test.go:7",
		"• [FAILED] [0.050 seconds]",
		"",
	}, "\n")

	if _, err := w.Write([]byte(stream)); err != nil {
		t.Fatalf("write: %v", err)
	}
	w.Flush()

	if len(sink.updates) == 0 {
		t.Fatal("expected at least one progress update")
	}

	last := sink.updates[len(sink.updates)-1]
	if !last.Pending {
		t.Error("package-level Test should be Pending during run")
	}
	if last.PackagePath != "pkg/foo" || last.Framework != Ginkgo {
		t.Errorf("package attrs wrong: got path=%q framework=%q", last.PackagePath, last.Framework)
	}

	// Count terminal children: exactly one passed + one failed.
	var passed, failed int
	for _, c := range last.Children {
		switch {
		case c.Passed:
			passed++
		case c.Failed:
			failed++
		}
	}
	if passed != 1 {
		t.Errorf("want 1 passed child, got %d", passed)
	}
	if failed != 1 {
		t.Errorf("want 1 failed child, got %d", failed)
	}
}

func TestGinkgoStreamWriterStripsANSIColors(t *testing.T) {
	sink := &capturingSink{}
	w := NewGinkgoStreamWriter("pkg", sink)

	// Colorized "• [0.010 seconds]" — the parser must see through the
	// escape codes to detect the terminal pattern.
	stream := "------------------------------\nDescribe A\n  spec\n/tmp/a_test.go:1\n\x1b[32m•\x1b[0m \x1b[90m[0.010 seconds]\x1b[0m\n"
	_, _ = w.Write([]byte(stream))
	w.Flush()

	if len(sink.updates) == 0 {
		t.Fatal("expected updates")
	}
	last := sink.updates[len(sink.updates)-1]
	var passed int
	for _, c := range last.Children {
		if c.Passed {
			passed++
		}
	}
	if passed != 1 {
		t.Errorf("want 1 passed child after ANSI-stripping, got %d (children=%+v)", passed, last.Children)
	}
}

func TestGinkgoStreamWriterHandlesPartialLines(t *testing.T) {
	sink := &capturingSink{}
	w := NewGinkgoStreamWriter("pkg", sink)

	// Write in small chunks to make sure the internal buffer stitches
	// lines correctly across Write() boundaries.
	chunks := []string{
		"------------",
		"------------------\n",
		"Describe Split\n",
		"  spec\n",
		"/tmp/s_test.go:9\n",
		"• [FAI",
		"LED] [0.001 seconds]\n",
	}
	for _, c := range chunks {
		if _, err := w.Write([]byte(c)); err != nil {
			t.Fatalf("write chunk %q: %v", c, err)
		}
	}
	w.Flush()

	if len(sink.updates) == 0 {
		t.Fatal("expected updates")
	}
	last := sink.updates[len(sink.updates)-1]
	var failed int
	for _, c := range last.Children {
		if c.Failed {
			failed++
		}
	}
	if failed != 1 {
		t.Errorf("want 1 failed child across chunked writes, got %d", failed)
	}
}
