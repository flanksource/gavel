package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestBuildCompactSummary(t *testing.T) {
	// Mirrors the shape cmd/gavel/test.go emits when --lint is set:
	// { "tests": [...], "lint": [...] }
	input := gavelResultJSON{
		Tests: []parsers.Test{
			{
				Package: "github.com/flanksource/gavel/serve",
				Name:    "TestServe",
				Passed:  true,
				Children: parsers.Tests{
					{
						Package:  "github.com/flanksource/gavel/serve",
						Name:     "receives all git tree contents into a worktree",
						Suite:    []string{"SSH Git Serve E2E"},
						Failed:   true,
						File:     "/runner/work/gavel/gavel/serve/e2e_test.go",
						Line:     125,
						Duration: 164 * time.Millisecond,
						Stderr:   longMultilineStderr(),
					},
					{
						Package:  "github.com/flanksource/gavel/serve",
						Name:     "runs linting on the pushed code",
						Suite:    []string{"SSH Git Serve E2E"},
						Failed:   true,
						File:     "/runner/work/gavel/gavel/serve/e2e_test.go",
						Line:     136,
						Duration: 99 * time.Millisecond,
						Message:  "git push failed: Permission denied",
					},
					{
						Package: "github.com/flanksource/gavel/serve",
						Name:    "initializes a bare git repo",
						Suite:   []string{"ensureBareRepo"},
						Passed:  true,
					},
				},
			},
			{
				Package: "github.com/flanksource/gavel/verify",
				Passed:  true,
				Children: parsers.Tests{
					{Package: "github.com/flanksource/gavel/verify", Name: "TestComputeOverallScore", Passed: true},
					{Package: "github.com/flanksource/gavel/verify", Name: "TestLoadConfig", Skipped: true},
				},
			},
		},
		Lint: []*linters.LinterResult{
			{
				Linter:   "golangci-lint",
				Success:  true,
				Duration: 4*time.Minute + 12*time.Second,
				Violations: []models.Violation{
					{
						File:    "cmd/gavel/pr_list.go",
						Line:    254,
						Source:  "errcheck",
						Message: models.StringPtr("Error return value of `mb.Run` is not checked"),
					},
				},
			},
			{
				Linter:   "gofmt",
				Success:  true,
				Duration: 500 * time.Millisecond,
			},
		},
	}

	out := buildCompactSummary(input, compactSummaryBudget{maxFailures: 5, maxLinesPerFailure: 5, maxCharsPerLine: 200})

	// Counts table by source — only sources that have failures or skips
	// should appear. The serve package has failures, the verify package
	// has a skipped test, and golangci-lint reports a violation, so all
	// three keep their row. Sources that are 100% passing (e.g. the
	// clean gofmt linter below) are folded into the passing-only summary
	// row + Totals line.
	if !strings.Contains(out, "github.com/flanksource/gavel/serve") {
		t.Errorf("expected serve package row in counts table, got:\n%s", out)
	}
	if !strings.Contains(out, "github.com/flanksource/gavel/verify") {
		t.Errorf("expected verify package row (has a skipped test) in counts table, got:\n%s", out)
	}
	if !strings.Contains(out, "golangci-lint") {
		t.Errorf("expected golangci-lint row in counts table, got:\n%s", out)
	}
	// gofmt is a passing-only linter — must not appear as its own row.
	if strings.Contains(out, "| lint: gofmt |") {
		t.Errorf("passing-only linter must be folded into the passing summary row, got:\n%s", out)
	}
	// The collapsed-passing row should mention the gofmt linter's pass count
	// (1) and the "more passing source(s)" text.
	if !strings.Contains(out, "more passing source(s)") {
		t.Errorf("expected collapsed passing-source row, got:\n%s", out)
	}

	// Passing rows present, but no per-test listing for passing tests in the detail section.
	if strings.Contains(out, "initializes a bare git repo") {
		t.Errorf("passing tests must not appear in failing detail section, got:\n%s", out)
	}
	if strings.Contains(out, "TestComputeOverallScore") {
		t.Errorf("passing tests must not appear in failing detail section, got:\n%s", out)
	}

	// Failing test details: both failures appear with their suite path.
	if !strings.Contains(out, "SSH Git Serve E2E") {
		t.Errorf("expected failing suite path, got:\n%s", out)
	}
	if !strings.Contains(out, "receives all git tree contents into a worktree") {
		t.Errorf("expected first failing test name, got:\n%s", out)
	}
	if !strings.Contains(out, "runs linting on the pushed code") {
		t.Errorf("expected second failing test name, got:\n%s", out)
	}

	// Stderr truncation: no line in the output should exceed max chars + markdown fencing,
	// and the long stderr fixture should be truncated to 5 lines.
	stderrBlockStart := strings.Index(out, "```")
	if stderrBlockStart < 0 {
		t.Fatalf("expected fenced code block for stderr, got:\n%s", out)
	}
	// Count lines in the first fenced block.
	rest := out[stderrBlockStart+3:]
	end := strings.Index(rest, "```")
	if end < 0 {
		t.Fatalf("expected closing fence for stderr block, got:\n%s", out)
	}
	block := rest[:end]
	// Strip leading newline after opening fence.
	block = strings.TrimPrefix(block, "\n")
	lines := strings.Split(strings.TrimRight(block, "\n"), "\n")
	if len(lines) > 5 {
		t.Errorf("expected ≤5 lines per failure stderr block, got %d:\n%s", len(lines), block)
	}
	for i, line := range lines {
		if len(line) > 200 {
			t.Errorf("line %d exceeds 200 chars (%d):\n%s", i, len(line), line)
		}
	}

	// Linter violations section: the single violation appears with file:line and message.
	if !strings.Contains(out, "cmd/gavel/pr_list.go:254") {
		t.Errorf("expected linter violation file:line, got:\n%s", out)
	}

	// gofmt ran cleanly with no violations — it should appear in the counts table but NOT
	// in the failing-linters section.
	failingSection := out
	if idx := strings.Index(out, "Failing linters"); idx >= 0 {
		failingSection = out[idx:]
	} else {
		failingSection = ""
	}
	if strings.Contains(failingSection, "gofmt") {
		t.Errorf("clean linter must not appear in Failing linters section, got:\n%s", out)
	}
}

