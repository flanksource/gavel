package runners

import (
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"github.com/flanksource/commons/logger"
)

// hasGinkgoImports reports whether the given Go file imports Ginkgo (v1 or
// v2). Uses the AST (ImportsOnly) so string literals that mention the import
// path in non-test code or other test files don't false-positive. Fails
// closed: parse errors return false so a malformed file is treated as
// "not ginkgo" — the Go runner will surface the parse error itself.
//
// Shared between Ginkgo.Detect/DiscoverPackages and GoTest's Ginkgo exclusion
// logic so the two agree on what "imports ginkgo" means.
func hasGinkgoImports(path string) bool {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		logger.V(2).Infof("ginkgo detect: parse %s: %v", path, err)
		return false
	}
	for _, imp := range file.Imports {
		if imp.Path == nil {
			continue
		}
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if importPath == "github.com/onsi/ginkgo" ||
			importPath == "github.com/onsi/ginkgo/v2" ||
			strings.HasPrefix(importPath, "github.com/onsi/ginkgo/v2/") {
			return true
		}
	}
	return false
}
