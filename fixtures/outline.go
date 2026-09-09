package fixtures

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type OutlineOptions struct {
	Paths   []string `json:"paths,omitempty"`
	WorkDir string   `json:"work_dir,omitempty"`
	Filter  string   `json:"filter,omitempty"`
}

type OutlineReport struct {
	WorkDir  string         `json:"work_dir,omitempty"`
	Files    int            `json:"files"`
	Fixtures int            `json:"fixtures"`
	Counts   map[string]int `json:"counts,omitempty"`
	Tree     []*OutlineNode `json:"tree,omitempty"`
}

type OutlineNode struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Kind       string         `json:"kind,omitempty"`
	OriginKind string         `json:"origin_kind,omitempty"`
	File       string         `json:"file,omitempty"`
	Line       int            `json:"line,omitempty"`
	Summary    string         `json:"summary,omitempty"`
	Fixtures   int            `json:"fixtures,omitempty"`
	Origin     *FixtureOrigin `json:"origin,omitempty"`
	Children   []*OutlineNode `json:"children,omitempty"`
}

func Outline(opts OutlineOptions) (*OutlineReport, error) {
	runner, err := NewRunner(RunnerOptions{
		Paths:   opts.Paths,
		WorkDir: opts.WorkDir,
		Filter:  opts.Filter,
	})
	if err != nil {
		return nil, err
	}

	tree, err := runner.Parse()
	if err != nil {
		return nil, err
	}

	return BuildOutline(tree, opts.WorkDir), nil
}

func BuildOutline(tree *FixtureNode, workDir string) *OutlineReport {
	report := &OutlineReport{
		WorkDir: workDir,
		Counts:  map[string]int{},
	}
	if tree == nil {
		return report
	}

	for _, child := range tree.Children {
		node := buildOutlineNode(child, report)
		if node != nil {
			report.Tree = append(report.Tree, node)
		}
	}
	return report
}

func (r *OutlineReport) Pretty() api.Text {
	t := clicky.Text(fmt.Sprintf("Fixture outline: %d fixtures in %d files", r.Fixtures, r.Files), "bold text-blue-500")
	if len(r.Counts) == 0 {
		return t
	}
	return t.Space().Append("("+formatOutlineCounts(r.Counts)+")", "text-muted")
}

func (r *OutlineReport) GetChildren() []api.TreeNode {
	children := make([]api.TreeNode, 0, len(r.Tree))
	for _, child := range r.Tree {
		children = append(children, child)
	}
	return children
}

func (n *OutlineNode) Pretty() api.Text {
	switch n.Type {
	case "file":
		t := clicky.Text(filepath.Base(n.Name), "text-cyan-600")
		if n.Fixtures > 0 {
			t = t.Space().Append(fmt.Sprintf("(%d fixtures)", n.Fixtures), "text-muted")
		}
		return t
	case "section", "table":
		t := clicky.Text(n.Name, "text-blue-500")
		if n.Type == "table" {
			t = clicky.Text(n.Name, "text-purple-500")
		}
		if n.Fixtures > 0 {
			t = t.Space().Append(fmt.Sprintf("(%d fixtures)", n.Fixtures), "text-muted")
		}
		return t
	}

	t := clicky.Text("")
	if n.Line > 0 {
		t = t.Append(fmt.Sprintf(":%d", n.Line), "text-muted").Space()
	}
	t = t.Append(n.Name, "bold wrap-space")
	if n.Kind != "" {
		t = t.Space().Append(n.Kind, "text-green-500")
	}
	if n.OriginKind != "" && n.OriginKind != n.Kind {
		t = t.Space().Append(n.OriginKind, "text-muted")
	}
	if n.Summary != "" {
		t = t.Space().Append("- "+n.Summary, "text-muted")
	}
	return t
}

func (n *OutlineNode) GetChildren() []api.TreeNode {
	children := make([]api.TreeNode, 0, len(n.Children))
	for _, child := range n.Children {
		children = append(children, child)
	}
	return children
}

func buildOutlineNode(node *FixtureNode, report *OutlineReport) *OutlineNode {
	if node == nil {
		return nil
	}

	out := &OutlineNode{
		Name:       node.Name,
		Type:       node.Type.String(),
		Origin:     node.Origin,
		OriginKind: originKind(node),
		File:       originFile(node),
		Line:       originLine(node),
	}
	if node.Test != nil {
		out.Kind = OutlineKind(*node.Test)
		out.Summary = outlineSummary(*node.Test)
		out.Children = outlineSyntheticChildren(*node.Test)
		report.Fixtures++
		report.Counts[out.Kind]++
		out.Fixtures = 1
		return out
	}

	if node.Type == FileNode {
		report.Files++
	}

	for _, child := range node.Children {
		childNode := buildOutlineNode(child, report)
		if childNode == nil {
			continue
		}
		out.Fixtures += childNode.Fixtures
		out.Children = append(out.Children, childNode)
	}
	if node.Type != FileNode && out.Fixtures == 0 {
		return nil
	}
	return out
}

func OutlineKind(fixture FixtureTest) string {
	switch {
	case fixture.IsAIStep():
		return "ai"
	case fixture.IsTestStep():
		return "test"
	case fixture.IsLintStep():
		return "lint"
	case fixture.Query != "":
		return "query"
	case !fixture.ExecBase().IsEmpty():
		return "exec"
	default:
		return "fixture"
	}
}

func outlineSummary(fixture FixtureTest) string {
	switch {
	case fixture.IsAIStep():
		if fixture.AIStep == nil {
			return "ai verification"
		}
		return fmt.Sprintf("%d criteria", len(fixture.AIStep.Criteria))
	case fixture.IsTestStep():
		return "test runner step"
	case fixture.IsLintStep():
		return "lint runner step"
	case fixture.Query != "":
		return "query fixture"
	}

	exec := fixture.ExecBase()
	if exec.Exec == "" {
		return ""
	}
	if len(exec.Args) == 0 {
		return fmt.Sprintf("exec %s", exec.Exec)
	}
	return fmt.Sprintf("exec %s (%d args)", exec.Exec, len(exec.Args))
}

func outlineSyntheticChildren(fixture FixtureTest) []*OutlineNode {
	if fixture.AIStep == nil {
		return nil
	}
	children := make([]*OutlineNode, 0, len(fixture.AIStep.Criteria))
	for i, criterion := range fixture.AIStep.Criteria {
		children = append(children, &OutlineNode{
			Name:    criterion.Text,
			Type:    "criterion",
			Kind:    "criterion",
			Summary: criterionSummary(i, criterion),
		})
	}
	return children
}

func criterionSummary(index int, criterion ChecklistItem) string {
	state := "unchecked"
	if criterion.Checked {
		state = "checked"
	}
	return fmt.Sprintf("#%d %s", index+1, state)
}

func formatOutlineCounts(counts map[string]int) string {
	keys := make([]string, 0, len(counts))
	for key, count := range counts {
		if count > 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", key, counts[key]))
	}
	return joinOutlineParts(parts)
}

func joinOutlineParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, part := range parts[1:] {
		out += ", " + part
	}
	return out
}

func originKind(node *FixtureNode) string {
	if node == nil || node.Origin == nil {
		return ""
	}
	return node.Origin.Kind
}

func originFile(node *FixtureNode) string {
	if node == nil || node.Origin == nil {
		return ""
	}
	return node.Origin.File
}

func originLine(node *FixtureNode) int {
	if node == nil || node.Origin == nil {
		return 0
	}
	return node.Origin.Line
}
