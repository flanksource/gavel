package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
)

// JestJSON parses the JSON reporter output emitted by Jest (`jest --json`)
// and by Vitest (`vitest run --reporter=json`). The two tools share a
// common schema, so one parser services both; the framework label is fixed
// at construction time.
type JestJSON struct {
	workDir   string
	framework Framework
}

// NewJestJSON returns a parser that stamps tests with the given framework.
// If framework is empty, it defaults to Jest.
func NewJestJSON(workDir string, framework Framework) *JestJSON {
	if framework == "" {
		framework = Jest
	}
	return &JestJSON{workDir: workDir, framework: framework}
}

func (p *JestJSON) Name() string {
	return string(p.framework) + " json"
}

func (p *JestJSON) ParseStream(output io.Reader, stdout io.Writer, t *task.Task) (int, int, error) {
	return 0, 0, nil
}

type jestLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type jestAssertion struct {
	AncestorTitles  []string     `json:"ancestorTitles"`
	Title           string       `json:"title"`
	FullName        string       `json:"fullName"`
	Status          string       `json:"status"`
	FailureMessages []string     `json:"failureMessages"`
	Location        jestLocation `json:"location"`
	Duration        float64      `json:"duration"`
	// Vitest 3+ extension: retry bookkeeping. Kept optional so Jest payloads
	// (which never set it) still decode cleanly.
	Retries      int      `json:"retries,omitempty"`
	RetryReasons []string `json:"retryReasons,omitempty"`
}

// jestConsoleEntry captures a single console.log / .error line emitted by
// the test. Vitest populates this when `console: true` (the default).
type jestConsoleEntry struct {
	Type    string `json:"type"` // "log" | "info" | "warn" | "error"
	Message string `json:"message"`
	Origin  string `json:"origin,omitempty"`
}

type jestFileResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	// Vitest sometimes emits endTime as a fractional millisecond
	// (e.g. 1777008361916.414), so both fields are float64 to accommodate
	// that. Jest always emits whole numbers; the extra precision is
	// ignored downstream.
	StartTime        float64            `json:"startTime,omitempty"`
	EndTime          float64            `json:"endTime,omitempty"`
	Message          string             `json:"message"`
	AssertionResults []jestAssertion    `json:"assertionResults"`
	Console          []jestConsoleEntry `json:"console,omitempty"`
}

type jestReport struct {
	TestResults []jestFileResult `json:"testResults"`
}

// vitestStackFrameRe matches a Vitest / V8 stack frame line. Any whitespace
// or bracket-style prefix (`at `, `❯ `, `(`) is fine — we just need a path
// ending in a source-file extension followed by `:line:col`. The path may
// be absolute (`/repo/...`), relative (`src/...`, `./src/...`), Windows
// (`C:\...`), or a `node:` internal module.
//
// Examples handled:
//
//	    at /repo/src/sum.test.ts:8:17
//	    at Object.<anonymous> (src/sum.test.ts:8:17)
//	    ❯ src/sum.test.ts:8:17
//	    at node:internal/timers:42:17
var vitestStackFrameRe = regexp.MustCompile(`([^\s()]+\.[A-Za-z]{1,4}):(\d+):(\d+)`)

// extractStackLocation pulls the first user-code file:line out of a Vitest
// / Jest failureMessage. Returns ("", 0) when no frame resembles a real
// file path. Frames inside node_modules or node: internals are skipped so
// we land on the test file rather than framework internals.
func extractStackLocation(message string) (string, int) {
	if message == "" {
		return "", 0
	}
	matches := vitestStackFrameRe.FindAllStringSubmatch(message, -1)
	for _, m := range matches {
		file := m[1]
		if file == "" {
			continue
		}
		if strings.Contains(file, "node_modules") || strings.HasPrefix(file, "node:") {
			continue
		}
		var line int
		fmt.Sscanf(m[2], "%d", &line)
		return file, line
	}
	return "", 0
}

