package archtest

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/tools/go/packages"
)

var _ = Describe("guarded recursive removal", func() {
	It("uses os.RemoveAll only inside safefs.SafeRemove", func() {
		root, err := filepath.Abs(filepath.Join("..", ".."))
		Expect(err).NotTo(HaveOccurred())
		pkgs, err := packages.Load(&packages.Config{
			Dir:   root,
			Tests: true,
			Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		}, "./...")
		Expect(err).NotTo(HaveOccurred())
		Expect(loadErrors(root, pkgs)).To(BeEmpty())
		var offenders []string
		for _, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				ast.Inspect(file, func(node ast.Node) bool {
					ident, ok := node.(*ast.Ident)
					if !ok {
						return true
					}
					fn, ok := pkg.TypesInfo.Uses[ident].(*types.Func)
					if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "os" || fn.Name() != "RemoveAll" {
						return true
					}
					pos := pkg.Fset.Position(ident.Pos())
					if relative(root, pos.Filename) != filepath.Join("internal", "safefs", "remove.go") || !insideSafeRemove(file, ident.Pos()) {
						offenders = append(offenders, fmt.Sprintf("%s: use safefs.SafeRemove", position(root, pkg.Fset, ident.Pos())))
					}
					return true
				})
			}
		}
		sort.Strings(offenders)
		Expect(offenders).To(BeEmpty())
	})
})

func insideSafeRemove(file *ast.File, pos token.Pos) bool {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == "SafeRemove" && fn.Body != nil && fn.Body.Pos() <= pos && pos <= fn.Body.End() {
			return true
		}
	}
	return false
}
