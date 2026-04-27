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
	GoTest     Framework = "go test"
	Ginkgo     Framework = "ginkgo"
	Jest       Framework = "jest"
	Vitest     Framework = "vitest"
	Playwright Framework = "playwright"
)

// String returns the string representation of the framework.
func (f Framework) String() string {
	return string(f)
}

// AllFrameworks lists every framework gavel knows how to run. Order is
// stable so help text and error messages read the same on every invocation.
var AllFrameworks = []Framework{GoTest, Ginkgo, Jest, Vitest, Playwright}

// ParseFramework resolves a user-supplied name to a known Framework. It
// accepts the canonical value ("go test", "ginkgo", ...) and the tolerant
// forms users type at the CLI ("go", "gotest", "go-test"). Returns an error
// listing known names when the input is unrecognized.
func ParseFramework(name string) (Framework, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "go", "gotest", "go-test", "go_test", "go test":
		return GoTest, nil
	case "ginkgo":
		return Ginkgo, nil
	case "jest":
		return Jest, nil
	case "vitest":
		return Vitest, nil
	case "playwright":
		return Playwright, nil
	}
	known := make([]string, len(AllFrameworks))
	for i, f := range AllFrameworks {
		known[i] = string(f)
	}
	return "", fmt.Errorf("unknown framework %q; known: %s", name, strings.Join(known, ", "))
}

// Test represents a single test failure.
type Test struct {
	Name        string        `json:"name,omitempty"`
	Package     string        `json:"package,omitempty"`
	PackagePath string        `json:"package_path,omitempty"` // Relative path to the package (e.g., "./pkg/testrunner")
	WorkDir     string        `json:"work_dir,omitempty"`
	Command     string        `json:"command,omitempty"` // Command used to run this test
	Suite       []string      `json:"suite,omitempty"`   // Hierarchical suite path (e.g., ["Outer Describe", "Inner Context"])
	Message     string        `json:"message,omitempty"`
	File        string        `json:"file,omitempty"`
	Line        int           `json:"line,omitempty"`
	Framework   Framework     `json:"framework,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	Skipped     bool          `json:"skipped,omitempty"`
	Failed      bool          `json:"failed,omitempty"`
	Passed      bool          `json:"passed,omitempty"`
	Pending     bool          `json:"pending,omitempty"`
	Cached      bool          `json:"cached,omitempty"`    // True when this result came from gavel's run-cache, not a fresh run
	TimedOut    bool          `json:"timed_out,omitempty"` // True when the test package subprocess was killed by the --test-timeout or --timeout supervisor
	TaskID      string        `json:"task_id,omitempty"`
	CanStop     bool          `json:"can_stop,omitempty"`
	// IsGinkgoBootstrap marks a Go test function whose body only invokes ginkgo's RunSpecs.
	// These wrappers still carry pass/fail/duration for the whole suite when a Ginkgo
	// JSON report file is unavailable, but are deduped against real specs from the
	// Ginkgo parser at merge time in runner.go.
	IsGinkgoBootstrap bool             `json:"is_ginkgo_bootstrap,omitempty"`
	Stdout            string           `json:"stdout,omitempty"`
	Stderr            string           `json:"stderr,omitempty"`
	Children          Tests            `json:"children,omitempty"`
	Summary           *TestSummary     `json:"summary,omitempty"`
	Context           any              `json:"context,omitempty"`
	Benchmark         *BenchmarkResult `json:"benchmark,omitempty"`
	// Attempts is the per-run execution history for this test. A fresh run
	// appends a TestAttempt to the tail; reruns (via the UI) append further
	// attempts without discarding earlier ones. The Test's top-level
	// Passed/Failed/Skipped/Pending/TimedOut flags always reflect the most
	// recent attempt so existing filters keep working.
	Attempts []TestAttempt `json:"attempts,omitempty"`
	// FailureDetail is a structured view of Message recognised at parse time
	// (gomega expected/actual, panic+stack, go test trailers). Renderers use
	// it to show side-by-side diffs and a short summary; Message stays as the
	// canonical raw form so consumers that want the original still see it.
	FailureDetail *FailureDetail `json:"failure_detail,omitempty"`
}

// TestAttempt records one execution of a test: when it ran, on what pid,
// the command invoked, and the final state. Populated by the runner on
// completion and extended by the UI on rerun.
type TestAttempt struct {
	Sequence    int           `json:"sequence"`
	RunKind     string        `json:"run_kind,omitempty"` // "initial", "rerun"
	Started     time.Time     `json:"started,omitempty"`
	Ended       time.Time     `json:"ended,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	PID         int           `json:"pid,omitempty"`
	Command     string        `json:"command,omitempty"`
	Framework   Framework     `json:"framework,omitempty"`
	ExitCode    *int          `json:"exit_code,omitempty"`
	Passed      bool          `json:"passed,omitempty"`
	Failed      bool          `json:"failed,omitempty"`
	Skipped     bool          `json:"skipped,omitempty"`
	Pending     bool          `json:"pending,omitempty"`
	TimedOut    bool          `json:"timed_out,omitempty"`
	Message     string        `json:"message,omitempty"`
	Stdout      string        `json:"stdout,omitempty"`
	Stderr      string        `json:"stderr,omitempty"`
	StackTrace  string        `json:"stack_trace,omitempty"`
	CPUPercent  float64       `json:"cpu_percent,omitempty"`
	RSS         uint64        `json:"rss,omitempty"`
	GoroutineCt int           `json:"goroutine_count,omitempty"`
}

