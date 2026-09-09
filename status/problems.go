package status

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// ProblemKind distinguishes the source of a Problem so the renderer can pick
// the right icon and so callers can filter by kind.
type ProblemKind string

const (
	ProblemKindTest ProblemKind = "test"
	ProblemKindLint ProblemKind = "lint"
	// ProblemKindAI is assembled at render time from FileStatus.AIError; it
	// only appears in `--ai` runs where a per-file summary failed.
	ProblemKindAI ProblemKind = "ai"
)

// Problem is one failing test, lint violation, or AI-summary error tied to a
// changed file. It carries the human-readable message the trailing "Problems"
// section renders — the counts on TestStatus/LintStatus drive the inline
// badges instead.
type Problem struct {
	Kind ProblemKind
	// Severity is "failed"/"warned" for tests and "error"/"warning" for lint.
	// It selects the icon colour and the sort rank (red before amber).
	Severity string
	// Label identifies the problem: the test name (suite-qualified) or the
	// lint rule / linter name.
	Label   string
	Line    int
	Message string
}

const (
	maxProblemsPerFileDefault = 2
	maxProblemMessageRunes    = 100
)

func collectTestProblems(tests []parsers.Test, workDir string, out map[string][]Problem) {
	for _, t := range tests {
		// A failed parent is represented by its failing leaves; only emit
		// problem lines for leaves so container rollups don't duplicate them.
		if len(t.Children) > 0 {
			collectTestProblems(t.Children, workDir, out)
			continue
		}
		if !t.Failed && !t.Warned {
			continue
		}
		path := normalisePath(t.File, workDir)
		if path == "" {
			continue
		}
		severity := "failed"
		if t.Warned && !t.Failed {
			severity = "warned"
		}
		out[path] = append(out[path], Problem{
			Kind:     ProblemKindTest,
			Severity: severity,
			Label:    testProblemLabel(t),
			Line:     t.Line,
			Message:  parsers.StripANSI(strings.TrimSpace(t.Message)),
		})
	}
}

func testProblemLabel(t parsers.Test) string {
	name := strings.TrimSpace(t.Name)
	if len(t.Suite) > 0 {
		suite := strings.Join(t.Suite, " › ")
		if name != "" {
			return suite + " › " + name
		}
		return suite
	}
	if name != "" {
		return name
	}
	return "(unnamed test)"
}

func collectLintProblems(lint []*linters.LinterResult, workDir string, out map[string][]Problem) {
	for _, lr := range lint {
		if lr == nil {
			continue
		}
		for _, v := range lr.Violations {
			severity := normaliseLintSeverity(v.Severity)
			if severity == "" {
				// Infos stay badge-only; the section lists actionable problems.
				continue
			}
			path := normalisePath(v.File, workDir)
			if path == "" {
				continue
			}
			out[path] = append(out[path], Problem{
				Kind:     ProblemKindLint,
				Severity: severity,
				Label:    lintProblemLabel(v),
				Line:     v.Line,
				Message:  lintProblemMessage(v),
			})
		}
	}
}

// normaliseLintSeverity folds a linter's severity vocabulary onto "error" or
// "warning", returning "" for info-level (and unknown) severities.
func normaliseLintSeverity(s models.ViolationSeverity) string {
	switch strings.ToLower(string(s)) {
	case "error", "critical", "high":
		return "error"
	case "warning", "medium":
		return "warning"
	default:
		return ""
	}
}

func lintProblemLabel(v models.Violation) string {
	if v.Rule != nil && strings.TrimSpace(v.Rule.Method) != "" {
		return v.Rule.Method
	}
	if strings.TrimSpace(v.Source) != "" {
		return v.Source
	}
	return "lint"
}

func lintProblemMessage(v models.Violation) string {
	if v.Message == nil {
		return ""
	}
	return parsers.StripANSI(strings.TrimSpace(*v.Message))
}

// sortProblems orders a file's problems so red items (test failures, lint
// errors) precede amber ones, keeping the most urgent lines at the top when
// the default view truncates.
func sortProblems(problems []Problem) []Problem {
	sort.SliceStable(problems, func(i, j int) bool {
		return problemRank(problems[i]) < problemRank(problems[j])
	})
	return problems
}

func problemRank(p Problem) int {
	switch {
	case p.Kind == ProblemKindTest && p.Severity == "failed":
		return 0
	case p.Kind == ProblemKindLint && p.Severity == "error":
		return 1
	case p.Kind == ProblemKindAI:
		return 2
	case p.Kind == ProblemKindTest && p.Severity == "warned":
		return 3
	default:
		return 4
	}
}

