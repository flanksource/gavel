package parsers

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/flanksource/commons/logger"
)

// vitestFramePrefix marks a progress event written to stdout by gavel's
// vitest reporter (testrunner/runners/vitest_reporter.mjs).
const vitestFramePrefix = "GAVEL_VITEST_V1:"

type vitestEvent struct {
	Event    string   `json:"event"`
	File     string   `json:"file"`
	Name     string   `json:"name"`
	Suite    []string `json:"suite"`
	Line     int      `json:"line"`
	State    string   `json:"state"`
	Duration float64  `json:"duration"`
	Errors   []string `json:"errors"`
}

func (e vitestEvent) key() string {
	return e.File + "\x00" + strings.Join(e.Suite, "\x00") + "\x00" + e.Name
}

// VitestStreamWriter is an io.Writer that decodes progress frames emitted by
// gavel's vitest reporter and pushes a running test tree into a sink. The
// stream is advisory: the JSON report remains the authoritative result.
type VitestStreamWriter struct {
	sink        ProgressSink
	packagePath string
	jest        *JestJSON

	mu    sync.Mutex
	buf   strings.Builder
	order []string
	tests map[string]Test
}

// NewVitestStreamWriter builds a progress parser for a single package run.
// workDir must match the one given to the package's JestJSON parser so file
// paths line up with the final report.
func NewVitestStreamWriter(workDir, packagePath string, sink ProgressSink) *VitestStreamWriter {
	return &VitestStreamWriter{
		sink:        sink,
		packagePath: packagePath,
		jest:        NewJestJSON(workDir, Vitest),
		tests:       map[string]Test{},
	}
}

func (w *VitestStreamWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	raw := w.buf.String()
	for {
		idx := strings.IndexByte(raw, '\n')
		if idx < 0 {
			break
		}
		w.processLine(raw[:idx])
		raw = raw[idx+1:]
	}
	w.buf.Reset()
	w.buf.WriteString(raw)
	return len(p), nil
}

// Flush processes any trailing unterminated line.
func (w *VitestStreamWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buf.Len() > 0 {
		w.processLine(w.buf.String())
		w.buf.Reset()
	}
}

func (w *VitestStreamWriter) processLine(line string) {
	idx := strings.Index(line, vitestFramePrefix)
	if idx < 0 {
		return
	}
	frame := strings.TrimSpace(line[idx+len(vitestFramePrefix):])
	var event vitestEvent
	if err := json.Unmarshal([]byte(frame), &event); err != nil {
		logger.Warnf("vitest progress %s: invalid reporter frame %q: %v", w.packagePath, frame, err)
		return
	}
	relFile := w.jest.relPath(event.File)
	switch event.Event {
	case "test-ready":
		t := w.toTest(event, relFile)
		t.Pending = true
		w.put(event.key(), t)
	case "test-result":
		w.put(event.key(), w.toTest(event, relFile))
	case "module-end":
		// Like JestJSON.Parse, a module error only becomes a synthetic test
		// when the module produced no tests to carry it.
		if len(event.Errors) > 0 && !w.hasTests(event.File) {
			w.put(event.File, w.jest.collectionError(relFile, strings.Join(event.Errors, "\n"), ""))
		}
	default:
		return
	}
	w.emit()
}

func (w *VitestStreamWriter) toTest(event vitestEvent, relFile string) Test {
	t := w.jest.toTest(jestAssertion{
		AncestorTitles:  event.Suite,
		Title:           event.Name,
		Status:          event.State,
		FailureMessages: event.Errors,
		Location:        jestLocation{Line: event.Line},
		Duration:        event.Duration,
	}, relFile, "", "", "")
	t.PackagePath = w.packagePath
	return t
}

func (w *VitestStreamWriter) hasTests(file string) bool {
	for _, key := range w.order {
		if strings.HasPrefix(key, file+"\x00") {
			return true
		}
	}
	return false
}

func (w *VitestStreamWriter) put(key string, t Test) {
	t.PackagePath = w.packagePath
	if _, ok := w.tests[key]; !ok {
		w.order = append(w.order, key)
	}
	w.tests[key] = t
}

func (w *VitestStreamWriter) emit() {
	flat := make([]Test, 0, len(w.order))
	for _, key := range w.order {
		flat = append(flat, w.tests[key])
	}
	children := buildSuiteTree(flat)
	propagateFailureStatus(children)
	w.sink.UpdatePackageProgress(w.packagePath, Vitest, Test{
		Name:        w.packagePath,
		PackagePath: w.packagePath,
		Framework:   Vitest,
		Pending:     true,
		Children:    children,
	})
}
