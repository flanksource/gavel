package parsers

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/commons/logger"
	"github.com/samber/lo"
)

// Framework represents a test framework type.
type Framework string

const (
	GoTest Framework = "go test"
	Ginkgo Framework = "ginkgo"
)

// String returns the string representation of the framework.
func (f Framework) String() string {
	return string(f)
}

// Test represents a single test failure.
type Test struct {
	Name        string        `json:"name,omitempty"`
	Package     string        `json:"package,omitempty"`
	PackagePath string        `json:"package_path,omitempty"` // Relative path to the package (e.g., "./pkg/testrunner")
	Suite       []string      `json:"suite,omitempty"`        // Hierarchical suite path (e.g., ["Outer Describe", "Inner Context"])
	Message     string        `json:"message,omitempty"`
	File        string        `json:"file,omitempty"`
	Line        int           `json:"line,omitempty"`
	Framework   Framework     `json:"framework,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	Skipped     bool          `json:"skipped,omitempty"`
	Failed      bool          `json:"failed,omitempty"`
	Passed      bool          `json:"passed,omitempty"`
	Stdout      string        `json:"stdout,omitempty"`
	Stderr      string        `json:"stderr,omitempty"`
	Children    Tests         `json:"children,omitempty"`
	Summary     *TestSummary  `json:"summary,omitempty"`
}

func (t Test) IsFolder() bool {
	return !t.Skipped && !t.Failed && !t.Passed
}

func (t Test) IsEmpty() bool {
	if !t.IsFolder() {
		return false
	}
	return len(t.Children) == 0
}

func (t Test) SuitePath() string {
	return strings.Join(t.Suite, " > ")
}

func (t Test) FullName() string {
	if len(t.Suite) > 0 {
		return fmt.Sprintf("%s > %s", strings.Join(t.Suite, " > "), t.Name)
	}
	return t.Name
}

func (t Test) GetChildren() []api.TreeNode {
	var children []api.TreeNode
	for _, child := range t.Children {
		children = append(children, child)
	}
	return children
}

func (t Test) Pretty() api.Text {
	s := clicky.Text("")
	textStyle := "bold"
	if t.Skipped {
		s = s.Append(icons.Skip, "text-orange-500")
		textStyle = "text-yellow-500"
	} else if t.Failed {
		s = s.Append(icons.Fail, "text-red-500")
		textStyle = "text-red-500"
	} else {
		s = s.Add(icons.Pass)
	}

	// Add file:line if present, or just :line if file is cleared (for child nodes)
	if t.File != "" && t.Line > 0 {
		s = s.Append(fmt.Sprintf("%s:%d", t.File, t.Line), "text-muted wrap-space")
	} else if t.Line > 0 {
		s = s.Append(fmt.Sprintf(":%d", t.Line), "text-muted wrap-space")
	}
	s = s.Append(t.Name, "bold wrap-space", textStyle)

	// Add duration if non-zero
	if t.Duration > 0 {
		s = s.Append(fmt.Sprintf("(%s)", t.Duration), "text-muted")
	}

	// Add message if present
	if t.Message != "" {
		s = s.Space().Append(t.Message, textStyle)
	}
	if t.Failed && t.Stdout != "" {
		s = s.NewLine().Append(t.Stdout, "text-red-500 max-lines-10")
	}
	if t.Stderr != "" {
		s = s.NewLine().Append(t.Stderr, "text-red-500 max-lines-20")
	}

	return s
}

// RerunCommand returns the command to re-run this specific test.
func (t Test) RerunCommand() string {
	switch t.Framework {
	case GoTest:
		pkgPath := t.PackagePath
		if pkgPath == "" && t.File != "" {
			pkgPath = "./" + filepath.Dir(t.File)
		}
		return fmt.Sprintf("go test -run ^%s$ %s", t.Name, pkgPath)
	case Ginkgo:
		return fmt.Sprintf(`ginkgo --focus="%s"`, t.Name)
	default:
		return ""
	}
}

// PrettyTODO returns the markdown body for a TODO file (excluding frontmatter).
func (t Test) PrettyTODO() api.Text {
	text := clicky.Text(fmt.Sprintf("# TODO: Fix Test - %s", t.Name)).NewLine().NewLine()

	// Test Information section
	text = text.Append("## Test Information", "").NewLine().NewLine()
	text = text.Add(api.KeyValuePair{Key: "Test Name", Value: t.Name}).NewLine()
	if len(t.Suite) > 0 {
		text = text.Add(api.KeyValuePair{Key: "Suite", Value: strings.Join(t.Suite, " > ")}).NewLine()
	}
	text = text.Add(api.KeyValuePair{Key: "Package", Value: t.Package}).NewLine()
	text = text.Add(api.KeyValuePair{Key: "File", Value: fmt.Sprintf("%s:%d", t.File, t.Line)}).NewLine()

	// Re-run Command section
	text = text.NewLine().Append("## Re-run Command", "").NewLine().NewLine()
	text = text.Add(api.Code{Content: t.RerunCommand(), Language: "bash"}).NewLine()

	// Latest Failure section
	text = text.NewLine().Append("## Latest Failure", "").NewLine().NewLine()
	if t.Message != "" {
		text = text.Add(api.Code{Content: t.Message}).NewLine()
	}

	// Stdout if available
	if t.Stdout != "" {
		text = text.NewLine().Append("## Test Output (stdout)", "").NewLine().NewLine()
		text = text.Add(api.Code{Content: truncateOutput(t.Stdout, 5000)}).NewLine()
	}

	// Stderr if available
	if t.Stderr != "" {
		text = text.NewLine().Append("## Test Errors (stderr)", "").NewLine().NewLine()
		text = text.Add(api.Code{Content: truncateOutput(t.Stderr, 5000)}).NewLine()
	}

	// Failure History section placeholder
	text = text.NewLine().Append("## Failure History", "").NewLine()

	return text
}

func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n\n... (output truncated) ..."
}

func (tr Test) Sum() TestSummary {

	if tr.Summary != nil {
		return *tr.Summary
	}
	var summary = TestSummary{
		Duration: tr.Duration,
		Total:    1,
	}
	if tr.Failed {
		summary.Failed = 1
	} else if tr.Skipped {
		summary.Skipped = 1
	} else if tr.Passed {
		summary.Passed = 1
	} else {
		summary.Total = 0
	}

	for _, child := range tr.Children {
		childSummary := child.Sum()
		summary.Total += childSummary.Total
		summary.Passed += childSummary.Passed
		summary.Failed += childSummary.Failed
		summary.Skipped += childSummary.Skipped
		summary.Duration += childSummary.Duration
	}

	return summary
}

type Tests []Test

func (tests Tests) Prune() Tests {
	var pruned Tests
	for _, test := range tests {
		if !logger.V(1).Enabled() {
			pruned = append(pruned, test.Filter(TestFilter{
				ExcludePassed: true,
				ExcludeSpecs:  true,
				SlowTests:     lo.ToPtr(10 * time.Second),
			}))
		} else {
			pruned = append(pruned, test)
		}
	}
	return pruned
}

func (tests Tests) Sum() TestSummary {
	var summary TestSummary
	for _, test := range tests {
		summary = summary.Add(test.Sum())
	}
	return summary
}

type TestSuiteResults []TestResults

func (tsr TestSuiteResults) All() Tests {
	var all Tests
	for _, tr := range tsr {
		all = append(all, tr.Tests...)
	}
	all.Sort()
	return all
}

func (tsr TestSuiteResults) Failed() Tests {
	var all Tests
	for _, tr := range tsr {
		for _, test := range tr.Tests {
			if test.Failed {
				all = append(all, test)
			}
		}
	}
	all.Sort()
	return all
}

type TestResults struct {
	Command   string    `json:"command,omitempty"`
	Framework Framework `json:"framework,omitempty"`
	Tests     Tests     `json:"tests,omitempty"`
	Stdout    string    `json:"stdout,omitempty"`
	Stderr    string    `json:"stderr,omitempty"`
	ExitCode  int       `json:"exit_code,omitempty"`
}

func (tr TestResults) Pretty() api.Text {
	s := clicky.Text(tr.Command, "wrap-space")
	s = s.Add(tr.Sum().Pretty())
	return s
}

// SortTests sorts tests by Suite (hierarchically) then by Name.
func (tests Tests) Sort() {
	sort.Slice(tests, func(i, j int) bool {
		// Compare suites first
		suiteI := strings.Join(tests[i].Suite, "/")
		suiteJ := strings.Join(tests[j].Suite, "/")
		if suiteI != suiteJ {
			return suiteI < suiteJ
		}

		if tests[i].Passed && !tests[j].Passed {
			return false
		}
		if !tests[i].Passed && tests[j].Passed {
			return true
		}

		// If suites are equal, compare by name
		return tests[i].Name < tests[j].Name
	})
}

// SortTestResults sorts a slice of TestResults by the first test's Suite and Name.
func (results TestSuiteResults) Sort() {
	sort.Slice(results, func(i, j int) bool {
		// Handle empty test lists
		if len(results[i].Tests) == 0 && len(results[j].Tests) == 0 {
			return results[i].Command < results[j].Command
		}
		if len(results[i].Tests) == 0 {
			return true
		}
		if len(results[j].Tests) == 0 {
			return false
		}

		// Compare by first test's suite
		suiteI := strings.Join(results[i].Tests[0].Suite, "/")
		suiteJ := strings.Join(results[j].Tests[0].Suite, "/")
		if suiteI != suiteJ {
			return suiteI < suiteJ
		}
		// If suites are equal, compare by first test's name
		return results[i].Tests[0].Name < results[j].Tests[0].Name
	})
}

type TestSummary struct {
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
}

func (s TestSummary) Pretty() api.Text {
	t := clicky.Text("")

	if s.Passed > 0 {
		t = t.Add(clicky.KeyValue(" passed", s.Passed, "text-green-500")).Append(" ")
	}
	if s.Failed > 0 {
		t = t.Add(clicky.KeyValue(" failed", s.Failed, "text-red-500")).Append(" ")
	}
	if s.Skipped > 0 {
		t = t.Add(clicky.KeyValue(" skipped", s.Skipped, "text-yellow-500")).Append(" ")
	}
	t = t.Add(clicky.KeyValue(" total", s.Total, "muted"))
	t = t.Add(clicky.KeyValue(" duration", s.Duration, "muted"))
	return t
}

func (tr TestSummary) Add(other TestSummary) TestSummary {
	return TestSummary{
		Total:    tr.Total + other.Total,
		Passed:   tr.Passed + other.Passed,
		Failed:   tr.Failed + other.Failed,
		Skipped:  tr.Skipped + other.Skipped,
		Duration: tr.Duration + other.Duration,
	}
}

func (tr TestResults) Sum() TestSummary {
	var summary TestSummary
	for _, test := range tr.Tests {
		summary = summary.Add(test.Sum())
	}
	return summary
}

type TestFilter struct {
	ExcludePassed bool
	// Exclude It() entries or individial entries in data driven tests (unless they fail)
	ExcludeSpecs bool

	SlowTests *time.Duration
}

func (t Test) IsSpec() bool {
	return !t.IsFolder()
}

func (f TestFilter) Matches(t Test) bool {
	if f.ExcludePassed && t.Passed {
		return false
	}
	if f.SlowTests != nil && t.Duration <= *f.SlowTests {
		return false
	}
	return true
}

func (tr Test) Filter(filter TestFilter) Test {
	out := Test{
		Name:        tr.Name,
		Package:     tr.Package,
		PackagePath: tr.PackagePath,
		Suite:       tr.Suite,
		Message:     tr.Message,
		File:        tr.File,
		Line:        tr.Line,
		Duration:    tr.Duration,
		Failed:      tr.Failed,
		Skipped:     tr.Skipped,
		Passed:      tr.Passed,
		Stdout:      tr.Stdout,
		Stderr:      tr.Stderr,
		Framework:   tr.Framework,
		// Set summary before filtering to maintain summary even if child tests are filtered out
		Summary: lo.ToPtr(tr.Sum()),
	}

	for _, test := range tr.Children {
		if filter.Matches(test) {
			child := test.Filter(filter)
			if !child.IsEmpty() {
				out.Children = append(out.Children, child)
			}
		}
	}
	return out
}

func (tr TestResults) PrettySummary() api.Text {
	return tr.Sum().Pretty()
}

// GetChildren implements the TreeNode interface, organizing tests hierarchically by Suite path.
func (tr TestResults) GetChildren() []api.TreeNode {
	children := []api.TreeNode{}
	for _, test := range tr.Tests {
		children = append(children, test)
	}
	return children
}

// ExecutionResult contains the result of test execution.
type ExecutionResult struct {
	Framework Framework
	ExitCode  int
	Stdout    string
	Stderr    string
	Duration  time.Duration
}
