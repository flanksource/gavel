package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type summaryOptions struct {
	InputPath  string
	OutputPath string
}

type compactSummaryBudget struct {
	maxFailures        int
	maxLinesPerFailure int
	maxCharsPerLine    int
}

var defaultCompactBudget = compactSummaryBudget{
	maxFailures:        5,
	maxLinesPerFailure: 5,
	maxCharsPerLine:    200,
}

// gavelResultJSON mirrors the anonymous struct cmd/gavel/test.go returns when
// --lint is set. It's kept here as a consumer of the JSON wire format so the
// summary command can read any gavel test result file without depending on
// the internal testrunner types.
type gavelResultJSON struct {
	Tests []parsers.Test          `json:"tests"`
	Lint  []*linters.LinterResult `json:"lint"`
}

// UnmarshalJSON accepts both shapes gavel emits:
//   - plain `test`:        a JSON array of parsers.Test
//   - `test --lint`:       an object with `tests` and `lint` keys
func (g *gavelResultJSON) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "[") {
		var tests []parsers.Test
		if err := json.Unmarshal(data, &tests); err != nil {
			return err
		}
		g.Tests = tests
		return nil
	}
	type alias gavelResultJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*g = gavelResultJSON(a)
	return nil
}

func runSummary(opts summaryOptions) error {
	if opts.InputPath == "" {
		return fmt.Errorf("--input is required")
	}
	raw, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", opts.InputPath, err)
	}
	var data gavelResultJSON
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("parse %s: %w", opts.InputPath, err)
	}
	md := buildCompactSummary(data, defaultCompactBudget)
	if opts.OutputPath == "" {
		_, err := os.Stdout.WriteString(md)
		return err
	}
	if err := os.WriteFile(opts.OutputPath, []byte(md), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", opts.OutputPath, err)
	}
	return nil
}

type sourceCounts struct {
	name     string
	passed   int
	failed   int
	skipped  int
	duration time.Duration
}

func buildCompactSummary(data gavelResultJSON, budget compactSummaryBudget) string {
	sources := make(map[string]*sourceCounts)
	var failures []parsers.Test
	for _, root := range data.Tests {
		walkTests(root, sources, &failures)
	}

	// Linters become their own "source" rows and contribute to the failing section.
	var failingLinters []*linters.LinterResult
	for _, lr := range data.Lint {
		key := "lint: " + lr.Linter
		sc := ensureSource(sources, key)
		sc.duration += lr.Duration
		switch {
		case lr.Skipped:
			sc.skipped++
		case lr.TimedOut, !lr.Success, lr.HasViolations():
			sc.failed++
			failingLinters = append(failingLinters, lr)
		default:
			sc.passed++
		}
	}

	var b strings.Builder
	writeCountsTable(&b, sources)
	writeTotals(&b, sources)
	writeFailingTests(&b, failures, budget)
	writeFailingLinters(&b, failingLinters, budget)
	return b.String()
}

func walkTests(t parsers.Test, sources map[string]*sourceCounts, failures *[]parsers.Test) {
	// Recurse first so leaves are always processed, regardless of any
	// status flags set on parent group nodes.
	for _, child := range t.Children {
		walkTests(child, sources, failures)
	}
	// Only leaf nodes contribute to counts and failure details. A node with
	// children is a group/folder rollup whose status mirrors its children;
	// counting it would double-count, and surfacing it as a failure detail
	// produces noisy "./" / "linters/" / "testdata/" entries in the summary.
	if len(t.Children) > 0 {
		return
	}
	// IsFolder() returns true when no status flag is set — a pure organizational
	// node with nothing to report. Skip it.
	if t.IsFolder() {
		return
	}
	source := sourceKey(t)
	sc := ensureSource(sources, source)
	sc.duration += t.Duration
	switch {
	case t.Failed:
		sc.failed++
		*failures = append(*failures, t)
	case t.Skipped, t.Pending:
		sc.skipped++
	case t.Passed:
		sc.passed++
	}
}