func TestBuildCompactSummarySkipsGroupNodes(t *testing.T) {
	// A ginkgo-style tree where the parent group is marked Failed because its
	// child is. The walker must count only the leaf failure, not the parent
	// rollup — otherwise the summary shows noisy "./" / "linters/" entries
	// from folder rollups that mirror leaf state.
	input := gavelResultJSON{
		Tests: []parsers.Test{
			{
				Package: "pkg/a",
				Name:    "linters/",
				Failed:  true, // rollup flag
				Children: parsers.Tests{
					{
						Package: "pkg/a",
						Name:    "jscpd/",
						Failed:  true, // nested rollup
						Children: parsers.Tests{
							{Package: "pkg/a", Name: "real leaf test", Failed: true, Message: "boom"},
						},
					},
				},
			},
		},
	}
	out := buildCompactSummary(input, compactSummaryBudget{maxFailures: 5, maxLinesPerFailure: 5, maxCharsPerLine: 200})
	// Counts: exactly 1 failure in pkg/a (the leaf).
	if !strings.Contains(out, "| pkg/a | 0 | 1 | 0 |") {
		t.Errorf("expected exactly 1 failure for pkg/a, got:\n%s", out)
	}
	// Failing tests section must contain the leaf, NOT the rollup names.
	section := out
	if idx := strings.Index(out, "Failing tests"); idx >= 0 {
		section = out[idx:]
	}
	if !strings.Contains(section, "real leaf test") {
		t.Errorf("expected real leaf in failing tests, got:\n%s", out)
	}
	if strings.Contains(section, "linters/") {
		t.Errorf("rollup group 'linters/' must not appear in failing tests, got:\n%s", out)
	}
	if strings.Contains(section, "jscpd/") {
		t.Errorf("rollup group 'jscpd/' must not appear in failing tests, got:\n%s", out)
	}
}

func TestBuildCompactSummaryRespectsFailureCap(t *testing.T) {
	var leaves parsers.Tests
	for i := 0; i < 12; i++ {
		leaves = append(leaves, parsers.Test{
			Package: "pkg/a",
			Name:    "TestFail",
			Suite:   []string{"Group"},
			Failed:  true,
			Message: "boom",
		})
	}
	// Wrap in a group parent so we also exercise the "skip group, count leaves" path.
	input := gavelResultJSON{
		Tests: []parsers.Test{{Package: "pkg/a", Children: leaves}},
	}
	out := buildCompactSummary(input, compactSummaryBudget{maxFailures: 3, maxLinesPerFailure: 5, maxCharsPerLine: 200})
	// Exactly 3 failure blocks (each starts with "####") in the Failing tests section.
	section := out
	if idx := strings.Index(out, "Failing tests"); idx >= 0 {
		section = out[idx:]
	}
	if end := strings.Index(section, "Failing linters"); end >= 0 {
		section = section[:end]
	}
	blocks := strings.Count(section, "\n#### ")
	if blocks != 3 {
		t.Errorf("expected 3 failure blocks, got %d:\n%s", blocks, section)
	}
	// Truncation notice surfaces the dropped count.
	if !strings.Contains(out, "and 9 more") {
		t.Errorf("expected '... and 9 more' truncation notice, got:\n%s", out)
	}
}

