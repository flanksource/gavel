package testui

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// defaultLintGrouping mirrors the web export default (linter → rule → file) so
// the CLI tree and the UI render lint findings with the same hierarchy.
const defaultLintGrouping = "linter-rule-file"

// sectionNode is a top-level grouping node ("Tests", "Lint") in the exported
// tree. It exists so `gavel test --format html/markdown` renders a nested tree
// instead of clicky concatenating every suite flat: the section header carries
// a summary while its children hold the per-suite / per-finding ViewNode trees
// already built for the web UI.
type sectionNode struct {
	label    string
	style    string
	summary  api.Text
	children []*ViewNode
}

func (s *sectionNode) Pretty() api.Text {
	text := clicky.Text(s.label, s.style)
	if !s.summary.IsEmpty() {
		text = text.Space().Add(s.summary)
	}
	return text
}

func (s *sectionNode) GetChildren() []api.TreeNode {
	children := make([]api.TreeNode, 0, len(s.children))
	for _, child := range s.children {
		children = append(children, child)
	}
	return children
}

// GetChildren makes Snapshot a TreeNode so clicky renders a nested tree for
// non-pretty formats (html, markdown, …). Sections with no content are omitted
// so a test-only run doesn't emit an empty "Lint" header and vice versa.
func (s Snapshot) GetChildren() []api.TreeNode {
	sections := make([]api.TreeNode, 0, 2)

	if testRoots := s.testViewRoots(); len(testRoots) > 0 {
		summary := parsers.Tests(s.Tests).Sum()
		sections = append(sections, &sectionNode{
			label:    "Tests",
			style:    "bold text-blue-500",
			summary:  summary.Pretty(),
			children: testRoots,
		})
	}

	if lintRoots := buildLintViewNodes(s.Lint, lintRouteFilters{Grouping: defaultLintGrouping}); len(lintRoots) > 0 {
		sections = append(sections, &sectionNode{
			label:    "Lint",
			style:    "bold text-blue-500",
			summary:  clicky.Text(fmt.Sprintf("(%d findings)", lintNodeViolationCount(lintRoots)), "text-muted"),
			children: lintRoots,
		})
	}

	return sections
}

// testViewRoots converts the snapshot's test tree into the same ViewNode roots
// the web export builds, so suites/specs nest identically across CLI and UI.
func (s Snapshot) testViewRoots() []*ViewNode {
	roots := make([]*ViewNode, 0, len(s.Tests))
	for _, test := range s.Tests {
		roots = append(roots, testToViewNode(test))
	}
	return roots
}