// formatConsoleEntries renders a file's console output as a compact,
// non-ANSI block that slots into Test.Stdout. Each entry gets a prefix so
// warnings/errors stand out in the UI's AnsiHtml renderer.
func formatConsoleEntries(entries []jestConsoleEntry) string {
	if len(entries) == 0 {
		return ""
	}
	var b strings.Builder
	for _, e := range entries {
		tag := strings.ToUpper(e.Type)
		if tag == "" {
			tag = "LOG"
		}
		b.WriteString("[")
		b.WriteString(tag)
		b.WriteString("] ")
		if e.Origin != "" {
			b.WriteString("(")
			b.WriteString(e.Origin)
			b.WriteString(") ")
		}
		b.WriteString(strings.TrimRight(e.Message, "\n"))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (p *JestJSON) Parse(output io.Reader) ([]Test, error) {
	data, err := io.ReadAll(output)
	if err != nil {
		return nil, fmt.Errorf("read %s report: %w", p.framework, err)
	}
	var report jestReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse %s json: %w", p.framework, err)
	}

	var tests []Test
	for _, file := range report.TestResults {
		relFile := p.relPath(file.Name)
		consoleText := formatConsoleEntries(file.Console)
		setupMessage := StripANSI(strings.TrimSpace(file.Message))

		// File collection / setup errors (e.g. syntax error in the spec
		// file, failed before-all hook) surface with an empty
		// AssertionResults slice and a non-empty file.Message. Emit a
		// synthetic failing test so the UI has somewhere to show the
		// error instead of silently swallowing it.
		if len(file.AssertionResults) == 0 && file.Status == "failed" {
			tests = append(tests, Test{
				Name:      filepath.Base(relFile),
				File:      relFile,
				Failed:    true,
				Framework: p.framework,
				Message:   setupMessage,
				Stderr:    strings.TrimSpace(file.Message), // keep ANSI for AnsiHtml render
				Stdout:    consoleText,
			})
			continue
		}
		for _, a := range file.AssertionResults {
			tests = append(tests, p.toTest(a, relFile, setupMessage, consoleText, file.Message))
		}
	}
	return tests, nil
}

func (p *JestJSON) toTest(a jestAssertion, relFile, setupMessage, consoleText, rawFileMessage string) Test {
	t := Test{
		Name:      a.Title,
		Suite:     append([]string{}, a.AncestorTitles...),
		File:      relFile,
		Line:      a.Location.Line,
		// Multiply in float64 space so sub-millisecond durations (common
		// in Vitest output, e.g. 0.6406...ms) survive as nanoseconds
		// rather than being truncated to 0.
		Duration:  time.Duration(a.Duration * float64(time.Millisecond)),
		Framework: p.framework,
	}
	switch a.Status {
	case "passed":
		t.Passed = true
	case "failed":
		t.Failed = true
	case "pending", "skipped", "todo", "disabled":
		t.Skipped = true
	}
	if len(a.FailureMessages) > 0 {
		// failureMessages arrive ANSI-coloured from Vitest — keep the
		// coloured copy on Stderr so the UI's AnsiHtml renderer can show
		// it, and mirror a plain-text copy into Message so the inline
		// Error section and CLI task label stay readable.
		joined := strings.TrimSpace(strings.Join(a.FailureMessages, "\n"))
		t.Stderr = joined
		t.Message = StripANSI(joined)

		// Location.Line from Vitest/Jest points at the `it(...)` call;
		// the actual assertion failure lives deeper in the stack. Prefer
		// the first user-code frame when we can find one — it's what the
		// developer actually wants to jump to.
		if file, line := extractStackLocation(joined); file != "" {
			// Normalize against workDir the same way file names are.
			t.File = p.relPath(file)
			t.Line = line
		}
	}
	if a.Retries > 0 && t.Message != "" {
		t.Message = fmt.Sprintf("after %d retr%s:\n%s", a.Retries,
			pluralize(a.Retries, "y", "ies"), t.Message)
	}
	if consoleText != "" {
		t.Stdout = consoleText
	}
	// File-level collection error still bubbles through as context on
	// every failing assertion, so the user can see what went wrong at
	// import/setup time without losing the per-assertion detail.
	if t.Failed && setupMessage != "" && !strings.Contains(t.Message, setupMessage) {
		t.Message = setupMessage + "\n\n" + t.Message
		if t.Stderr == "" {
			t.Stderr = strings.TrimSpace(rawFileMessage)
		}
	}
	return t
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func (p *JestJSON) relPath(filePath string) string {
	if filePath == "" || p.workDir == "" {
		return filePath
	}
	if !filepath.IsAbs(filePath) && !strings.HasPrefix(filePath, "..") {
		return filePath
	}
	if rel, err := filepath.Rel(p.workDir, filePath); err == nil {
		return rel
	}
	return filePath
}