func TestBuildCompactSummaryRendersCrashStub(t *testing.T) {
	exitCode := 139
	input := gavelResultJSON{
		Error:    "gavel exited 139 before writing results",
		ExitCode: &exitCode,
		LogTail:  "panic: runtime error: invalid memory address\ngoroutine 1 [running]:\nmain.main()\n\t/src/cmd/gavel/main.go:72 +0x1a",
	}
	out := buildCompactSummary(input, compactSummaryBudget{maxFailures: 5, maxLinesPerFailure: 5, maxCharsPerLine: 200})

	// Crash heading surfaces the error + exit code.
	if !strings.Contains(out, "Gavel crashed before producing results") {
		t.Errorf("expected crash heading, got:\n%s", out)
	}
	if !strings.Contains(out, "139") {
		t.Errorf("expected exit code in output, got:\n%s", out)
	}
	if !strings.Contains(out, "gavel exited 139 before writing results") {
		t.Errorf("expected error message in output, got:\n%s", out)
	}
	if !strings.Contains(out, "panic: runtime error") {
		t.Errorf("expected log tail in output, got:\n%s", out)
	}
	// Should NOT emit the normal counts table (it would be empty and
	// misleading next to a crash notice).
	if strings.Contains(out, "**Totals:**") {
		t.Errorf("crash stub must not render empty counts totals, got:\n%s", out)
	}
	if strings.Contains(out, "| Source |") {
		t.Errorf("crash stub must not render empty counts table, got:\n%s", out)
	}
}

func TestBuildCompactSummaryPrefersRealResultsOverCrashField(t *testing.T) {
	// If gavel produced real results AND the JSON happens to also carry
	// an error field (e.g. partial-run reporting), the real results win.
	input := gavelResultJSON{
		Tests: []parsers.Test{
			{Package: "pkg/x", Children: parsers.Tests{
				{Package: "pkg/x", Name: "TestOne", Passed: true},
			}},
		},
		Error: "partial error that should be ignored",
	}
	out := buildCompactSummary(input, compactSummaryBudget{maxFailures: 5, maxLinesPerFailure: 5, maxCharsPerLine: 200})
	// pkg/x is 100% passing so its row is collapsed into the totals; we
	// only need to confirm the normal counts path ran (Totals line) and
	// that the crash-stub path did not render.
	if !strings.Contains(out, "**Totals:** 1 passed") {
		t.Errorf("expected normal Totals line for the single passing test, got:\n%s", out)
	}
	if strings.Contains(out, "Gavel crashed") {
		t.Errorf("normal results must not fall through to crash path, got:\n%s", out)
	}
}

