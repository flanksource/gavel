package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
)

// PlaywrightJSON parses Playwright's `--reporter=json` output.
type PlaywrightJSON struct {
	workDir string
}

func NewPlaywrightJSON(workDir string) *PlaywrightJSON {
	return &PlaywrightJSON{workDir: workDir}
}

func (p *PlaywrightJSON) Name() string { return "playwright json" }

func (p *PlaywrightJSON) ParseStream(output io.Reader, stdout io.Writer, t *task.Task) (int, int, error) {
	return 0, 0, nil
}

type pwError struct {
	Message string `json:"message"`
	Stack   string `json:"stack"`
}

type pwResult struct {
	Status   string    `json:"status"`
	Duration float64   `json:"duration"`
	Error    *pwError  `json:"error"`
	Errors   []pwError `json:"errors"`
}

type pwTest struct {
	ProjectName    string     `json:"projectName"`
	ExpectedStatus string     `json:"expectedStatus"`
	Results        []pwResult `json:"results"`
}

type pwSpec struct {
	Title string   `json:"title"`
	File  string   `json:"file"`
	Line  int      `json:"line"`
	Ok    bool     `json:"ok"`
	Tests []pwTest `json:"tests"`
}

type pwSuite struct {
	Title  string    `json:"title"`
	File   string    `json:"file"`
	Line   int       `json:"line"`
	Specs  []pwSpec  `json:"specs"`
	Suites []pwSuite `json:"suites"`
}

type pwReport struct {
	Suites []pwSuite `json:"suites"`
}

func (p *PlaywrightJSON) Parse(output io.Reader) ([]Test, error) {
	data, err := io.ReadAll(output)
	if err != nil {
		return nil, fmt.Errorf("read playwright report: %w", err)
	}
	var report pwReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse playwright json: %w", err)
	}

	var out []Test
	for _, s := range report.Suites {
		out = append(out, p.walk(s, nil)...)
	}
	return out, nil
}

func (p *PlaywrightJSON) walk(s pwSuite, parents []string) []Test {
	path := append(append([]string{}, parents...), s.Title)
	var out []Test
	for _, spec := range s.Specs {
		out = append(out, p.toTest(spec, path))
	}
	for _, child := range s.Suites {
		out = append(out, p.walk(child, path)...)
	}
	return out
}

func (p *PlaywrightJSON) toTest(spec pwSpec, suite []string) Test {
	t := Test{
		Name:      spec.Title,
		Suite:     append([]string{}, suite...),
		File:      p.relPath(spec.File),
		Line:      spec.Line,
		Framework: Playwright,
	}

	var finalResult *pwResult
	var totalDuration float64
	for i := range spec.Tests {
		for j := range spec.Tests[i].Results {
			r := spec.Tests[i].Results[j]
			totalDuration += r.Duration
			rr := r
			finalResult = &rr
		}
	}
	t.Duration = time.Duration(totalDuration) * time.Millisecond

	if finalResult == nil {
		t.Skipped = true
		return t
	}

	switch finalResult.Status {
	case "passed":
		t.Passed = true
	case "failed", "timedOut", "interrupted":
		t.Failed = true
	case "skipped":
		t.Skipped = true
	default:
		t.Failed = true
	}

	if finalResult.Error != nil {
		t.Message = StripANSI(strings.TrimSpace(finalResult.Error.Message))
		if finalResult.Error.Stack != "" {
			t.Stderr = StripANSI(finalResult.Error.Stack)
		}
	} else if len(finalResult.Errors) > 0 {
		msgs := make([]string, 0, len(finalResult.Errors))
		for _, e := range finalResult.Errors {
			msgs = append(msgs, e.Message)
		}
		t.Message = StripANSI(strings.TrimSpace(strings.Join(msgs, "\n")))
	}
	return t
}

func (p *PlaywrightJSON) relPath(filePath string) string {
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
