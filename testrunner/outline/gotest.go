package outline

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// extractGoTests returns one entry per Test* function in a plain (non-ginkgo)
// test file, with t.Run subtests as children named the way `go test` names
// them at runtime (spaces become underscores) so they join against history.
func extractGoTests(fset *token.FileSet, file *ast.File, relPath string) []*Entry {
	var entries []*Entry
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !isGoTestFunc(fn) || parsers.ContainsRunSpecs(fn) {
			continue
		}
		entry := &Entry{
			Framework:  parsers.GoTest,
			File:       relPath,
			Name:       fn.Name.Name,
			Line:       fset.Position(fn.Pos()).Line,
			EndLine:    fset.Position(fn.End()).Line,
			Complexity: CountComplexity(fn.Body),
			calls:      collectCalls(fn.Body, testingParamName(fn)),
		}
		entry.SizeLines = entry.EndLine - entry.Line + 1
		entry.Children = goSubtests(fset, relPath, fn.Name.Name, fn.Body, testingParamName(fn))
		entries = append(entries, entry)
	}
	return entries
}

// isGoTestFunc mirrors `go test` selection: a top-level Test* function whose
// sole parameter is *testing.T.
func isGoTestFunc(fn *ast.FuncDecl) bool {
	if fn.Recv != nil || fn.Body == nil || fn.Name.Name == "TestMain" {
		return false
	}
	if !strings.HasPrefix(fn.Name.Name, "Test") {
		return false
	}
	params := fn.Type.Params
	if params == nil || len(params.List) != 1 || len(params.List[0].Names) > 1 {
		return false
	}
	star, ok := params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "T"
}

func testingParamName(fn *ast.FuncDecl) string {
	names := fn.Type.Params.List[0].Names
	if len(names) == 0 {
		return ""
	}
	return names[0].Name
}

// goSubtests finds tName.Run(...) calls in body. Literal subtest names become
// child entries; non-literal names yield a single child flagged Dynamic.
func goSubtests(fset *token.FileSet, relPath, prefix string, body ast.Node, tName string) []*Entry {
	if tName == "" {
		return nil
	}
	var children []*Entry
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isRunCallOn(call, tName) || len(call.Args) < 2 {
			return true
		}
		child := &Entry{
			Framework: parsers.GoTest,
			File:      relPath,
			Line:      fset.Position(call.Pos()).Line,
			EndLine:   fset.Position(call.End()).Line,
		}
		if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			name, err := strconv.Unquote(lit.Value)
			if err != nil {
				name = strings.Trim(lit.Value, `"`)
			}
			child.Name = prefix + "/" + rewriteSubtestName(name)
		} else {
			child.Name = prefix + "/<dynamic>"
			child.Dynamic = true
		}
		if fnLit, ok := call.Args[1].(*ast.FuncLit); ok {
			child.Complexity = CountComplexity(fnLit.Body)
			child.calls = collectCalls(fnLit.Body, funcLitParamName(fnLit))
			child.Children = goSubtests(fset, relPath, child.Name, fnLit.Body, funcLitParamName(fnLit))
		}
		child.SizeLines = child.EndLine - child.Line + 1
		children = append(children, child)
		return false // children of this t.Run are handled by the recursion above
	})
	return children
}

func isRunCallOn(call *ast.CallExpr, tName string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Run" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == tName
}

func funcLitParamName(fn *ast.FuncLit) string {
	params := fn.Type.Params
	if params == nil || len(params.List) == 0 || len(params.List[0].Names) == 0 {
		return ""
	}
	return params.List[0].Names[0].Name
}

// rewriteSubtestName applies the same transformation `go test` applies to
// subtest names when building the runtime test name.
func rewriteSubtestName(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}

// collectCalls gathers called function names from a test body (e.g. "Build",
// "history.Load"), skipping calls on the *testing.T param; used for static
// descriptions.
func collectCalls(body ast.Node, tName string) []string {
	var calls []string
	seen := map[string]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := callName(call, tName)
		if name != "" && !seen[name] {
			seen[name] = true
			calls = append(calls, name)
		}
		return true
	})
	return calls
}

func callName(call *ast.CallExpr, tName string) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		ident, ok := fun.X.(*ast.Ident)
		if !ok || ident.Name == tName {
			return ""
		}
		return ident.Name + "." + fun.Sel.Name
	}
	return ""
}
