package types

import (
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// childRow is the part of a runner-step child a verify node is built from.
type childRow struct {
	Name      string
	Framework string
	Status    task.Status
	TimedOut  bool
	Error     string
}

func childRows(nodes []*fixtures.FixtureNode) []childRow {
	rows := make([]childRow, 0, len(nodes))
	for _, n := range nodes {
		Expect(n.Name).To(Equal(n.Results.Name))
		rows = append(rows, childRow{n.Name, n.Results.Type, n.Results.Status, n.Results.TimedOut, n.Results.Error})
	}
	return rows
}

var showAll = func() stepDisplay { yes := true; return stepDisplay{ShowPassed: &yes} }()

var _ = Describe("runner step children", func() {
	tests := parsers.Tests{
		{Name: "pkg", Framework: parsers.GoTest, Children: parsers.Tests{
			{Name: "TestPass", Framework: parsers.GoTest, Passed: true},
			{Name: "TestSlow", Framework: parsers.GoTest, Failed: true, TimedOut: true, Message: "panic: test timed out"},
			{Name: "TestQueued", Framework: parsers.GoTest, Pending: true},
		}},
		{Name: "adds", Suite: []string{"Calc"}, Framework: parsers.Ginkgo, Running: true},
		{Name: "rejects", Suite: []string{"Calc"}, Framework: parsers.Ginkgo, Failed: true, Message: "expected 1"},
		{Name: "flaky", Framework: parsers.Vitest, Warned: true, Message: "retried"},
		{Name: "todo", Framework: parsers.Vitest, Skipped: true},
		{Name: "placeholder", Framework: parsers.Vitest},
	}

	It("keeps each child's framework and never reports an unfinished or timed-out test as passed", func() {
		children, summary := runnerStepChildren(tests, showAll)

		Expect(childRows(children)).To(Equal([]childRow{
			{"TestPass", "go test", task.StatusPASS, false, ""},
			{"TestSlow", "go test", task.StatusFAIL, true, "panic: test timed out"},
			{"TestQueued", "go test", task.StatusPending, false, ""},
			{"Calc > adds", "ginkgo", task.StatusRunning, false, ""},
			{"Calc > rejects", "ginkgo", task.StatusFAIL, false, "expected 1"},
			{"flaky", "vitest", task.StatusWarning, false, "retried"},
			{"todo", "vitest", task.StatusSKIP, false, ""},
		}))
		Expect(summary).To(Equal(&api.VerifySummary{
			Total: 7, Passed: 1, Failed: 1, Warned: 1, Skipped: 1, Pending: 1, Running: 1, TimedOut: 1,
		}))
	})

	It("elides passed and skipped children by default but still counts them", func() {
		children, summary := runnerStepChildren(tests, stepDisplay{})

		names := make([]string, 0, len(children))
		for _, c := range children {
			names = append(names, c.Name)
		}
		Expect(names).To(Equal([]string{"TestSlow", "TestQueued", "Calc > adds", "Calc > rejects", "flaky"}))
		Expect(summary.Total).To(Equal(7))
	})

	Describe("lint results", func() {
		msg := func(s string) *string { return &s }
		results := []*linters.LinterResult{
			{Linter: "golangci-lint", Success: true, Duration: time.Second, Violations: []models.Violation{
				{File: "a.go", Line: 3, Message: msg("unchecked error"), Severity: models.SeverityError},
				{File: "a.go", Line: 9, Message: msg("unused param"), Severity: models.SeverityWarning},
				{File: "b.go", Line: 1, Message: msg("style nit"), Severity: models.SeverityInfo},
			}},
			{Linter: "ruff", Success: true},
			{Linter: "markdownlint", Error: "exit status 2"},
			{Linter: "tsc", Skipped: true, Error: "no tsconfig.json"},
		}
		findingsAndCrash := []childRow{
			{"a.go:3", "golangci-lint", task.StatusFAIL, false, "unchecked error"},
			{"a.go:9", "golangci-lint", task.StatusWarning, false, "unused param"},
			{"b.go:1", "golangci-lint", task.StatusSKIP, false, "style nit"},
		}
		crash := childRow{"markdownlint", "markdownlint", task.StatusFAIL, false, "exit status 2"}

		DescribeTable("shows findings of every severity and a crashed linter's row, failing the step on that row",
			func(disp stepDisplay, want []childRow) {
				result := fixtures.FixtureResult{}

				applyLintResults(&result, results, disp)

				Expect(childRows(result.Children)).To(Equal(want))
				Expect(result.Summary).To(Equal(&api.VerifySummary{Total: 5, Passed: 1, Failed: 2, Warned: 1, Skipped: 1}))
				Expect([]any{result.Status, result.Error, result.Metadata}).To(Equal([]any{
					task.StatusFAIL, "2/5 lint results failed", map[string]interface{}{"violations": 3},
				}))
			},
			Entry("by default, clean linters elided", stepDisplay{},
				append(append([]childRow{}, findingsAndCrash...), crash)),
			Entry("with show-passed, clean linters shown", showAll,
				append(append([]childRow{}, findingsAndCrash...), childRow{"ruff", "ruff", task.StatusPASS, false, ""}, crash)),
		)
	})

	DescribeTable("step verdict from the children's summary",
		func(summary api.VerifySummary, wantStatus task.Status, wantError string) {
			result := fixtures.FixtureResult{}
			applyStepVerdict(&result, summary, "tests")
			Expect([]any{result.Status, result.Error}).To(Equal([]any{wantStatus, wantError}))
		},
		Entry("all passed", api.VerifySummary{Total: 2, Passed: 2}, task.StatusPASS, ""),
		Entry("only warnings", api.VerifySummary{Total: 2, Passed: 1, Warned: 1}, task.StatusWarning, "1/2 tests warned"),
		Entry("a failure", api.VerifySummary{Total: 3, Passed: 1, Failed: 1, Warned: 1}, task.StatusFAIL, "1/3 tests failed"),
		Entry("a timeout", api.VerifySummary{Total: 2, Passed: 1, TimedOut: 1}, task.StatusFAIL, "1/2 tests failed"),
		Entry("unfinished", api.VerifySummary{Total: 3, Passed: 1, Pending: 1, Running: 1}, task.StatusFAIL, "2/3 tests did not finish"),
	)
})
