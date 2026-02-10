package parsers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
)

// GoTestJSON parses go test -json output format (shared by go test and Ginkgo with --gojson-report).
type GoTestJSON struct {
	LocationMap map[string]TestLocation
}

// NewGoTestJSON creates a new Go test JSON parser.
// workDir is used to scan test files and build a location map for accurate file:line info.
func NewGoTestJSON(workDir string) *GoTestJSON {
	locations, _ := BuildTestLocationMap(workDir)
	return &GoTestJSON{LocationMap: locations}
}

// Name returns the parser name.
func (p *GoTestJSON) Name() string {
	return "go test json"
}

// goTestEvent represents a single event from go test -json output.
type goTestEvent struct {
	Time       string  `json:"Time"`
	Action     string  `json:"Action"`
	Package    string  `json:"Package"`
	ImportPath string  `json:"ImportPath"`
	Test       string  `json:"Test"`
	Elapsed    float64 `json:"Elapsed"`
	Output     string  `json:"Output"`
}

// getPackage returns the package name from either Package or ImportPath field.
func (e *goTestEvent) getPackage() string {
	if e.Package != "" {
		return e.Package
	}
	return e.ImportPath
}

// truncateFailure converts multi-line error messages to single line and truncates to maxLen.
func truncateFailure(message string, maxLen int) string {
	singleLine := strings.ReplaceAll(message, "\n", " ")
	singleLine = regexp.MustCompile(`\s+`).ReplaceAllString(singleLine, " ")
	singleLine = strings.TrimSpace(singleLine)
	if len(singleLine) > maxLen {
		return singleLine[:maxLen-3] + "..."
	}
	return singleLine
}

// ParseStream reads test output line-by-line and updates task progress.
// Returns (passCount, failCount, error).
func (p *GoTestJSON) ParseStream(output io.Reader, stdout io.Writer, t *task.Task) (int, int, error) {
	failuresByTest := make(map[string]string)
	testCount := 0
	passCount := 0
	failureCount := 0

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()

		// Write raw line to stdout buffer
		_, _ = stdout.Write([]byte(line + "\n"))

		// Try to parse as JSON
		var event goTestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Log unparseable line for debugging with reason
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 0 {
				truncated := truncateFailure(trimmed, 200)
				if strings.Contains(strings.ToUpper(trimmed), "FAIL") ||
					strings.Contains(strings.ToLower(trimmed), "error") ||
					strings.Contains(strings.ToLower(trimmed), "panic") {
					if t != nil {
						t.Warnf("Non-JSON output: %s", truncated)
					}
				} else {
					logger.Debugf("JSON unmarshal failed (%v): %s", err, truncated)
				}
			}
			continue
		}

		// Update task progress based on action
		switch event.Action {
		case "run":
			if event.Test != "" {
				testCount++
			}
		case "pass":
			if event.Test != "" {
				passCount++
				// Only log passing tests at verbose level
				if t != nil && logger.V(2).Enabled() {
					t.Warnf("✓ %s", event.Test)
				}
			}
		case "skip":
			if event.Test != "" {
				// Only log skipped tests at verbose level
				if t != nil && logger.V(1).Enabled() {
					t.Warnf("⊝ %s", event.Test)
				}
			}
		case "fail":
			if event.Test != "" {
				failureCount++
				if event.Output != "" {
					failuresByTest[event.Test] = event.Output
				}
				// Always log failing tests
				if t != nil && logger.V(1).Enabled() {
					if event.Output != "" {
						truncated := truncateFailure(event.Output, 200)
						t.Warnf("✗ %s: %s", event.Test, truncated)
					} else {
						t.Warnf("✗ %s", event.Test)
					}
				}
			}
		}

		// Update task with pretty formatting
		if t != nil && (testCount > 0 || failureCount > 0) {
			p.updateTaskStatus(t, passCount, failureCount, testCount)
		}
	}

	if err := scanner.Err(); err != nil {
		return passCount, failureCount, err
	}

	return passCount, failureCount, nil
}

// updateTaskStatus updates the task name with pretty-formatted pass/fail counts.
func (p *GoTestJSON) updateTaskStatus(t *task.Task, passed, failed, total int) {
	var status string
	switch {
	case failed > 0:
		status = fmt.Sprintf("tests (%d passed, %d failed, %d total)", passed, failed, total)
	case total > 0:
		status = fmt.Sprintf("tests (%d passed, %d total)", passed, total)
	default:
		status = "tests"
	}
	t.SetName(status)
}