func TestRunSummaryReadsJSONFile(t *testing.T) {
	tmp := t.TempDir()
	inputPath := filepath.Join(tmp, "gavel-results.json")
	outputPath := filepath.Join(tmp, "summary.md")

	data := gavelResultJSON{
		Tests: []parsers.Test{
			{Package: "pkg/x", Children: parsers.Tests{
				{Package: "pkg/x", Name: "TestOne", Passed: true},
				{Package: "pkg/x", Name: "TestTwo", Failed: true, Message: "nope"},
			}},
		},
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(inputPath, raw, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if err := runSummary(summaryOptions{InputPath: inputPath, OutputPath: outputPath}); err != nil {
		t.Fatalf("runSummary: %v", err)
	}

	out, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	body := string(out)
	if !strings.Contains(body, "pkg/x") {
		t.Errorf("expected source row, got:\n%s", body)
	}
	if !strings.Contains(body, "TestTwo") {
		t.Errorf("expected failing test, got:\n%s", body)
	}
	if strings.Contains(body, "TestOne") {
		t.Errorf("passing test must not appear in details, got:\n%s", body)
	}
}

// TestBuildCompactSummaryCollapsesAllPassingSources exercises the
// "PR comment must stay short" behaviour: when every source is 100%
// passing, the per-source table disappears entirely and the Totals
// line carries the full headline (count + duration). Regressing to
// the old behaviour (one row per package) blows up PR comments on
// large monorepos.
func TestBuildCompactSummaryCollapsesAllPassingSources(t *testing.T) {
	input := gavelResultJSON{
		Tests: []parsers.Test{
			{Package: "pkg/a", Children: parsers.Tests{
				{Package: "pkg/a", Name: "TestOne", Passed: true, Duration: 10 * time.Millisecond},
				{Package: "pkg/a", Name: "TestTwo", Passed: true, Duration: 20 * time.Millisecond},
			}},
			{Package: "pkg/b", Children: parsers.Tests{
				{Package: "pkg/b", Name: "TestThree", Passed: true, Duration: 30 * time.Millisecond},
			}},
		},
	}
	out := buildCompactSummary(input, compactSummaryBudget{maxFailures: 5, maxLinesPerFailure: 5, maxCharsPerLine: 200})

	// No per-package rows when every source is clean.
	if strings.Contains(out, "| pkg/a |") {
		t.Errorf("all-passing pkg/a must not appear as its own row, got:\n%s", out)
	}
	if strings.Contains(out, "| pkg/b |") {
		t.Errorf("all-passing pkg/b must not appear as its own row, got:\n%s", out)
	}
	// No table headers either — the table itself collapses.
	if strings.Contains(out, "| Source |") {
		t.Errorf("counts table must collapse when every source is clean, got:\n%s", out)
	}
	// Totals line still carries the headline.
	if !strings.Contains(out, "**Totals:** 3 passed · 0 failed · 0 skipped") {
		t.Errorf("expected Totals line with aggregate counts, got:\n%s", out)
	}
}

// TestBuildCompactSummaryCollapsesPassingAlongsideFailing checks the
// mixed case: failing sources keep their own row, but every all-passing
// source is folded into a single trailing summary row whose pass count
// matches the sum of the hidden sources.
func TestBuildCompactSummaryCollapsesPassingAlongsideFailing(t *testing.T) {
	input := gavelResultJSON{
		Tests: []parsers.Test{
			{Package: "pkg/fail", Children: parsers.Tests{
				{Package: "pkg/fail", Name: "TestBoom", Failed: true, Message: "boom"},
			}},
			{Package: "pkg/clean1", Children: parsers.Tests{
				{Package: "pkg/clean1", Name: "TestOne", Passed: true, Duration: 100 * time.Millisecond},
				{Package: "pkg/clean1", Name: "TestTwo", Passed: true, Duration: 100 * time.Millisecond},
			}},
			{Package: "pkg/clean2", Children: parsers.Tests{
				{Package: "pkg/clean2", Name: "TestThree", Passed: true, Duration: 50 * time.Millisecond},
			}},
		},
	}
	out := buildCompactSummary(input, compactSummaryBudget{maxFailures: 5, maxLinesPerFailure: 5, maxCharsPerLine: 200})

	// Failing source keeps its dedicated row.
	if !strings.Contains(out, "| pkg/fail |") {
		t.Errorf("failing pkg/fail row must be present, got:\n%s", out)
	}
	// Clean sources do not appear by name…
	if strings.Contains(out, "pkg/clean1") {
		t.Errorf("clean pkg/clean1 must not appear by name, got:\n%s", out)
	}
	if strings.Contains(out, "pkg/clean2") {
		t.Errorf("clean pkg/clean2 must not appear by name, got:\n%s", out)
	}
	// …but a single collapsed row reports the count of hidden sources
	// AND the sum of their pass counts (2 + 1 = 3).
	if !strings.Contains(out, "_2 more passing source(s)_") {
		t.Errorf("expected '2 more passing source(s)' collapse row, got:\n%s", out)
	}
	if !strings.Contains(out, "| _2 more passing source(s)_ | 3 | 0 | 0 |") {
		t.Errorf("expected collapsed row with summed pass count, got:\n%s", out)
	}
}

func longMultilineStderr() string {
	var sb strings.Builder
	// 8 lines of 250 chars each — must be truncated to 5 lines of ≤200 chars.
	for i := 0; i < 8; i++ {
		sb.WriteString(strings.Repeat("x", 250))
		sb.WriteString("\n")
	}
	return sb.String()
}