// sourceKey picks the best attribution label for a leaf test result. Ginkgo
// specs often have an empty Package but carry the suite info in the Suite
// slice; go tests carry Package. Fall back to File dir, then "(unknown)".
func sourceKey(t parsers.Test) string {
	if t.Package != "" {
		return t.Package
	}
	if t.PackagePath != "" {
		return t.PackagePath
	}
	if t.Command != "" {
		return t.Command
	}
	if len(t.Suite) > 0 {
		return t.Suite[0]
	}
	if t.File != "" {
		// Use the directory portion as a proxy when the parser didn't set Package.
		if idx := strings.LastIndex(t.File, "/"); idx > 0 {
			return t.File[:idx]
		}
		return t.File
	}
	return "(unknown)"
}

func ensureSource(sources map[string]*sourceCounts, name string) *sourceCounts {
	if sc, ok := sources[name]; ok {
		return sc
	}
	sc := &sourceCounts{name: name}
	sources[name] = sc
	return sc
}

func writeCountsTable(b *strings.Builder, sources map[string]*sourceCounts) {
	rows := make([]*sourceCounts, 0, len(sources))
	for _, sc := range sources {
		rows = append(rows, sc)
	}
	sort.Slice(rows, func(i, j int) bool {
		// Failing sources first so they're visible above the fold.
		if (rows[i].failed > 0) != (rows[j].failed > 0) {
			return rows[i].failed > 0
		}
		return rows[i].name < rows[j].name
	})

	b.WriteString("## Gavel summary\n\n")
	b.WriteString("| Source | Pass | Fail | Skip | Duration |\n")
	b.WriteString("|---|---:|---:|---:|---:|\n")
	for _, sc := range rows {
		fmt.Fprintf(b, "| %s | %d | %d | %d | %s |\n",
			escapePipe(sc.name), sc.passed, sc.failed, sc.skipped, formatDuration(sc.duration))
	}
	b.WriteString("\n")
}

func writeTotals(b *strings.Builder, sources map[string]*sourceCounts) {
	var totals sourceCounts
	for _, sc := range sources {
		totals.passed += sc.passed
		totals.failed += sc.failed
		totals.skipped += sc.skipped
		totals.duration += sc.duration
	}
	fmt.Fprintf(b, "**Totals:** %d passed · %d failed · %d skipped · %s\n\n",
		totals.passed, totals.failed, totals.skipped, formatDuration(totals.duration))
}

func writeFailingTests(b *strings.Builder, failures []parsers.Test, budget compactSummaryBudget) {
	if len(failures) == 0 {
		return
	}
	b.WriteString("### Failing tests\n\n")
	shown := failures
	if len(shown) > budget.maxFailures {
		shown = shown[:budget.maxFailures]
	}
	for _, t := range shown {
		writeFailureBlock(b, t, budget)
	}
	if dropped := len(failures) - len(shown); dropped > 0 {
		fmt.Fprintf(b, "_... and %d more failing tests — see the full gavel-results artifact._\n\n", dropped)
	}
}

func writeFailureBlock(b *strings.Builder, t parsers.Test, budget compactSummaryBudget) {
	title := t.FullName()
	if t.Package != "" {
		fmt.Fprintf(b, "#### %s — %s\n", escapeMarkdown(t.Package), escapeMarkdown(title))
	} else {
		fmt.Fprintf(b, "#### %s\n", escapeMarkdown(title))
	}
	if t.File != "" {
		loc := t.File
		if t.Line > 0 {
			loc = fmt.Sprintf("%s:%d", t.File, t.Line)
		}
		fmt.Fprintf(b, "`%s`\n\n", loc)
	}
	body := firstNonEmpty(t.Stderr, t.Stdout, t.Message)
	if body == "" {
		return
	}
	b.WriteString("```\n")
	b.WriteString(truncateBlock(body, budget.maxLinesPerFailure, budget.maxCharsPerLine))
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("```\n\n")
}