// Parse reads test output and returns all test results (pass, fail, skip).
func (p *GoTestJSON) Parse(output io.Reader) ([]Test, error) {
	tests := make(map[string]*Test)
	testOutputs := make(map[string]*strings.Builder)
	buildErrors := make(map[string]*strings.Builder)
	scanner := bufio.NewScanner(output)

	for scanner.Scan() {
		line := scanner.Text()
		var event goTestEvent

		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("failed to parse go test json line: %w", err)
		}

		// Handle package-level events (no test name)
		if event.Test == "" {
			pkg := event.getPackage()
			switch event.Action {
			case "build-output":
				if _, exists := buildErrors[pkg]; !exists {
					buildErrors[pkg] = &strings.Builder{}
				}
				buildErrors[pkg].WriteString(event.Output)
			case "build-fail":
				testKey := pkg + "::build-failed"
				buildMsg := ""
				if builder, exists := buildErrors[pkg]; exists {
					buildMsg = builder.String()
					delete(buildErrors, pkg)
				}
				tests[testKey] = &Test{
					Name:      "Build Failed",
					Package:   pkg,
					Framework: GoTest,
					Failed:    true,
					Message:   buildMsg,
				}
			case "skip":
				testKey := pkg + "::package-skipped"
				tests[testKey] = &Test{
					Name:      "No test files",
					Package:   pkg,
					Framework: GoTest,
					Skipped:   true,
					Message:   "[no test files]",
				}
			}
			continue
		}

		testKey := event.Package + "::" + event.Test

		// Collect output for each test (skip header/footer lines)
		if event.Action == "output" && event.Output != "" {
			// Skip test header/footer lines
			trimmed := strings.TrimSpace(event.Output)
			if strings.HasPrefix(trimmed, "=== RUN") || strings.HasPrefix(trimmed, "--- PASS") ||
				strings.HasPrefix(trimmed, "--- FAIL") || strings.HasPrefix(trimmed, "--- SKIP") {
				continue
			}
			if _, exists := testOutputs[testKey]; !exists {
				testOutputs[testKey] = &strings.Builder{}
			}
			testOutputs[testKey].WriteString(event.Output)
		}

		switch event.Action {
		case "run":
			// Initialize test entry on first run
			if _, exists := tests[testKey]; !exists {
				tests[testKey] = &Test{
					Name:      event.Test,
					Package:   event.Package,
					Framework: GoTest,
				}
			}

		case "pass":
			// Mark test as passed
			duration := time.Duration(event.Elapsed * float64(time.Second))
			if test, exists := tests[testKey]; exists {
				test.Failed = false
				test.Skipped = false
				test.Duration = duration
			} else {
				tests[testKey] = &Test{
					Name:      event.Test,
					Package:   event.Package,
					Framework: GoTest,
					Failed:    false,
					Skipped:   false,
					Duration:  duration,
				}
			}

		case "fail":
			// Parse failure details and mark as failed
			test := p.parseTestOutput(event.Test, event.Package, event.Output)
			test.Duration = time.Duration(event.Elapsed * float64(time.Second))
			tests[testKey] = test

		case "skip":
			// Mark test as skipped
			duration := time.Duration(event.Elapsed * float64(time.Second))
			if test, exists := tests[testKey]; exists {
				test.Skipped = true
				test.Failed = false
				test.Duration = duration
			} else {
				tests[testKey] = &Test{
					Name:      event.Test,
					Package:   event.Package,
					Framework: GoTest,
					Skipped:   true,
					Duration:  duration,
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice and attach individual test output and locations
	results := make([]Test, 0, len(tests))
	for testKey, test := range tests {
		// Skip Ginkgo bootstrap tests (tests that only call RunSpecs)
		if p.isGinkgoBootstrap(test.Name) {
			continue
		}
		// Set test-specific stdout from collected output events
		if output, exists := testOutputs[testKey]; exists {
			test.Stdout = strings.TrimSpace(output.String())
		}
		// Look up file:line from AST-based location map
		p.applyLocationFromMap(test)
		results = append(results, *test)
	}

	return results, nil
}

// isGinkgoBootstrap checks if the test is a Ginkgo bootstrap function.
func (p *GoTestJSON) isGinkgoBootstrap(testName string) bool {
	if p.LocationMap == nil {
		return false
	}
	if loc, ok := p.LocationMap[testName]; ok {
		return loc.IsGinkgoBootstrap
	}
	return false
}

// applyLocationFromMap sets File and Line from the location map if available.
// For subtests like "TestFoo/subtest", it looks up the parent test "TestFoo".
func (p *GoTestJSON) applyLocationFromMap(test *Test) {
	if p.LocationMap == nil {
		return
	}

	// Try exact match first
	testName := test.Name
	if loc, ok := p.LocationMap[testName]; ok {
		test.File = loc.File
		test.Line = loc.Line
		return
	}

	// For subtests, extract parent test name (before first /)
	if idx := strings.Index(testName, "/"); idx > 0 {
		parentName := testName[:idx]
		if loc, ok := p.LocationMap[parentName]; ok {
			test.File = loc.File
			test.Line = loc.Line
		}
	}
}

// parseTestOutput extracts failure message from go test output.
// File and Line are set later via applyLocationFromMap using AST data.
func (p *GoTestJSON) parseTestOutput(testName, pkgName, output string) *Test {
	lines := strings.Split(output, "\n")
	var message string

	for _, logLine := range lines {
		if strings.TrimSpace(logLine) != "" && !strings.HasPrefix(logLine, "---") {
			message += logLine + "\n"
		}
	}

	return &Test{
		Name:      testName,
		Package:   pkgName,
		Message:   strings.TrimSpace(message),
		Framework: GoTest,
		Failed:    true,
	}
}
