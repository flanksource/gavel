package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/task"
)

// GinkgoJSON parses Ginkgo --json-report output format.
type GinkgoJSON struct{}

// NewGinkgoJSON creates a new Ginkgo JSON parser.
func NewGinkgoJSON() *GinkgoJSON {
	return &GinkgoJSON{}
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
	ContainerHierarchyTexts []string       `json:"ContainerHierarchyTexts"`
	LeafNodeText            string         `json:"LeafNodeText"`
	LeafNodeLocation        ginkgoLocation `json:"LeafNodeLocation"`
	State                   string         `json:"State"`
	RunTime                 int64          `json:"RunTime"`
	Failure                 *ginkgoFailure `json:"Failure"`
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

	test := Test{
		Name:      spec.LeafNodeText,
		Suite:     suiteHierarchy,
		File:      spec.LeafNodeLocation.FileName,
		Line:      spec.LeafNodeLocation.LineNumber,
		Framework: Ginkgo,
		Duration:  time.Duration(spec.RunTime),
		Package:   packagePath,
	}

	// Set pass/fail/skip status
	switch spec.State {
	case "passed":
		test.Passed = true
	case "failed":
		test.Failed = true
		if spec.Failure != nil {
			test.Message = spec.Failure.Message
			// Keep LeafNodeLocation for File/Line - it's correct for table-driven tests
			// where each Entry has a unique location but shares the same Failure.Location
		}
	case "skipped":
		test.Skipped = true
	}

	return test
}
