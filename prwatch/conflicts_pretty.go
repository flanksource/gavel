package prwatch

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/github"
)

// prettyConflicts explains a CONFLICTING PR. The header line already reports
// `Mergeable: CONFLICTING (conflicts with base)`; this section says *what*
// conflicts and how to clear it, which is the part no GitHub API returns.
func (r PRWatchResult) prettyConflicts() api.Text {
	if r.Conflicts == nil {
		return clicky.Text("")
	}
	return clicky.Text("").Add(treeText(&conflictSection{report: r.Conflicts}))
}

type conflictSection struct {
	report *github.MergeConflictReport
}

func (s *conflictSection) Pretty() api.Text {
	text := clicky.Text("").Add(icons.Fail).Space().
		Append("Merge conflicts", "font-bold text-red-600").
		Append(" ", "").
		Append(s.report.BaseRefName, "text-cyan-600").
		Append(" ← ", "text-gray-500").
		Append(s.report.HeadRefName, "text-cyan-600")
	if n := len(s.report.Files); n > 0 {
		suffix := " files"
		if n == 1 {
			suffix = " file"
		}
		text = text.Append(fmt.Sprintf(" (%d%s)", n, suffix), "text-gray-500")
	}
	return text
}

func (s *conflictSection) GetChildren() []api.TreeNode {
	var children []api.TreeNode
	if s.report.Unavailable != "" {
		children = append(children, &conflictNoteNode{note: s.report.Unavailable})
	}
	for _, file := range s.report.Files {
		children = append(children, &conflictFileNode{file: file})
	}
	if commands := s.report.ResolveCommands(); len(commands) > 0 {
		children = append(children, &commandsNode{label: "Resolve locally", commands: commands})
	}
	return children
}

type conflictFileNode struct {
	file github.MergeConflict
}

func (n *conflictFileNode) Pretty() api.Text {
	text := clicky.Text(n.file.Path, "text-red-500")
	if n.file.Kind != "" {
		text = text.Append(" ("+n.file.Kind+")", "text-gray-500")
	}
	return text
}

func (n *conflictFileNode) GetChildren() []api.TreeNode { return nil }

// conflictNoteNode carries the reason the conflicting paths could not be listed.
// Rendered amber rather than red: the merge is still blocked, but this line is
// about gavel's own reach, not about the PR.
type conflictNoteNode struct {
	note string
}

func (n *conflictNoteNode) Pretty() api.Text {
	return clicky.Text("").Add(icons.Warning).Space().Append(n.note, "text-yellow-600")
}

func (n *conflictNoteNode) GetChildren() []api.TreeNode { return nil }
