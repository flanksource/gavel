package prwatch

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func (r PRWatchResult) prettyGavelResults() api.Text {
	if len(r.GavelResults) == 0 {
		return clicky.Text("")
	}
	text := clicky.Text("Gavel Results:", "font-bold")
	for _, summary := range r.GavelResults {
		if summary != nil {
			text = text.NewLine().Add(treeText(summary))
		}
	}
	return text
}

// treeText adapts a clicky TreeNode into an inline TextTree so a nested tree
// can be embedded inside a Text block. clicky exposes Tree() for constructing
// one but has no adapter from the TreeNode interface, and splicing in a
// pre-rendered MustFormat string would leak ANSI into String()/Markdown()/HTML().
func treeText(node api.TreeNode) api.TextTree {
	children := node.GetChildren()
	branches := make([]api.TextTree, 0, len(children))
	for _, child := range children {
		branches = append(branches, treeText(child))
	}
	return clicky.Tree(node.Pretty(), branches...)
}

// Pretty renders the shard header: status, sticky id, and the same pass/fail
// roll-up `gavel test` prints at the end of a run.
func (s GavelResultsSummary) Pretty() api.Text {
	text := clicky.Text("").Add(s.statusIcon()).Space().Append(s.label(), "font-bold")
	if s.TestsTotal > 0 {
		text = text.Space().Add(s.testSummary().Pretty())
	}
	if s.BenchRegressions > 0 {
		text = text.Space().Add(clicky.KeyValue(" benchmark regressions", s.BenchRegressions, "text-red-500"))
	}
	if s.Error != "" {
		text = text.NewLine().Append(s.Error, "text-yellow-600")
		if s.ExitCode != nil {
			text = text.Append(fmt.Sprintf(" (exit %d)", *s.ExitCode), "text-gray-500")
		}
	}
	return text
}

// GetChildren hangs the failing tests, the lint tree, and the artifact link
// beneath the header. Both detail sections delegate to the renderers `gavel
// test` and `gavel lint` use, so PR status cannot drift from a local run.
func (s GavelResultsSummary) GetChildren() []api.TreeNode {
	var children []api.TreeNode
	if len(s.Failures) > 0 {
		children = append(children, &testFailuresNode{failures: s.Failures, total: s.TestsFailed})
	}
	if len(s.Lint) > 0 {
		children = append(children, s.lintSection())
	}
	if s.ArtifactURL != "" {
		children = append(children, &artifactLinkNode{url: s.ArtifactURL})
	}
	return children
}

func (s GavelResultsSummary) testSummary() parsers.TestSummary {
	return parsers.TestSummary{
		Total:    s.TestsTotal,
		Passed:   s.TestsPassed,
		Failed:   s.TestsFailed,
		Skipped:  s.TestsSkipped,
		Duration: s.Duration,
	}
}

func (s GavelResultsSummary) lintSection() api.TreeNode {
	view := linters.NewSummaryView(s.Lint, 0)
	view.Total = s.LintViolations
	return view
}

func (s GavelResultsSummary) statusIcon() icons.Icon {
	switch {
	case s.Error != "":
		return icons.Warning
	case s.TestsFailed > 0 || s.BenchRegressions > 0:
		return icons.Fail
	case s.LintViolations > 0 || s.hasLintFailure():
		return icons.Warning
	default:
		return icons.Pass
	}
}

// hasLintFailure reports whether a linter died rather than finding nothing. A
// linter that could not run produces zero violations, so without this a broken
// lint shard would render as a clean pass.
func (s GavelResultsSummary) hasLintFailure() bool {
	for _, result := range s.Lint {
		if result.TimedOut || (!result.Success && result.Error != "") {
			return true
		}
	}
	return false
}

func (s GavelResultsSummary) label() string {
	if s.StickyID != "" {
		return s.StickyID
	}
	return fmt.Sprintf("artifact %d", s.ArtifactID)
}

// testFailuresNode mirrors the "Test failures (N)" section `gavel test` prints,
// with every leaf rendered by parsers.Test's own renderer.
type testFailuresNode struct {
	failures []parsers.Test
	total    int
}

func (n *testFailuresNode) Pretty() api.Text {
	return clicky.Text(fmt.Sprintf("Test failures (%d)", n.total), "bold text-red-600")
}

func (n *testFailuresNode) GetChildren() []api.TreeNode {
	children := make([]api.TreeNode, 0, len(n.failures)+1)
	for _, failure := range n.failures {
		children = append(children, failure)
	}
	if dropped := n.total - len(n.failures); dropped > 0 {
		children = append(children, &moreNode{remaining: dropped})
	}
	return children
}

type moreNode struct {
	remaining int
}

func (n *moreNode) Pretty() api.Text {
	return clicky.Text(fmt.Sprintf("… %d more", n.remaining), "text-muted")
}

func (n *moreNode) GetChildren() []api.TreeNode { return nil }

type artifactLinkNode struct {
	url string
}

func (n *artifactLinkNode) Pretty() api.Text {
	return clicky.Text("View full results: "+n.url, "text-blue-600")
}

func (n *artifactLinkNode) GetChildren() []api.TreeNode { return nil }