func writeFailingLinters(b *strings.Builder, failing []*linters.LinterResult, budget compactSummaryBudget) {
	if len(failing) == 0 {
		return
	}
	b.WriteString("### Failing linters\n\n")
	for _, lr := range failing {
		if lr.TimedOut {
			fmt.Fprintf(b, "#### %s — timed out after %s\n\n", lr.Linter, formatDuration(lr.Duration))
			continue
		}
		if !lr.Success && lr.Error != "" {
			fmt.Fprintf(b, "#### %s — error\n\n", lr.Linter)
			b.WriteString("```\n")
			b.WriteString(truncateBlock(lr.Error, budget.maxLinesPerFailure, budget.maxCharsPerLine))
			b.WriteString("\n```\n\n")
			continue
		}
		fmt.Fprintf(b, "#### %s — %d violation(s)\n", lr.Linter, len(lr.Violations))
		shown := lr.Violations
		if len(shown) > budget.maxFailures {
			shown = shown[:budget.maxFailures]
		}
		for _, v := range shown {
			loc := v.File
			if v.Line > 0 {
				loc = fmt.Sprintf("%s:%d", v.File, v.Line)
			}
			rule := v.Source
			if v.Rule != nil && v.Rule.Pattern != "" {
				rule = v.Rule.Pattern
			}
			msg := ""
			if v.Message != nil {
				msg = *v.Message
			}
			suffix := ""
			if rule != "" {
				suffix = fmt.Sprintf(" (%s)", rule)
			}
			fmt.Fprintf(b, "- `%s` — %s%s\n", loc, truncateLine(msg, budget.maxCharsPerLine), suffix)
		}
		if dropped := len(lr.Violations) - len(shown); dropped > 0 {
			fmt.Fprintf(b, "- _... and %d more violations_\n", dropped)
		}
		b.WriteString("\n")
	}
}

func truncateBlock(body string, maxLines, maxCharsPerLine int) string {
	all := strings.Split(strings.TrimRight(body, "\n"), "\n")
	lines := all
	if len(all) > maxLines {
		// Reserve the last slot for the truncation notice so the block fits
		// within maxLines even with the extra line appended.
		keep := maxLines - 1
		if keep < 0 {
			keep = 0
		}
		lines = append([]string{}, all[:keep]...)
		lines = append(lines, fmt.Sprintf("... (%d more lines truncated)", len(all)-keep))
	}
	for i, line := range lines {
		lines[i] = truncateLine(line, maxCharsPerLine)
	}
	return strings.Join(lines, "\n")
}

func truncateLine(s string, maxChars int) string {
	if maxChars <= 0 || len(s) <= maxChars {
		return s
	}
	if maxChars <= 3 {
		return s[:maxChars]
	}
	return s[:maxChars-3] + "..."
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	if d < time.Millisecond {
		return d.String()
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return d.Round(time.Second).String()
}

func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

func escapeMarkdown(s string) string {
	// Minimal escaping for headings: backticks and brackets only. Full markdown
	// escaping would be noise for test names the user will actually read.
	return strings.NewReplacer("`", "\\`").Replace(s)
}

func init() {
	opts := summaryOptions{}
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Build a compact markdown PR-comment summary from a gavel test JSON file",
		Long: `Read a gavel test/lint JSON result file and emit a compact markdown
summary suitable for a GitHub PR comment or job step summary. The output
contains a counts table grouped by source (test package or linter), totals,
and up to 5 failing tests with the first 5 lines (≤200 chars each) of their
stderr/stdout/message. Passing tests and clean linters do not appear in the
detail sections. The full report remains in the JSON/HTML artifacts.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSummary(opts)
		},
	}
	cmd.Flags().StringVar(&opts.InputPath, "input", "gavel-results.json", "Path to gavel JSON result file")
	cmd.Flags().StringVar(&opts.OutputPath, "output", "", "Path to write compact markdown (default: stdout)")
	rootCmd.AddCommand(cmd)
}
