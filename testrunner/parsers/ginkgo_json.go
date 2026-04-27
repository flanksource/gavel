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

// relToWorkDir makes absolute file paths relative to workDir. Pure-local
// duplicate of the helper in runners — parsers own path normalization now.
func (p *GinkgoJSON) relToWorkDir(filePath string) string {
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

// GinkgoJSON parses Ginkgo --json-report output format.
type GinkgoJSON struct {
	workDir string // used to make File paths relative to the project root
}

// NewGinkgoJSON creates a new Ginkgo JSON parser.
func NewGinkgoJSON(workDir string) *GinkgoJSON {
	return &GinkgoJSON{workDir: workDir}
}

// Name returns the parser name.
func (p *GinkgoJSON) Name() string {
	return "ginkgo json"
}

// ParseStream is not used for Ginkgo JSON format (reports are generated at the end).
// Returns (passCount, failCount, error). This is a no-op for Ginkgo.
func (p *GinkgoJSON) ParseStream(output io.Reader, stdout io.Writer, t *task.Task) (int, int, error) {
	// Ginkgo JSON report is not streamed; it's written as a single file at the end.
	// Real-time progress updates are not available for Ginkgo JSON format.
	return 0, 0, nil
}

// ginkgoLocation represents a file location in Ginkgo JSON format.
type ginkgoLocation struct {
	FileName   string `json:"FileName"`
	LineNumber int    `json:"LineNumber"`
}

// ginkgoFailure represents a failure in Ginkgo JSON format.
type ginkgoFailure struct {
	Message  string         `json:"Message"`
	Location ginkgoLocation `json:"Location"`
}

// ginkgoSpecReport represents a single test spec in Ginkgo JSON format.
type ginkgoSpecReport struct {
	ContainerHierarchyTexts    []string       `json:"ContainerHierarchyTexts"`
	LeafNodeText               string         `json:"LeafNodeText"`
	LeafNodeLocation           ginkgoLocation `json:"LeafNodeLocation"`
	State                      string         `json:"State"`
	RunTime                    int64          `json:"RunTime"`
	Failure                    *ginkgoFailure `json:"Failure"`
	CapturedStdOutOutput       string         `json:"CapturedStdOutOutput"`
	CapturedStdErrOutput       string         `json:"CapturedStdErrOutput"`
	CapturedGinkgoWriterOutput string         `json:"CapturedGinkgoWriterOutput"`
}

// ginkoSuiteReport represents a complete test suite in Ginkgo JSON format.
type ginkgoSuiteReport struct {
	SuitePath        string             `json:"SuitePath"`
	SuiteDescription string             `json:"SuiteDescription"`
	SpecReports      []ginkgoSpecReport `json:"SpecReports"`
}

// Parse reads Ginkgo JSON report and returns all test results (pass, fail, skip).
func (p *GinkgoJSON) Parse(output io.Reader) ([]Test, error) {
	var suiteReports []ginkgoSuiteReport
	if err := json.NewDecoder(output).Decode(&suiteReports); err != nil {
		return nil, fmt.Errorf("failed to parse Ginkgo JSON: %w", err)
	}

	var tests []Test
	for _, suite := range suiteReports {
		for _, spec := range suite.SpecReports {
			test := p.specReportToTest(spec, suite)
			tests = append(tests, test)
		}
	}

	return tests, nil
}

// specReportToTest converts a Ginkgo SpecReport to a Test struct.
func (p *GinkgoJSON) specReportToTest(spec ginkgoSpecReport, suite ginkgoSuiteReport) Test {
	// Get suite hierarchy from containers
	var suiteHierarchy []string
	if len(spec.ContainerHierarchyTexts) > 0 {
		suiteHierarchy = spec.ContainerHierarchyTexts
	} else if suite.SuiteDescription != "" {
		suiteHierarchy = []string{suite.SuiteDescription}
	}

	// Extract package name from suite path (directory name)
	packagePath := filepath.Base(suite.SuitePath)

	ctx := GinkgoContext{
		SuiteDescription: suite.SuiteDescription,
		SuitePath:        suite.SuitePath,
	}

	stdout := spec.CapturedStdOutOutput
	if spec.CapturedGinkgoWriterOutput != "" {
		if stdout != "" {
			stdout += "\n"
		}
		stdout += spec.CapturedGinkgoWriterOutput
	}

	// Ginkgo emits BeforeSuite / AfterSuite / DeferCleanup failures with no
	// LeafNodeText. Fall back to a synthetic name so the UI tree row and
	// detail heading aren't blank.
	leafName := spec.LeafNodeText
	if leafName == "" {
		switch {
		case suite.SuiteDescription != "":
			leafName = "[" + suite.SuiteDescription + " setup/teardown]"
		case spec.LeafNodeLocation.FileName != "":
			leafName = "[setup at " + p.relToWorkDir(spec.LeafNodeLocation.FileName) + "]"
		default:
			leafName = "[suite setup]"
		}
	}

	test := Test{
		Name:      leafName,
		Suite:     suiteHierarchy,
		File:      p.relToWorkDir(spec.LeafNodeLocation.FileName),
		Line:      spec.LeafNodeLocation.LineNumber,
		Framework: Ginkgo,
		Duration:  time.Duration(spec.RunTime),
		Package:   packagePath,
		Stdout:    stdout,
		Stderr:    spec.CapturedStdErrOutput,
	}

	switch spec.State {
	case "passed":
		test.Passed = true
	case "failed":
		test.Failed = true
		if spec.Failure != nil {
			test.Message = spec.Failure.Message
			test.FailureDetail = ParseFailureDetail(spec.Failure.Message)
			failLoc := fmt.Sprintf("%s:%d", p.relToWorkDir(spec.Failure.Location.FileName), spec.Failure.Location.LineNumber)
			testLoc := fmt.Sprintf("%s:%d", p.relToWorkDir(spec.LeafNodeLocation.FileName), spec.LeafNodeLocation.LineNumber)
			if failLoc != testLoc {
				ctx.FailureLocation = failLoc
			}
			if test.FailureDetail != nil && test.FailureDetail.Location == "" && ctx.FailureLocation != "" {
				test.FailureDetail.Location = ctx.FailureLocation
			}
		}
	case "skipped":
		test.Skipped = true
	}

	test.Context = ctx
	return test
}
