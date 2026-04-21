package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

// lintSummaryView is a TreeNode wrapper that renders []*LinterResult as a
// compact tree: root -> linter -> rule -> example locations (capped).
type lintSummaryView struct {
	Results []*linters.LinterResult `json:"results"`
	Limit   int                     `json:"-"`
}

func newLintSummaryView(results []*linters.LinterResult, limit int) *lintSummaryView {
	if limit < 1 {
		limit = 5
	}
	return &lintSummaryView{Results: results, Limit: limit}
}

func (s *lintSummaryView) Pretty() api.Text {
	violations, skipped := 0, 0
	for _, r := range s.Results {
		if r == nil {
			continue
		}
		if r.Skipped {
			skipped++
			continue
		}
		violations += len(r.Violations)
	}
	t := api.Text{}.Append(fmt.Sprintf("Lint summary: %d violations", violations), "text-blue-500")
	if skipped > 0 {
		t = t.Append(fmt.Sprintf(" (%d linters skipped)", skipped), "text-muted")
	}
	return t
}

func (s *lintSummaryView) GetChildren() []api.TreeNode {
	// Aggregate across multiple results for the same linter (the per-project
	// fan-out may produce more than one result per linter name).
	type linterBucket struct {
		linter     string
		workDir    string
		violations []models.Violation
		skipped    bool
		skipReason string
		errorMsg   string // non-skip failure (e.g. eslint config error)
		command    string // argv[0] captured at invocation
		args       []string
	}
	byLinter := map[string]*linterBucket{}
	var order []string
	for _, r := range s.Results {
		if r == nil {
			continue
		}
		b, ok := byLinter[r.Linter]
		if !ok {
			b = &linterBucket{linter: r.Linter, workDir: r.WorkDir}
			byLinter[r.Linter] = b
			order = append(order, r.Linter)
		}
		if r.Skipped {
			b.skipped = true
			b.skipReason = r.Error
			continue
		}
		b.skipped = false
		b.violations = append(b.violations, r.Violations...)
		if b.workDir == "" {
			b.workDir = r.WorkDir
		}
		if !r.Success && r.Error != "" && b.errorMsg == "" {
			b.errorMsg = r.Error
		}
		if b.command == "" && r.Command != "" {
			b.command = r.Command
			b.args = r.Args
		}
	}
	sort.Strings(order)

	var children []api.TreeNode
	for _, name := range order {
		b := byLinter[name]
		children = append(children, &linterSummaryNode{
			linter:     b.linter,
			workDir:    b.workDir,
			violations: b.violations,
			skipped:    b.skipped && len(b.violations) == 0 && b.errorMsg == "",
			skipReason: b.skipReason,
			errorMsg:   b.errorMsg,
			command:    b.command,
			args:       b.args,
			limit:      s.Limit,
		})
	}
	return children
}

type linterSummaryNode struct {
	linter     string
	workDir    string
	violations []models.Violation
	skipped    bool
	skipReason string
	errorMsg   string
	command    string
	args       []string
	limit      int
}

func (n *linterSummaryNode) Pretty() api.Text {
	if n.skipped {
		return api.Text{}.
			Append("⊘ ", "text-muted").
			Append(n.linter, "text-muted").
			Append(" (skipped: "+n.skipReason+")", "text-muted")
	}
	if n.errorMsg != "" {
		summary := firstErrorLine(n.errorMsg)
		t := api.Text{}.
			Append("❌ ", "text-red-600").
			Append(n.linter, "text-red-600")
		if count := len(n.violations); count > 0 {
			t = t.Append(fmt.Sprintf(" (%d violations, error)", count), "text-muted")
		} else {
			t = t.Append(" (error)", "text-muted")
		}
		if summary != "" {
			t = t.Append(" — " + summary)
		}
		return t
	}
	count := len(n.violations)
	if count == 0 {
		return api.Text{}.
			Append("✅ ", "text-green-600").
			Append(n.linter, "text-green-600")
	}
	return api.Text{}.
		Append("⚠️ ", "text-yellow-600").
		Append(n.linter, "text-yellow-600").
		Append(fmt.Sprintf(" (%d violations)", count), "text-muted")
}

func (n *linterSummaryNode) GetChildren() []api.TreeNode {
	if n.skipped {
		return nil
	}
	var children []api.TreeNode
	// For errored runs, show the argv alongside the error so users can see
	// exactly what was invoked (and copy-paste it to reproduce).
	if n.errorMsg != "" && n.command != "" {
		children = append(children, &linterCommandNode{command: n.command, args: n.args, workDir: n.workDir})
	}
	if n.errorMsg != "" {
		children = append(children, &linterErrorNode{message: n.errorMsg})
	}
	if len(n.violations) == 0 {
		return children
	}
	// Group by rule (Rule.Method). Violations without a rule fall into a
	// single "(no rule)" bucket so they remain visible.
	type ruleGroup struct {
		rule       string
		violations []models.Violation
	}
	byRule := map[string]*ruleGroup{}
	var order []string
	for _, v := range n.violations {
		key := "(no rule)"
		if v.Rule != nil && v.Rule.Method != "" {
			key = v.Rule.Method
		}
		g, ok := byRule[key]
		if !ok {
			g = &ruleGroup{rule: key}
			byRule[key] = g
			order = append(order, key)
		}
		g.violations = append(g.violations, v)
	}
	sort.Slice(order, func(i, j int) bool {
		return len(byRule[order[i]].violations) > len(byRule[order[j]].violations)
	})

	for _, key := range order {
		g := byRule[key]
		children = append(children, &ruleSummaryNode{
			rule:       g.rule,
			workDir:    n.workDir,
			violations: g.violations,
			limit:      n.limit,
		})
	}
	return children
}

