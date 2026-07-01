package outline

import (
	"go/ast"
	"go/token"
)

// CountComplexity returns the cyclomatic complexity of n: 1 plus the number
// of branching constructs (if, for, range, non-default case/comm clauses,
// and && / || operators).
func CountComplexity(n ast.Node) int {
	complexity := 1
	ast.Inspect(n, func(node ast.Node) bool {
		switch v := node.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			if v.List != nil {
				complexity++
			}
		case *ast.CommClause:
			if v.Comm != nil {
				complexity++
			}
		case *ast.BinaryExpr:
			if v.Op == token.LAND || v.Op == token.LOR {
				complexity++
			}
		}
		return true
	})
	return complexity
}
