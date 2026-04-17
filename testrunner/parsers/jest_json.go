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
}

type jestFileResult struct {
	Name             string          `json:"name"`
	Status           string          `json:"status"`
	Message          string          `json:"message"`
	AssertionResults []jestAssertion `json:"assertionResults"`
}

type jestReport struct {
	TestResults []jestFileResult `json:"testResults"`
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
		if len(file.AssertionResults) == 0 && file.Status == "failed" {
			tests = append(tests, Test{
				Name:      filepath.Base(relFile),
				File:      relFile,
				Failed:    true,
				Framework: p.framework,
				Message:   StripANSI(strings.TrimSpace(file.Message)),
			})
			continue
		}
		for _, a := range file.AssertionResults {
			tests = append(tests, p.toTest(a, relFile))
		}
	}
	return tests, nil
}

func (p *JestJSON) toTest(a jestAssertion, relFile string) Test {
	t := Test{
		Name:      a.Title,
		Suite:     append([]string{}, a.AncestorTitles...),
		File:      relFile,
		Line:      a.Location.Line,
		Duration:  time.Duration(a.Duration) * time.Millisecond,
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
		t.Message = StripANSI(strings.TrimSpace(strings.Join(a.FailureMessages, "\n")))
	}
	return t
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
