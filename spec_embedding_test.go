package gavel

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoAnonymousSpecEmbedding fails when any type embeds captain's api.Spec
// anonymously.
//
// api.Spec declares value-receiver MarshalJSON and MarshalYAML so it can omit
// its empty sections. Embedding it promotes both onto the embedder, where they
// shadow the default struct encoding and emit only the spec — silently dropping
// every field the embedder added. That has already happened three times: the
// dashboard's run payload lost ref/agent/mode/driver/runMode/plan/resume on the
// wire, and PromptSpec lost `file:` out of .gavel.yaml on every settings save.
//
// The failure is silent at the point of damage — a truncated config or a lossy
// request, not a compile error — so it is caught structurally here instead. Use
// a named field (`Spec api.Spec`), and give the type its own marshaller when the
// serialized shape has to stay flat.
//
// This walks source rather than registered types deliberately: a type nobody
// remembered to register is exactly the one that regresses.
func TestNoAnonymousSpecEmbedding(t *testing.T) {
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	var offenders []string
	fset := token.NewFileSet()

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			// A file this test cannot parse is not evidence of compliance.
			return err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			structType, ok := n.(*ast.StructType)
			if !ok || structType.Fields == nil {
				return true
			}
			for _, field := range structType.Fields.List {
				if len(field.Names) > 0 {
					continue // named field — the shape we want
				}
				if !isAPISpec(field.Type) {
					continue
				}
				rel, relErr := filepath.Rel(root, path)
				if relErr != nil {
					rel = path
				}
				offenders = append(offenders,
					rel+":"+fsetLine(fset, field.Pos()))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}

	if len(offenders) > 0 {
		t.Errorf("these types embed api.Spec anonymously, so its promoted "+
			"MarshalJSON/MarshalYAML will silently drop their own fields.\n"+
			"Use a named field (Spec api.Spec) instead:\n  %s",
			strings.Join(offenders, "\n  "))
	}
}

// isAPISpec reports whether an embedded field's type is api.Spec, in either the
// value or pointer form.
func isAPISpec(expr ast.Expr) bool {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Spec" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "api"
}

func fsetLine(fset *token.FileSet, pos token.Pos) string {
	return strings.TrimPrefix(fset.Position(pos).String(),
		fset.Position(pos).Filename+":")
}

func skipDir(name string) bool {
	switch name {
	case "node_modules", "dist", "testdata", "vendor", ".git", ".tmp", ".bin", "hack":
		return true
	}
	return false
}
