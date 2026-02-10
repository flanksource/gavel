package parsers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestContainsRunSpecs(t *testing.T) {
	tests := map[string]struct {
		code     string
		expected bool
	}{
		"ginkgo_bootstrap": {
			code: `package foo
func TestFixtures(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fixtures Suite")
}`,
			expected: true,
		},
		"regular_test": {
			code: `package foo
func TestSomething(t *testing.T) {
	if 1 != 1 {
		t.Error("math is broken")
	}
}`,
			expected: false,
		},
		"test_with_subtests": {
			code: `package foo
func TestSubtests(t *testing.T) {
	t.Run("subtest", func(t *testing.T) {
		t.Log("hello")
	})
}`,
			expected: false,
		},
		"empty_body": {
			code: `package foo
func TestEmpty(t *testing.T) {}`,
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "test.go", tc.code, 0)
			if err != nil {
				t.Fatalf("failed to parse code: %v", err)
			}

			var fn *ast.FuncDecl
			for _, decl := range node.Decls {
				if f, ok := decl.(*ast.FuncDecl); ok {
					fn = f
					break
				}
			}
			if fn == nil {
				t.Fatal("no function found")
			}

			result := containsRunSpecs(fn)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestBuildTestLocationMapDetectsGinkgoBootstrap(t *testing.T) {
	// Use actual test files in this repo that contain Ginkgo bootstrap tests
	locations, err := BuildTestLocationMap("../../fixtures")
	if err != nil {
		t.Fatalf("failed to build location map: %v", err)
	}

	// TestFixtures in fixtures/fixtures_suite_test.go should be detected as Ginkgo bootstrap
	if loc, ok := locations["TestFixtures"]; ok {
		if !loc.IsGinkgoBootstrap {
			t.Errorf("expected TestFixtures to be detected as Ginkgo bootstrap")
		}
	} else {
		t.Errorf("expected to find TestFixtures in location map")
	}
}
