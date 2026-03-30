package parsers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"

	"github.com/flanksource/gavel/utils"
)

type TestLocation struct {
	File              string
	Line              int
	IsGinkgoBootstrap bool
}

// BuildTestLocationMap scans Go test files and builds a map of test names to locations.
// The key format is "TestName" for top-level tests.
func BuildTestLocationMap(dir string) (map[string]TestLocation, error) {
	locations := make(map[string]TestLocation)

	err := utils.WalkGitIgnored(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil // Skip unparseable files
		}

		ast.Inspect(node, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if strings.HasPrefix(fn.Name.Name, "Test") {
				pos := fset.Position(fn.Pos())
				locations[fn.Name.Name] = TestLocation{
					File:              pos.Filename,
					Line:              pos.Line,
					IsGinkgoBootstrap: containsRunSpecs(fn),
				}
			}
			return true
		})
		return nil
	})

	return locations, err
}

// containsRunSpecs checks if a function body contains a call to RunSpecs,
// which indicates a Ginkgo bootstrap test function.
func containsRunSpecs(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "RunSpecs" {
			found = true
			return false
		}
		return true
	})
	return found
}
