package parsers

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// GinkgoProgressSink receives incremental updates while a ginkgo suite is
// running. Implementations fan these events out to the UI streamer.
// Progress is a summary Test with Children containing the running/finished
// specs observed so far.
type GinkgoProgressSink interface {
	UpdateGinkgoProgress(progress Test)
}

// ginkgo -v output markers. The format is remarkably stable: each spec
// completion prints a single-line header (•, P, S, or the leaf letter)
// followed by an optional `[FAILED]`/`[PENDING]`/`[SKIPPED]` tag and
// a `[N.NNN seconds]` duration. Spec boundaries are also delimited by
// 30 dashes, and the code location block gives us the hierarchy text.
var (
	ginkgoDelimiterRe  = regexp.MustCompile(`^-{5,}\s*$`)
	ginkgoRuntimeRe    = regexp.MustCompile(`\[\s*(\d+\.\d+)\s+seconds\s*\]`)
	ginkgoFailedRe     = regexp.MustCompile(`\[FAILED\]|\[PANICKED\]|\[TIMEDOUT\]`)
	ginkgoSkippedRe    = regexp.MustCompile(`\[SKIPPED\]`)
	ginkgoPendingRe    = regexp.MustCompile(`\[PENDING\]`)
	ginkgoRunningRunRe = regexp.MustCompile(`^\s*Running Suite:\s+(.+)$`)
)

// GinkgoStreamWriter is an io.Writer that parses a ginkgo `-v` stdout stream
// and pushes incremental progress into a sink. Safe for concurrent Write
// calls (the Process may flush in multiple goroutines).
type GinkgoStreamWriter struct {
	sink        GinkgoProgressSink
	packageName string

	mu        sync.Mutex
	buf       strings.Builder
	started   time.Time
	curSpec   *ginkgoSpecState
	completed []Test
	running   bool
}

type ginkgoSpecState struct {
	hierarchy []string
	leaf      string
	startAt   time.Time
	afterDash bool // true when we just saw the delimiter and are collecting hierarchy lines
}

// NewGinkgoStreamWriter builds a progress parser for a single package run.
func NewGinkgoStreamWriter(packageName string, sink GinkgoProgressSink) *GinkgoStreamWriter {
	return &GinkgoStreamWriter{
		sink:        sink,
		packageName: packageName,
		started:     time.Now(),
	}
}

// Write implements io.Writer. It buffers bytes until a newline, parses each
// complete line, and emits an update to the sink when something meaningful
// happens. Never returns an error — the wrapped stream is advisory.
func (w *GinkgoStreamWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	raw := w.buf.String()
	for {
		idx := strings.IndexByte(raw, '\n')
		if idx < 0 {
			break
		}
		line := raw[:idx]
		raw = raw[idx+1:]
		w.processLine(stripANSI(line))
	}
	w.buf.Reset()
	w.buf.WriteString(raw)
	return len(p), nil
}

// Flush finishes any trailing line and emits a final progress snapshot.
func (w *GinkgoStreamWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buf.Len() > 0 {
		w.processLine(stripANSI(w.buf.String()))
		w.buf.Reset()
	}
}

func (w *GinkgoStreamWriter) processLine(line string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return
	}

	if m := ginkgoRunningRunRe.FindStringSubmatch(trimmed); m != nil {
		w.running = true
		w.emit()
		return
	}

	if ginkgoDelimiterRe.MatchString(trimmed) {
		// boundary between spec blocks — start collecting a new hierarchy
		w.curSpec = &ginkgoSpecState{afterDash: true, startAt: time.Now()}
		return
	}

	if w.curSpec != nil && w.curSpec.afterDash {
		// Terminal lines end the "collecting hierarchy" phase.
		if isGinkgoTerminalLine(trimmed) {
			w.curSpec.afterDash = false
			w.finishCurrentSpec(trimmed)
			return
		}
		// Ignore code-location lines (start with a path and line number).
		if isCodeLocationLine(trimmed) {
			return
		}
		// Otherwise treat it as hierarchy text for the running spec.
		if w.curSpec.leaf == "" {
			w.curSpec.hierarchy = append(w.curSpec.hierarchy, trimmed)
		}
		w.curSpec.leaf = trimmed
		w.running = true
		w.emit()
		return
	}

	// Spec completion lines can appear without a preceding delimiter in some
	// output configurations (e.g. parallel mode aggregation).
	if isGinkgoTerminalLine(trimmed) {
		w.finishCurrentSpec(trimmed)
	}
}

func (w *GinkgoStreamWriter) finishCurrentSpec(line string) {
	if w.curSpec == nil {
		return
	}
	spec := w.curSpec
	w.curSpec = nil

	name := spec.leaf
	if name == "" && len(spec.hierarchy) > 0 {
		name = spec.hierarchy[len(spec.hierarchy)-1]
	}
	if name == "" {
		// nothing useful captured — don't invent a fake spec
		return
	}
	t := Test{
		Name:      name,
		Framework: Ginkgo,
	}
	switch {
	case ginkgoFailedRe.MatchString(line):
		t.Failed = true
	case ginkgoSkippedRe.MatchString(line):
		t.Skipped = true
	case ginkgoPendingRe.MatchString(line):
		t.Pending = true
	default:
		t.Passed = true
	}
	if m := ginkgoRuntimeRe.FindStringSubmatch(line); m != nil {
		if d, err := time.ParseDuration(m[1] + "s"); err == nil {
			t.Duration = d
		}
	}
	w.completed = append(w.completed, t)
	w.emit()
}

func (w *GinkgoStreamWriter) emit() {
	if w.sink == nil {
		return
	}
	progress := Test{
		Name:        w.packageName,
		Framework:   Ginkgo,
		PackagePath: w.packageName,
		Pending:     true,
	}
	progress.Children = append(progress.Children, w.completed...)
	if w.curSpec != nil && w.curSpec.leaf != "" {
		progress.Children = append(progress.Children, Test{
			Name:      w.curSpec.leaf,
			Framework: Ginkgo,
			Pending:   true,
		})
	}
	w.sink.UpdateGinkgoProgress(progress)
}

// isGinkgoTerminalLine is true when the line looks like a ginkgo spec-end
// marker: either a bare bullet, or one annotated with a [STATUS] or
// [duration] tag.
func isGinkgoTerminalLine(line string) bool {
	if line == "" {
		return false
	}
	if ginkgoRuntimeRe.MatchString(line) {
		return true
	}
	if ginkgoFailedRe.MatchString(line) || ginkgoSkippedRe.MatchString(line) || ginkgoPendingRe.MatchString(line) {
		return true
	}
	return false
}

// isCodeLocationLine matches ginkgo's hierarchy footer like
// "/path/to/foo_test.go:42".
var codeLocationRe = regexp.MustCompile(`\.go:\d+(\s|$)`)

func isCodeLocationLine(line string) bool {
	return codeLocationRe.MatchString(strings.TrimSpace(line))
}

var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI drops SGR color codes so delimiter / marker detection works even
// when ginkgo is printing with {{red}}…{{/}} style formatting translated to
// real escape sequences.
func stripANSI(s string) string {
	return ansiEscapeRe.ReplaceAllString(s, "")
}