type GoTestContext struct {
	ParentTest string `json:"parent_test,omitempty"`
	ImportPath string `json:"import_path,omitempty"`
}

type GinkgoContext struct {
	SuiteDescription string `json:"suite_description,omitempty"`
	SuitePath        string `json:"suite_path,omitempty"`
	FailureLocation  string `json:"failure_location,omitempty"`
}

type BenchmarkResult struct {
	Iterations  int       `json:"iterations,omitempty"`
	NsPerOp     float64   `json:"ns_per_op,omitempty"`
	BytesPerOp  int64     `json:"bytes_per_op,omitempty"`
	AllocsPerOp int64     `json:"allocs_per_op,omitempty"`
	MBPerSec    float64   `json:"mb_per_sec,omitempty"`
	Samples     []float64 `json:"samples,omitempty"` // ns/op across -count=N runs; NsPerOp is the last/mean
}

func (b BenchmarkResult) Pretty() string {
	parts := []string{fmt.Sprintf("%d iterations", b.Iterations)}
	if b.NsPerOp > 0 {
		parts = append(parts, fmt.Sprintf("%.2f ns/op", b.NsPerOp))
	}
	if b.MBPerSec > 0 {
		parts = append(parts, fmt.Sprintf("%.2f MB/s", b.MBPerSec))
	}
	if b.BytesPerOp > 0 {
		parts = append(parts, fmt.Sprintf("%d B/op", b.BytesPerOp))
	}
	if b.AllocsPerOp > 0 {
		parts = append(parts, fmt.Sprintf("%d allocs/op", b.AllocsPerOp))
	}
	return " [" + strings.Join(parts, ", ") + "]"
}

