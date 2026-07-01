package outline

import (
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func (r *Report) Pretty() api.Text {
	leaves := r.Leaves()
	files := map[string]bool{}
	for _, leaf := range leaves {
		files[leaf.File] = true
	}
	t := clicky.Text(fmt.Sprintf("Test outline: %d tests in %d files", len(leaves), len(files)), "bold text-blue-500")
	if r.RunCount > 0 {
		t = t.Space().Append(fmt.Sprintf("(history from %d runs)", r.RunCount), "text-muted")
	} else {
		t = t.Space().Append("(no run history)", "text-muted")
	}
	return t
}

func (r *Report) GetChildren() []api.TreeNode {
	byDir := map[string][]*Entry{}
	var dirs []string
	for _, entry := range r.Entries {
		dir := filepath.ToSlash(filepath.Dir(entry.File))
		if _, ok := byDir[dir]; !ok {
			dirs = append(dirs, dir)
		}
		byDir[dir] = append(byDir[dir], entry)
	}
	sort.Strings(dirs)

	children := make([]api.TreeNode, 0, len(dirs))
	for _, dir := range dirs {
		children = append(children, dirNode(dir, byDir[dir]))
	}
	return children
}

type outlineGroup struct {
	label    string
	style    string
	leafCt   int
	children []api.TreeNode
}

func (g *outlineGroup) Pretty() api.Text {
	t := clicky.Text(g.label, g.style)
	if g.leafCt > 0 {
		t = t.Space().Append(fmt.Sprintf("(%d tests)", g.leafCt), "text-muted")
	}
	return t
}

func (g *outlineGroup) GetChildren() []api.TreeNode { return g.children }

func dirNode(dir string, entries []*Entry) api.TreeNode {
	byFile := map[string][]*Entry{}
	var files []string
	for _, entry := range entries {
		if _, ok := byFile[entry.File]; !ok {
			files = append(files, entry.File)
		}
		byFile[entry.File] = append(byFile[entry.File], entry)
	}
	sort.Strings(files)

	var children []api.TreeNode
	leafCt := 0
	for _, file := range files {
		node := fileNode(file, byFile[file])
		leafCt += node.leafCt
		children = append(children, node)
	}
	return &outlineGroup{label: dir + "/", style: "bold", leafCt: leafCt, children: children}
}

func fileNode(file string, entries []*Entry) *outlineGroup {
	var children []api.TreeNode
	leafCt := 0
	for _, entry := range entries {
		children = append(children, &entryNode{entry: entry})
		leafCt += countLeaves(entry)
	}
	return &outlineGroup{label: filepath.Base(file), style: "text-cyan-600", leafCt: leafCt, children: children}
}

func countLeaves(e *Entry) int {
	if e.Error != "" {
		return 0
	}
	if len(e.Children) == 0 && !e.Container {
		return 1
	}
	count := 0
	for _, child := range e.Children {
		count += countLeaves(child)
	}
	return count
}

type entryNode struct {
	entry *Entry
}

func (n *entryNode) GetChildren() []api.TreeNode {
	children := make([]api.TreeNode, 0, len(n.entry.Children))
	for _, child := range n.entry.Children {
		children = append(children, &entryNode{entry: child})
	}
	return children
}

func (n *entryNode) Pretty() api.Text {
	e := n.entry
	if e.Error != "" {
		t := clicky.Text(e.Name, "bold text-red-600")
		return t.Space().Append("— "+e.Error, "text-red-600")
	}
	if e.Container {
		t := clicky.Text(e.Name, "bold text-muted")
		return appendBadges(t, e)
	}

	t := clicky.Text("", "")
	if e.Line > 0 {
		t = t.Append(fmt.Sprintf(":%d", e.Line), "text-muted")
		t = t.Space()
	}
	t = t.Append(e.Name, "bold wrap-space")
	if e.SizeLines > 0 {
		t = t.Space().Append(fmt.Sprintf("%dL", e.SizeLines), "text-muted")
	}
	if e.Complexity > 0 {
		t = t.Space().Append(fmt.Sprintf("cx %d", e.Complexity), complexityStyle(e.Complexity))
	}
	if e.DuplicationPct > 0 {
		t = t.Space().Append(fmt.Sprintf("dup %.0f%%", e.DuplicationPct), duplicationStyle(e.DuplicationPct))
	}
	t = appendBadges(t, e)
	t = t.Space().Append(historySegment(e), historyStyle(e))
	if desc := e.AISummary; desc != "" {
		t = t.Space().Append("— "+desc, "text-muted italic")
	} else if e.Description != "" && e.Framework == parsers.GoTest {
		// Ginkgo/vitest descriptions are the suite chain, already visible as
		// the tree path; repeating them per row is noise. JSON keeps them.
		t = t.Space().Append("— "+e.Description, "text-muted")
	}
	return t
}

func appendBadges(t api.Text, e *Entry) api.Text {
	if e.Dynamic {
		t = t.Space().Append("dynamic", "text-yellow-600")
	}
	if e.Pending {
		t = t.Space().Append("pending", "text-yellow-600")
	}
	if e.Focused {
		t = t.Space().Append("FOCUSED", "text-red-600 bold")
	}
	return t
}

func historySegment(e *Entry) string {
	h := e.History
	if h == nil {
		return "[no runs]"
	}
	return fmt.Sprintf("[pass %.0f%% · %d runs · avg %s]",
		h.PassRate*100, h.ExecutionCount, h.AvgDuration.Round(time.Millisecond))
}

func historyStyle(e *Entry) string {
	if e.History == nil {
		return "text-muted"
	}
	switch {
	case e.History.PassRate >= 1:
		return "text-green-600"
	case e.History.PassRate <= 0:
		return "text-red-600"
	default:
		return "text-yellow-600"
	}
}

func complexityStyle(complexity int) string {
	switch {
	case complexity > 10:
		return "text-red-600"
	case complexity > 5:
		return "text-yellow-600"
	default:
		return "text-muted"
	}
}

func duplicationStyle(pct float64) string {
	switch {
	case pct > 50:
		return "text-red-600"
	case pct > 25:
		return "text-yellow-600"
	default:
		return "text-muted"
	}
}
