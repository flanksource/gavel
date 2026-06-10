package outline

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/flanksource/gavel/testrunner/parsers"
	ginkgooutline "github.com/onsi/ginkgo/v2/ginkgo/outline"
)

// dynamicName replaces ginkgo's "undefined" marker for specs/containers whose
// text is not a string literal (e.g. It(fmt.Sprintf(...))).
const dynamicName = "<dynamic>"

// ginkgoOutlineNode mirrors the JSON shape produced by ginkgo's own
// `ginkgo outline --format json`. We call outline.FromASTFile as a library
// (the exact code behind that command) instead of shelling out, and round-trip
// through JSON because the returned type is unexported. The shape is pinned by
// the ginkgo version in go.mod.
type ginkgoOutlineNode struct {
	Name    string               `json:"name"`
	Text    string               `json:"text"`
	Start   int                  `json:"start"`
	End     int                  `json:"end"`
	Spec    bool                 `json:"spec"`
	Focused bool                 `json:"focused"`
	Pending bool                 `json:"pending"`
	Labels  []string             `json:"labels"`
	Nodes   []*ginkgoOutlineNode `json:"nodes"`
}

// extractGinkgoTests outlines a ginkgo test file: containers
// (Describe/Context/When/DescribeTable) become Container entries, specs
// (It/Specify/Entry) become leaves. Size and complexity come from the spec's
// closure body located via byte offsets in the same parsed AST.
func extractGinkgoTests(fset *token.FileSet, file *ast.File, relPath string) ([]*Entry, error) {
	o, err := ginkgooutline.FromASTFile(fset, file)
	if err != nil {
		return nil, fmt.Errorf("ginkgo outline for %s: %w", relPath, err)
	}
	data, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("marshal ginkgo outline for %s: %w", relPath, err)
	}
	var nodes []*ginkgoOutlineNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("decode ginkgo outline for %s: %w", relPath, err)
	}

	conv := &ginkgoConverter{
		fset:      fset,
		tokenFile: fset.File(file.Pos()),
		relPath:   relPath,
		calls:     indexCallsByOffset(fset, file),
	}
	var entries []*Entry
	for _, node := range nodes {
		entries = append(entries, conv.entry(node, nil))
	}
	return entries, nil
}

type ginkgoConverter struct {
	fset      *token.FileSet
	tokenFile *token.File
	relPath   string
	calls     map[[2]int]*ast.CallExpr
}

func (c *ginkgoConverter) entry(node *ginkgoOutlineNode, suite []string) *Entry {
	e := &Entry{
		Framework: parsers.Ginkgo,
		File:      c.relPath,
		Name:      node.Text,
		Suite:     append([]string(nil), suite...),
		Container: !node.Spec,
		Pending:   node.Pending,
		Focused:   node.Focused,
		Labels:    node.Labels,
		Line:      c.lineForOffset(node.Start),
		EndLine:   c.lineForOffset(node.End - 1),
	}
	if node.Text == "undefined" {
		e.Name = dynamicName
		e.Dynamic = true
	}

	if call := c.calls[[2]int{node.Start, node.End}]; call != nil && node.Spec {
		if body := trailingFuncLit(call); body != nil {
			start := c.fset.Position(body.Pos()).Line
			end := c.fset.Position(body.End()).Line
			e.Line, e.EndLine = start, end
			e.SizeLines = end - start + 1
			e.Complexity = CountComplexity(body.Body)
			e.calls = collectCalls(body.Body, "")
		} else {
			// DescribeTable Entry rows have no closure of their own.
			e.SizeLines = e.EndLine - e.Line + 1
		}
	}

	childSuite := append(append([]string(nil), suite...), e.Name)
	for _, child := range node.Nodes {
		e.Children = append(e.Children, c.entry(child, childSuite))
	}
	return e
}

func (c *ginkgoConverter) lineForOffset(offset int) int {
	if offset < 0 || offset >= c.tokenFile.Size() {
		return 0
	}
	return c.tokenFile.Line(c.tokenFile.Pos(offset))
}

// indexCallsByOffset maps (start, end) byte offsets to call expressions so
// outline nodes can be matched back to their AST call.
func indexCallsByOffset(fset *token.FileSet, file *ast.File) map[[2]int]*ast.CallExpr {
	calls := map[[2]int]*ast.CallExpr{}
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			start := fset.PositionFor(call.Pos(), false).Offset
			end := fset.PositionFor(call.End(), false).Offset
			calls[[2]int{start, end}] = call
		}
		return true
	})
	return calls
}

func trailingFuncLit(call *ast.CallExpr) *ast.FuncLit {
	for i := len(call.Args) - 1; i >= 0; i-- {
		if fn, ok := call.Args[i].(*ast.FuncLit); ok {
			return fn
		}
	}
	return nil
}
