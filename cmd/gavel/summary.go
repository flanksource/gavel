package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/report"
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

func (b compactSummaryBudget) report() report.Budget {
	return report.Budget{
		MaxFailures:        b.maxFailures,
		MaxLinesPerFailure: b.maxLinesPerFailure,
		MaxCharsPerLine:    b.maxCharsPerLine,
	}
}

// gavelResultJSON mirrors the anonymous struct cmd/gavel/test.go returns when
// --lint is set. It's kept here as a consumer of the JSON wire format so the
// summary command can read any gavel test result file without depending on
// the internal testrunner types.
type gavelResultJSON struct {
	Tests []parsers.Test          `json:"tests"`
	Lint  []*linters.LinterResult `json:"lint"`
	// Error / ExitCode / LogTail are populated by the composite action
	// when gavel crashes before writing results. Stub files carry these
	// fields so `gavel summary` can emit a useful crash marker instead
	// of an empty table.
	Error    string `json:"error,omitempty"`
	ExitCode *int   `json:"exit_code,omitempty"`
	LogTail  string `json:"log_tail,omitempty"`
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

// buildCompactSummary turns a parsed gavel result file into compact markdown.
// Crash stubs (no results + an error field) short-circuit to a crash marker;
// everything else delegates to the shared report renderer.
func buildCompactSummary(data gavelResultJSON, budget compactSummaryBudget) string {
	// Crash-stub short-circuit: if gavel never produced any test or lint
	// results AND the stub carries an error field, emit a crash marker
	// block instead of an empty counts table. The composite action
	// writes these stubs when gavel dies before serialising results.
	if len(data.Tests) == 0 && len(data.Lint) == 0 && data.Error != "" {
		return renderCrashSummary(data, budget)
	}
	return report.BuildCompact(data.Tests, data.Lint, budget.report())
}

// renderCrashSummary emits a PR-comment-ready markdown block for gavel
// crash stubs produced by the composite action. Includes the reported
// error, exit code, and a truncated tail of the captured gavel.log so
// the reader can see *why* the run died without having to download the
// artifact.
func renderCrashSummary(data gavelResultJSON, budget compactSummaryBudget) string {
	var b strings.Builder
	b.WriteString("## Gavel crashed before producing results\n\n")
	if data.ExitCode != nil {
		fmt.Fprintf(&b, "**Exit code:** %d  \n", *data.ExitCode)
	}
	fmt.Fprintf(&b, "**Error:** %s\n\n", data.Error)
	if data.LogTail != "" {
		b.WriteString("### Last lines of gavel.log\n\n```\n")
		b.WriteString(report.TruncateBlock(data.LogTail, budget.maxLinesPerFailure, budget.maxCharsPerLine))
		if !strings.HasSuffix(data.LogTail, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("```\n\n")
	}
	b.WriteString("_Full `gavel.log`, JSON stub, and HTML stub are in the workflow artifact._\n")
	return b.String()
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