// linterCommandNode renders the exact argv the linter was invoked with, plus
// the cwd. Shown next to a failing linter so the user can see and reproduce
// the call that produced the error.
type linterCommandNode struct {
	command string
	args    []string
	workDir string
}

func (n *linterCommandNode) Pretty() api.Text {
	line := n.command
	if len(n.args) > 0 {
		line = line + " " + strings.Join(n.args, " ")
	}
	t := api.Text{}.Append("$ ", "text-muted").Append(line, "text-blue-400")
	if n.workDir != "" {
		t = t.NewLine().Append("  cwd: "+n.workDir, "text-muted")
	}
	return t
}

func (n *linterCommandNode) GetChildren() []api.TreeNode { return nil }

// linterErrorNode renders a multi-line execution error as a single tree leaf.
// The full text is preserved; each line appears on its own row so clicky's
// tree renderer keeps vertical alignment under the parent.
type linterErrorNode struct {
	message string
}

func (n *linterErrorNode) Pretty() api.Text {
	t := api.Text{}
	lines := strings.Split(strings.TrimRight(n.message, "\n"), "\n")
	for i, line := range lines {
		if i > 0 {
			t = t.NewLine()
		}
		t = t.Append(line, "text-red-500")
	}
	return t
}

func (n *linterErrorNode) GetChildren() []api.TreeNode { return nil }

// firstErrorLine returns the first non-empty line of msg, stripped of trailing
// whitespace. Used for one-line error previews next to the linter name.
func firstErrorLine(msg string) string {
	for _, line := range strings.Split(msg, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

type ruleSummaryNode struct {
	rule       string
	workDir    string
	violations []models.Violation
	limit      int
}

func (n *ruleSummaryNode) Pretty() api.Text {
	t := api.Text{}.Append(n.rule, "text-blue-500").
		Append(fmt.Sprintf(" (%d)", len(n.violations)), "text-muted")
	if msg := firstMessage(n.violations); msg != "" {
		t = t.Append(" — " + msg)
	}
	return t
}

func (n *ruleSummaryNode) GetChildren() []api.TreeNode {
	// Collapse to one entry per file so a rule that fires many times in the
	// same file doesn't drown out other affected files within the limit.
	type fileBucket struct {
		file  string
		first models.Violation
		count int
	}
	byFile := map[string]*fileBucket{}
	var order []string
	for _, v := range n.violations {
		b, ok := byFile[v.File]
		if !ok {
			b = &fileBucket{file: v.File, first: v}
			byFile[v.File] = b
			order = append(order, v.File)
		}
		b.count++
	}

	limit := min(n.limit, len(order))
	children := make([]api.TreeNode, 0, limit)
	for i := range limit {
		b := byFile[order[i]]
		children = append(children, &locationSummaryNode{
			workDir:   n.workDir,
			violation: b.first,
			count:     b.count,
		})
	}
	if remaining := len(order) - limit; remaining > 0 {
		children = append(children, &moreLocationsNode{remaining: remaining})
	}
	return children
}

type locationSummaryNode struct {
	workDir   string
	violation models.Violation
	count     int // number of violations of this rule in this file
}

func (n *locationSummaryNode) Pretty() api.Text {
	file := n.violation.File
	if n.workDir != "" && filepath.IsAbs(file) {
		if rel, err := filepath.Rel(n.workDir, file); err == nil {
			file = rel
		}
	}
	t := api.Text{}.Append("📄 ", "")
	if n.count > 1 {
		t = t.Append(file, "text-muted").
			Append(fmt.Sprintf(" (%d)", n.count), "text-muted")
		return t
	}
	loc := file
	if n.violation.Line > 0 {
		loc = fmt.Sprintf("%s:%d", file, n.violation.Line)
		if n.violation.Column > 0 {
			loc = fmt.Sprintf("%s:%d", loc, n.violation.Column)
		}
	}
	return t.Append(loc, "text-muted")
}

func (n *locationSummaryNode) GetChildren() []api.TreeNode { return nil }

type moreLocationsNode struct {
	remaining int
}

func (n *moreLocationsNode) Pretty() api.Text {
	return api.Text{}.Append(fmt.Sprintf("… %d more", n.remaining), "text-muted")
}

func (n *moreLocationsNode) GetChildren() []api.TreeNode { return nil }

func firstMessage(vs []models.Violation) string {
	for _, v := range vs {
		if v.Message != nil && *v.Message != "" {
			return *v.Message
		}
	}
	return ""
}
