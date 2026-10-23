package parsers

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"strings"
	"testing"
	"time"
)

type capturingSink struct {
	updates []Test
}

func (c *capturingSink) UpdatePackageProgress(pkgPath string, framework Framework, progress Test) {
	progress.PackagePath, progress.Framework = pkgPath, framework
	c.updates = append(c.updates, progress)
}

func (c *capturingSink) last(t *testing.T) Test {
	t.Helper()
	if len(c.updates) == 0 {
		t.Fatal("expected at least one progress update")
	}
	return c.updates[len(c.updates)-1]
}

const vitestWorkDir = "/repo"

func vitestFrame(json string) string {
	return vitestFramePrefix + json + "\n"
}

func TestVitestStreamWriterReportsRunningTest(t *testing.T) {
	sink := &capturingSink{}
	w := NewVitestStreamWriter(vitestWorkDir, "./web", sink)
	_, _ = w.Write([]byte(vitestFrame(`{"event":"test-ready","file":"/repo/web/sum.test.ts","name":"adds","suite":["math"],"line":4}`)))

	want := Test{
		Name: "./web", PackagePath: "./web", Framework: Vitest, Pending: true,
		Children: []Test{{Name: "sum.test.ts", File: "web/sum.test.ts", Children: []Test{{
			Name: "math", Suite: []string{"math"}, Children: []Test{{
				Name: "adds", Suite: []string{"math"}, Line: 4, Framework: Vitest, PackagePath: "./web", Pending: true,
			}},
		}}}},
	}
	if diff := cmp.Diff(want, sink.last(t), cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("progress tree mismatch (-want +got):\n%s", diff)
	}
}

func TestVitestStreamWriterReplacesRunningTestWithResult(t *testing.T) {
	sink := &capturingSink{}
	w := NewVitestStreamWriter(vitestWorkDir, "./web", sink)
	stream := vitestFrame(`{"event":"module-start","file":"/repo/web/sum.test.ts"}`) +
		vitestFrame(`{"event":"test-ready","file":"/repo/web/sum.test.ts","name":"adds","line":4}`) +
		"stray console output from a test\n" +
		vitestFrame(`{"event":"test-result","file":"/repo/web/sum.test.ts","name":"adds","line":4,"state":"passed","duration":1.5}`) +
		vitestFrame(`{"event":"test-result","file":"/repo/web/sum.test.ts","name":"subtracts","line":9,"state":"failed","duration":2,"errors":["\u001b[31mexpected 1 to be 2\u001b[39m"]}`) +
		vitestFrame(`{"event":"test-result","file":"/repo/web/sum.test.ts","name":"later","line":12,"state":"skipped"}`)

	// Feed in small chunks so frames straddle Write boundaries.
	for len(stream) > 0 {
		n := min(7, len(stream))
		if _, err := w.Write([]byte(stream[:n])); err != nil {
			t.Fatalf("write: %v", err)
		}
		stream = stream[n:]
	}
	w.Flush()

	want := Tests{
		{Name: "adds", Line: 4, Framework: Vitest, PackagePath: "./web", Passed: true, Duration: 1500 * time.Microsecond},
		{Name: "subtracts", Line: 9, Framework: Vitest, PackagePath: "./web", Failed: true, Duration: 2 * time.Millisecond,
			Message: "expected 1 to be 2", Stderr: "\x1b[31mexpected 1 to be 2\x1b[39m"},
		{Name: "later", Line: 12, Framework: Vitest, PackagePath: "./web", Skipped: true},
	}
	got := sink.last(t)
	if len(got.Children) != 1 || got.Children[0].File != "web/sum.test.ts" {
		t.Fatalf("want one file node for web/sum.test.ts, got %+v", got.Children)
	}
	if diff := cmp.Diff(want, got.Children[0].Children, cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("file children mismatch (-want +got):\n%s", diff)
	}
}

func TestVitestStreamWriterReportsModuleCollectionError(t *testing.T) {
	sink := &capturingSink{}
	w := NewVitestStreamWriter(vitestWorkDir, ".", sink)
	_, _ = w.Write([]byte(vitestFrame(`{"event":"module-end","file":"/repo/broken.test.ts","state":"failed","errors":["SyntaxError: Unexpected token"]}`)))

	want := Tests{{
		Name: "broken.test.ts", File: "broken.test.ts", Failed: true, Children: []Test{{
			Name: "broken.test.ts", Framework: Vitest, PackagePath: ".", Failed: true,
			Message: "SyntaxError: Unexpected token", Stderr: "SyntaxError: Unexpected token",
		}},
	}}
	if diff := cmp.Diff(want, sink.last(t).Children, cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("collection error mismatch (-want +got):\n%s", diff)
	}
}

func TestVitestStreamWriterIgnoresNonFrameOutput(t *testing.T) {
	sink := &capturingSink{}
	w := NewVitestStreamWriter(vitestWorkDir, ".", sink)
	_, _ = w.Write([]byte(strings.Join([]string{"RUN v5.0.5 /repo", "", "Test Files 1 passed"}, "\n")))
	w.Flush()
	if len(sink.updates) != 0 {
		t.Fatalf("want no updates for plain output, got %+v", sink.updates)
	}
}