// fileProblems bundles a file's display path with its problems (including any
// live AI-summary error) for the renderer.
type fileProblems struct {
	path     string
	base     string
	problems []Problem
}

func (r *Result) collectFileProblems() []fileProblems {
	var groups []fileProblems
	for _, f := range r.Files {
		problems := f.Problems
		if strings.TrimSpace(f.AIError) != "" {
			problems = append(append([]Problem{}, problems...), Problem{
				Kind:     ProblemKindAI,
				Severity: "error",
				Label:    "ai summary",
				Message:  strings.TrimSpace(f.AIError),
			})
		}
		if len(problems) == 0 {
			continue
		}
		groups = append(groups, fileProblems{
			path:     f.Path,
			base:     filepath.Base(f.Path),
			problems: sortProblems(problems),
		})
	}
	return groups
}

// prettyProblems renders the trailing "Problems" section: failing tests and
// lint violations grouped by file. The default view caps each file to
// maxProblemsPerFileDefault and single-line messages, then points at
// `gavel status -v`; verbose mode shows every problem with full messages.
func (r *Result) prettyProblems() api.Text {
	t := clicky.Text("")
	groups := r.collectFileProblems()
	total := 0
	for _, g := range groups {
		total += len(g.problems)
	}
	if total == 0 {
		return t
	}

	t = t.NewLine().Append(fmt.Sprintf("Problems (%d)", total), "font-bold "+styleError).NewLine()

	anyCapped := false
	for _, g := range groups {
		t = t.Append(" ", "").Append(g.path, styleScope).NewLine()

		shown := g.problems
		if !r.Verbose && len(shown) > maxProblemsPerFileDefault {
			shown = shown[:maxProblemsPerFileDefault]
		}
		labelWidth := problemLabelWidth(shown)
		for _, p := range shown {
			t = t.Add(prettyProblemLine(p, g.base, labelWidth, r.Verbose)).NewLine()
		}
		if len(shown) < len(g.problems) {
			anyCapped = true
			t = t.Append(fmt.Sprintf("   … %d more · run ", len(g.problems)-len(shown)), styleMuted).
				Append("gavel status -v", styleRunning).NewLine()
		}
	}

	// Guarantee the hint appears even when no file was capped — the default
	// view still truncates messages to a single line, so `-v` shows more.
	if !r.Verbose && !anyCapped {
		t = t.Append(" run ", styleMuted).
			Append("gavel status -v", styleRunning).
			Append(" for full test/lint logs", styleMuted).NewLine()
	}
	return t
}

func prettyProblemLine(p Problem, base string, labelWidth int, verbose bool) api.Text {
	icon, style := problemIcon(p)
	t := clicky.Text("   ").Append(icon, style).Space().Append(p.Label, style)
	pad := max(labelWidth-runeLen(p.Label), 0)
	t = t.Append(strings.Repeat(" ", pad+2), "")

	if loc := problemLocation(base, p.Line); loc != "" {
		t = t.Append(loc, styleMuted).Append("  ", "")
	}

	if verbose {
		return appendMultilineMessage(t, p.Message)
	}
	msg := truncateRunes(firstLine(p.Message), maxProblemMessageRunes)
	if msg != "" {
		t = t.Append(msg, styleMuted)
	}
	return t
}

func problemIcon(p Problem) (string, string) {
	switch {
	case p.Kind == ProblemKindTest && p.Severity == "failed":
		return "✘", styleDeleted
	case p.Kind == ProblemKindLint && p.Severity == "error":
		return "⚠", styleDeleted
	case p.Kind == ProblemKindAI:
		return "⚠", styleError
	default:
		return "⚠", styleModified
	}
}

func problemLocation(base string, line int) string {
	if base == "" {
		return ""
	}
	if line > 0 {
		return fmt.Sprintf("%s:%d", base, line)
	}
	return base
}

func problemLabelWidth(problems []Problem) int {
	width := 0
	for _, p := range problems {
		if n := runeLen(p.Label); n > width {
			width = n
		}
	}
	return width
}

func appendMultilineMessage(t api.Text, message string) api.Text {
	lines := strings.Split(strings.TrimRight(message, "\n"), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return t
	}
	t = t.Append(lines[0], styleMuted)
	for _, line := range lines[1:] {
		t = t.NewLine().Append("     ", "").Append(line, styleMuted)
	}
	return t
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

func truncateRunes(s string, limit int) string {
	if runeLen(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit]) + "…"
}