type FixtureContext struct {
	Command       string         `json:"command,omitempty"`
	ExitCode      int            `json:"exit_code,omitempty"`
	CWD           string         `json:"cwd,omitempty"`
	CELExpression string         `json:"cel_expression,omitempty"`
	CELVars       map[string]any `json:"cel_vars,omitempty"`
	Expected      any            `json:"expected,omitempty"`
	Actual        any            `json:"actual,omitempty"`
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
	switch {
	case t.TimedOut:
		s = s.Append("⏳", "text-amber-600 font-bold")
		textStyle = "text-amber-600"
	case t.Skipped:
		s = s.Append(icons.Skip, "text-orange-500")
		textStyle = "text-yellow-500"
	case t.Failed:
		s = s.Append(icons.Fail, "text-red-500")
		textStyle = "text-red-500"
	default:
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
	if t.TimedOut {
		s = s.Space().Append("[timed out]", "text-amber-600 font-bold")
	}

	// Add benchmark metrics if present
	if t.Benchmark != nil {
		s = s.Append(t.Benchmark.Pretty(), "text-muted")
	}

	// Add message if present. Prefer the structured FailureDetail summary
	// when available so the per-test line stays scannable; the full body
	// drops underneath in collapsed sections.
	if d := t.FailureDetail; d != nil && d.Summary != "" {
		s = s.Space().Append(d.Summary, textStyle)
	} else if t.Message != "" {
		s = s.Space().Append(t.Message, textStyle)
	}
	if t.FailureDetail != nil {
		s = appendFailureDetail(s, t.FailureDetail, t.Message)
	}
	if t.Failed && t.Stdout != "" {
		s = s.NewLine().Add(api.Collapsed{
			Label:   "stdout",
			Content: clicky.Text(t.Stdout, "text-red-500 font-mono text-xs whitespace-pre-wrap"),
		})
	}
	if t.Stderr != "" {
		s = s.NewLine().Add(api.Collapsed{
			Label:   "stderr",
			Content: clicky.Text(t.Stderr, "text-red-500 font-mono text-xs whitespace-pre-wrap"),
		})
	}

	return s
}

// appendFailureDetail renders the structured parts of a gomega/panic/go-test
// failure beneath the per-test summary line. Uses collapsed code blocks for
// long values so the terminal stays readable on a screen full of failures.
func appendFailureDetail(s api.Text, d *FailureDetail, raw string) api.Text {
	switch d.Kind {
	case FailureKindGomega:
		if d.Actual != "" {
			s = s.NewLine().Add(api.Collapsed{
				Label:   "actual",
				Content: clicky.Text("").Add(api.Code{Content: d.Actual}),
			})
		}
		if d.Expected != "" {
			s = s.NewLine().Add(api.Collapsed{
				Label:   "expected (" + d.Matcher + ")",
				Content: clicky.Text("").Add(api.Code{Content: d.Expected}),
			})
		}
	case FailureKindPanic:
		if d.Stack != "" {
			s = s.NewLine().Add(api.Collapsed{
				Label:   "stack trace",
				Content: clicky.Text(d.Stack, "text-red-500 font-mono text-xs whitespace-pre-wrap"),
			})
		}
	case FailureKindGoTest:
		if d.Location != "" {
			s = s.NewLine().Append(d.Location, "text-muted font-mono")
		}
	}
	// Always keep the raw message reachable so nothing is hidden by parsing
	// heuristics — but only when the raw form has more than the one-line
	// summary, otherwise it's pure noise.
	if raw != "" && raw != d.Summary && strings.Contains(raw, "\n") {
		s = s.NewLine().Add(api.Collapsed{
			Label:   "full message",
			Content: clicky.Text(raw, "text-muted font-mono text-xs whitespace-pre-wrap"),
		})
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
	case Jest:
		if t.File != "" {
			return fmt.Sprintf(`npx jest -t %q %s`, t.FullName(), t.File)
		}
		return fmt.Sprintf(`npx jest -t %q`, t.FullName())
	case Vitest:
		if t.File != "" {
			return fmt.Sprintf(`npx vitest run -t %q %s`, t.FullName(), t.File)
		}
		return fmt.Sprintf(`npx vitest run -t %q`, t.FullName())
	case Playwright:
		if t.File != "" && t.Line > 0 {
			return fmt.Sprintf(`npx playwright test %s:%d`, t.File, t.Line)
		}
		if t.File != "" {
			return fmt.Sprintf(`npx playwright test %s`, t.File)
		}
		return fmt.Sprintf(`npx playwright test -g %q`, t.FullName())
	default:
		return ""
	}
}

// PrettyTODO returns the markdown body for a TODO file (excluding frontmatter).
func (t Test) PrettyTODO() api.Text {
	text := clicky.Text("")
	text = text.Add(api.KeyValuePair{Key: "Test Name", Value: t.Name}).NewLine()
	if len(t.Suite) > 0 {
		text = text.Add(api.KeyValuePair{Key: "Suite", Value: strings.Join(t.Suite, " > ")}).NewLine()
	}
	text = text.Add(api.KeyValuePair{Key: "Package", Value: t.Package}).NewLine()
	text = text.Add(api.KeyValuePair{Key: "File", Value: fmt.Sprintf("%s:%d", t.File, t.Line)}).NewLine()

	// Re-run Command section
	text = text.NewLine().Append("## Re-run Command", "").NewLine().NewLine()
	text = text.Add(api.Code{Content: t.RerunCommand(), Language: "bash"}).NewLine()

	// Latest Failure section. When we have a structured FailureDetail (gomega),
	// emit Expected/Actual as labelled fenced blocks so the markdown reader
	// sees the assertion shape directly instead of one big code dump.
	text = text.NewLine().Append("## Latest Failure", "").NewLine().NewLine()
	if d := t.FailureDetail; d != nil && d.Kind == FailureKindGomega {
		if d.Summary != "" {
			text = text.Append(d.Summary, "").NewLine().NewLine()
		}
		if d.Actual != "" {
			text = text.Append("**Actual**", "").NewLine().NewLine()
			text = text.Add(api.Code{Content: d.Actual}).NewLine().NewLine()
		}
		if d.Expected != "" {
			label := "**Expected**"
			if d.Matcher != "" {
				label = "**Expected (" + d.Matcher + ")**"
			}
			text = text.Append(label, "").NewLine().NewLine()
			text = text.Add(api.Code{Content: d.Expected}).NewLine()
		}
	} else if t.Message != "" {
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
	} else if tr.Pending {
		summary.Pending = 1
	} else {
		summary.Total = 0
	}

	for _, child := range tr.Children {
		childSummary := child.Sum()
		summary.Total += childSummary.Total
		summary.Passed += childSummary.Passed
		summary.Failed += childSummary.Failed
		summary.Skipped += childSummary.Skipped
		summary.Pending += childSummary.Pending
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
	Pending  int
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
		Pending:  tr.Pending + other.Pending,
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
		Name:              tr.Name,
		Package:           tr.Package,
		PackagePath:       tr.PackagePath,
		Command:           tr.Command,
		Suite:             tr.Suite,
		Message:           tr.Message,
		File:              tr.File,
		Line:              tr.Line,
		Duration:          tr.Duration,
		Failed:            tr.Failed,
		Skipped:           tr.Skipped,
		Passed:            tr.Passed,
		Stdout:            tr.Stdout,
		Stderr:            tr.Stderr,
		Framework:         tr.Framework,
		Benchmark:         tr.Benchmark,
		IsGinkgoBootstrap: tr.IsGinkgoBootstrap,
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
