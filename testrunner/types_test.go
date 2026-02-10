package testrunner

import (
	"testing"
)

func TestFrameworkString(t *testing.T) {
	tests := []struct {
		name     string
		f        Framework
		expected string
	}{
		{
			name:     "GoTest framework",
			f:        GoTest,
			expected: "go test",
		},
		{
			name:     "Ginkgo framework",
			f:        Ginkgo,
			expected: "ginkgo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.f.String()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTestFailureStructure(t *testing.T) {
	failure := TestFailure{
		Name:      "TestUserLogin",
		Package:   "github.com/flanksource/gavel/auth",
		Message:   "expected nil, got error",
		File:      "pkg/auth/user_test.go",
		Line:      42,
		Framework: GoTest,
	}

	if failure.Name != "TestUserLogin" {
		t.Errorf("Name: got %q, want %q", failure.Name, "TestUserLogin")
	}
	if failure.Line != 42 {
		t.Errorf("Line: got %d, want %d", failure.Line, 42)
	}
	if failure.Framework != GoTest {
		t.Errorf("Framework: got %v, want %v", failure.Framework, GoTest)
	}
}

func TestGinkgoTestFailureStructure(t *testing.T) {
	failure := TestFailure{
		Name:      "should validate email",
		Package:   "github.com/flanksource/gavel/config",
		Suite:     []string{"Config Parser"},
		Message:   "Expected <nil> to match error",
		File:      "pkg/config/parser_test.go",
		Line:      78,
		Framework: Ginkgo,
	}

	expectedSuite := []string{"Config Parser"}
	if len(failure.Suite) != len(expectedSuite) {
		t.Errorf("Suite length: got %d, want %d", len(failure.Suite), len(expectedSuite))
	}
	for i, expectedVal := range expectedSuite {
		if i >= len(failure.Suite) || failure.Suite[i] != expectedVal {
			t.Errorf("Suite[%d]: got %q, want %q", i, failure.Suite[i], expectedVal)
		}
	}
	if failure.Framework != Ginkgo {
		t.Errorf("Framework: got %v, want %v", failure.Framework, Ginkgo)
	}
}
