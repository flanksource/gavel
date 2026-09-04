package outline_test

import (
	"go/ast"
	"go/parser"
	"go/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/outline"
)

func parseFuncBody(src string) ast.Node {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", "package p\nfunc f() {\n"+src+"\n}", 0)
	Expect(err).NotTo(HaveOccurred())
	fn := file.Decls[0].(*ast.FuncDecl)
	return fn.Body
}

var _ = Describe("CountComplexity", func() {
	DescribeTable("counts branching constructs",
		func(src string, expected int) {
			Expect(outline.CountComplexity(parseFuncBody(src))).To(Equal(expected))
		},
		Entry("straight-line code", `x := 1; _ = x`, 1),
		Entry("single if", `if true { return }`, 2),
		Entry("if with && condition", `x := 1; if x > 0 && x < 2 { return }`, 3),
		Entry("for loop with nested if", `for i := 0; i < 3; i++ { if i == 1 { break } }`, 3),
		Entry("range loop", `for range []int{1} { }`, 2),
		Entry("switch with two cases and default", `switch 1 { case 1: case 2: default: }`, 3),
		Entry("select with one comm clause and default", `ch := make(chan int, 1)
select { case <-ch: default: }`, 2),
		Entry("|| operator", `x := true || false; _ = x`, 2),
	)
})
