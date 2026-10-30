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

// relToWorkDir makes reported file paths relative to workDir. Pure-local
// duplicate of the helper in runners — parsers own path normalization now.
//
// Two path shapes arrive here. Without -trimpath ginkgo reports an absolute
// path, made relative against workDir. With -trimpath it reports an import path
// (github.com/org/repo/pkg/foo_test.go), which is neither absolute nor
// ..-prefixed, so it is stripped of the module prefix instead.
func (p *GinkgoJSON) relToWorkDir(filePath string) string {
	if filePath == "" || p.workDir == "" {
		return filePath
	}
	if !filepath.IsAbs(filePath) && !strings.HasPrefix(filePath, "..") {
		if p.modulePath != "" {
			if rel, ok := strings.CutPrefix(filePath, p.modulePath+"/"); ok {
				return rel
			}
		}
		return filePath
	}
	if rel, err := filepath.Rel(p.workDir, filePath); err == nil {
		return rel
	}
	return filePath
}

// GinkgoJSON parses Ginkgo --json-report output format.
type GinkgoJSON struct {
	workDir    string // used to make File paths relative to the project root
	modulePath string // go.mod module path, stripped from -trimpath locations
}

// NewGinkgoJSON creates a new Ginkgo JSON parser.
func NewGinkgoJSON(workDir string) *GinkgoJSON {
	p := &GinkgoJSON{workDir: workDir}
	if workDir != "" {
		p.modulePath, _ = GoModuleName(workDir)
	}
	return p
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
	ContainerHierarchyTexts    []string            `json:"ContainerHierarchyTexts"`
	ContainerHierarchyLabels   [][]string          `json:"ContainerHierarchyLabels"`
	LeafNodeText               string              `json:"LeafNodeText"`
	LeafNodeLabels             []string            `json:"LeafNodeLabels"`
	LeafNodeLocation           ginkgoLocation      `json:"LeafNodeLocation"`
	State                      string              `json:"State"`
	RunTime                    int64               `json:"RunTime"`
	NumAttempts                int                 `json:"NumAttempts"`
	ParallelProcess            int                 `json:"ParallelProcess"`
	ReportEntries              []ginkgoReportEntry `json:"ReportEntries"`
	Failure                    *ginkgoFailure      `json:"Failure"`
	CapturedStdOutOutput       string              `json:"CapturedStdOutOutput"`
	CapturedStdErrOutput       string              `json:"CapturedStdErrOutput"`
	CapturedGinkgoWriterOutput string              `json:"CapturedGinkgoWriterOutput"`
}

type ginkgoReportEntry struct {
	Name  string `json:"Name"`
	Value struct {
		AsJSON         string `json:"AsJSON"`
		Representation string `json:"Representation"`
	} `json:"Value"`
	Location ginkgoLocation `json:"Location"`
}

// ginkoSuiteReport represents a complete test suite in Ginkgo JSON format.
type ginkgoSuiteReport struct {
	SuitePath                  string             `json:"SuitePath"`
	SuiteDescription           string             `json:"SuiteDescription"`
	SuiteSucceeded             *bool              `json:"SuiteSucceeded"`
	SpecialSuiteFailureReasons []string           `json:"SpecialSuiteFailureReasons"`
	SpecReports                []ginkgoSpecReport `json:"SpecReports"`
}

// Parse reads Ginkgo JSON report and returns all test results (pass, fail, skip).
func (p *GinkgoJSON) Parse(output io.Reader) ([]Test, error) {
	var suiteReports []ginkgoSuiteReport
	if err := json.NewDecoder(output).Decode(&suiteReports); err != nil {
		return nil, fmt.Errorf("failed to parse Ginkgo JSON: %w", err)
	}

	var tests []Test
	for _, suite := range suiteReports {
		suiteHasFailedSpec := false
		for _, spec := range suite.SpecReports {
			test := p.specReportToTest(spec, suite)
			suiteHasFailedSpec = suiteHasFailedSpec || test.Failed
			tests = append(tests, test)
		}

		// Ginkgo records failures such as a suite timeout on the suite report,
		// even when every spec report is passed or skipped. Keep one explicit
		// failure for that verdict without double-counting ordinary spec failures.
		suiteFailed := len(suite.SpecialSuiteFailureReasons) > 0 ||
			suite.SuiteSucceeded != nil && !*suite.SuiteSucceeded
		if suiteFailed && !suiteHasFailedSpec {
			message := strings.Join(suite.SpecialSuiteFailureReasons, "\n")
			if message == "" {
				message = "Ginkgo suite failed"
			}
			timedOut := false
			for _, reason := range suite.SpecialSuiteFailureReasons {
				if reason == "Suite Timeout Elapsed" {
					timedOut = true
					break
				}
			}
			test := Test{
				Name:      "[suite failure]",
				Framework: Ginkgo,
				Package:   filepath.Base(suite.SuitePath),
				Failed:    true,
				TimedOut:  timedOut,
				Message:   message,
				Context: GinkgoContext{
					SuiteDescription: suite.SuiteDescription,
					SuitePath:        suite.SuitePath,
				},
			}
			if suite.SuiteDescription != "" {
				test.Suite = []string{suite.SuiteDescription}
			}
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
		NumAttempts:      spec.NumAttempts,
		ParallelProcess:  spec.ParallelProcess,
	}
	seenLabels := map[string]bool{}
	for _, labels := range append(spec.ContainerHierarchyLabels, spec.LeafNodeLabels) {
		for _, label := range labels {
			if !seenLabels[label] {
				ctx.Labels = append(ctx.Labels, label)
				seenLabels[label] = true
			}
		}
	}
	for _, entry := range spec.ReportEntries {
		ctx.ReportEntries = append(ctx.ReportEntries, GinkgoReportEntry{
			Name:           entry.Name,
			Value:          json.RawMessage(entry.Value.AsJSON),
			Representation: entry.Value.Representation,
			File:           p.relToWorkDir(entry.Location.FileName),
			Line:           entry.Location.LineNumber,
		})
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
	// Every ginkgo failure state (types.SpecStateFailureStates). "interrupted"
	// is a signal or another parallel process aborting; the JSON report does not
	// expose a suite timeout as its cause, so it is not marked TimedOut.
	case "failed", "timedout", "panicked", "aborted", "interrupted":
		test.Failed = true
		test.TimedOut = spec.State == "timedout"
		test.Aborted = spec.State == "aborted"
		p.applyFailure(&test, &ctx, spec)
	case "skipped", "pending":
		test.Skipped = true
	// SpecStateInvalid ("INVALID SPEC STATE") or a state a newer ginkgo adds:
	// fail it so it can never read as passed or drop out of the counts.
	default:
		test.Failed = true
		p.applyFailure(&test, &ctx, spec)
		unknown := fmt.Sprintf("unknown ginkgo spec state %q", spec.State)
		if test.Message != "" {
			unknown += ": " + test.Message
		}
		test.Message = unknown
	}

	test.Context = ctx
	return test
}

func (p *GinkgoJSON) applyFailure(test *Test, ctx *GinkgoContext, spec ginkgoSpecReport) {
	if spec.Failure == nil {
		return
	}
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
