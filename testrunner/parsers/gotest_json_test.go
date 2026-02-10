package parsers

import (
	"bytes"
	"strings"
	"testing"
)

func TestGoTestJSONParseStreamSimple(t *testing.T) {
	input := strings.NewReader(`{"Time":"2025-01-01T12:00:00Z","Action":"run","Package":"./foo","Test":"TestFoo"}
{"Time":"2025-01-01T12:00:00Z","Action":"pass","Package":"./foo","Test":"TestFoo","Elapsed":0.1}`)

	parser := NewGoTestJSON("")
	var stdout bytes.Buffer

	pass, fail, err := parser.ParseStream(input, &stdout, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pass != 1 {
		t.Errorf("expected 1 pass, got %d", pass)
	}
	if fail != 0 {
		t.Errorf("expected 0 fails, got %d", fail)
	}
}

func TestGoTestJSONParseStreamWithFailure(t *testing.T) {
	input := strings.NewReader(`{"Time":"2025-01-01T12:00:00Z","Action":"run","Package":"./foo","Test":"TestFail"}
{"Time":"2025-01-01T12:00:00Z","Action":"fail","Package":"./foo","Test":"TestFail","Output":"--- FAIL: TestFail\n    main_test.go:10: expected true, got false"}`)

	parser := NewGoTestJSON("")
	var stdout bytes.Buffer

	pass, fail, err := parser.ParseStream(input, &stdout, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pass != 0 {
		t.Errorf("expected 0 pass, got %d", pass)
	}
	if fail != 1 {
		t.Errorf("expected 1 fail, got %d", fail)
	}
}

func TestGoTestJSONParse(t *testing.T) {
	input := `{"Time":"2025-01-01T12:00:00Z","Action":"run","Package":"./foo","Test":"TestExample"}
{"Time":"2025-01-01T12:00:00Z","Action":"fail","Package":"./foo","Test":"TestExample","Output":"--- FAIL: TestExample\n    example_test.go:15: expected 42, got 24"}
{"Time":"2025-01-01T12:00:00Z","Action":"run","Package":"./foo","Test":"TestOther"}
{"Time":"2025-01-01T12:00:00Z","Action":"pass","Package":"./foo","Test":"TestOther"}`

	parser := NewGoTestJSON("")
	results, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 tests, got %d", len(results))
	}

	// Find failed and passed tests
	var failedTest, passedTest *Test
	for i := range results {
		if results[i].Name == "TestExample" {
			failedTest = &results[i]
		} else if results[i].Name == "TestOther" {
			passedTest = &results[i]
		}
	}

	if failedTest == nil {
		t.Fatal("expected to find TestExample")
	}
	if !failedTest.Failed {
		t.Errorf("expected TestExample to be marked as Failed")
	}
	if failedTest.Package != "./foo" {
		t.Errorf("expected package ./foo, got %s", failedTest.Package)
	}
	if failedTest.Framework != GoTest {
		t.Errorf("expected GoTest framework, got %v", failedTest.Framework)
	}

	if passedTest == nil {
		t.Fatal("expected to find TestOther")
	}
	if passedTest.Failed {
		t.Errorf("expected TestOther to not be marked as Failed")
	}
	if passedTest.Skipped {
		t.Errorf("expected TestOther to not be marked as Skipped")
	}
}

func TestGoTestJSONParseBuildFailure(t *testing.T) {
	input := `{"ImportPath":"github.com/example/pkg","Action":"build-output","Output":"# github.com/example/pkg\n"}
{"ImportPath":"github.com/example/pkg","Action":"build-output","Output":"main.go:10:5: undefined: someFunc\n"}
{"ImportPath":"github.com/example/pkg","Action":"build-output","Output":"main.go:20:3: syntax error: unexpected newline\n"}
{"ImportPath":"github.com/example/pkg","Action":"build-fail"}
{"Time":"2025-01-01T12:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2025-01-01T12:00:00Z","Action":"output","Package":"github.com/example/pkg","Output":"FAIL\tgithub.com/example/pkg [build failed]\n"}
{"Time":"2025-01-01T12:00:00Z","Action":"fail","Package":"github.com/example/pkg","Elapsed":0}`

	parser := NewGoTestJSON("")
	results, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 test result (build failure), got %d", len(results))
	}

	buildFailure := results[0]
	if buildFailure.Name != "Build Failed" {
		t.Errorf("expected test name 'Build Failed', got %s", buildFailure.Name)
	}
	if !buildFailure.Failed {
		t.Errorf("expected build failure to be marked as Failed")
	}
	if buildFailure.Package != "github.com/example/pkg" {
		t.Errorf("expected package github.com/example/pkg, got %s", buildFailure.Package)
	}

	expectedMsg := "# github.com/example/pkg\nmain.go:10:5: undefined: someFunc\nmain.go:20:3: syntax error: unexpected newline\n"
	if buildFailure.Message != expectedMsg {
		t.Errorf("expected build error message:\n%s\ngot:\n%s", expectedMsg, buildFailure.Message)
	}
}

func TestGoTestJSONParseNoTestFiles(t *testing.T) {
	input := `{"Time":"2025-01-01T12:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2025-01-01T12:00:00Z","Action":"output","Package":"github.com/example/pkg","Output":"?   \tgithub.com/example/pkg\t[no test files]\n"}
{"Time":"2025-01-01T12:00:00Z","Action":"skip","Package":"github.com/example/pkg","Elapsed":0}`

	parser := NewGoTestJSON("")
	results, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 test result (package skip), got %d", len(results))
	}

	packageSkip := results[0]
	if packageSkip.Name != "No test files" {
		t.Errorf("expected test name 'No test files', got %s", packageSkip.Name)
	}
	if !packageSkip.Skipped {
		t.Errorf("expected package to be marked as Skipped")
	}
	if packageSkip.Failed {
		t.Errorf("expected package skip to not be marked as Failed")
	}
	if packageSkip.Package != "github.com/example/pkg" {
		t.Errorf("expected package github.com/example/pkg, got %s", packageSkip.Package)
	}
	if packageSkip.Message != "[no test files]" {
		t.Errorf("expected message '[no test files]', got %s", packageSkip.Message)
	}
}

func TestGoTestJSONParserName(t *testing.T) {
	parser := NewGoTestJSON("")
	if parser.Name() != "go test json" {
		t.Errorf("expected name 'go test json', got %s", parser.Name())
	}
}

func TestTruncateFailure(t *testing.T) {
	tests := map[string]struct {
		input    string
		maxLen   int
		expected string
	}{
		"simple":     {"hello world", 20, "hello world"},
		"multiline":  {"hello\nworld", 20, "hello world"},
		"spaces":     {"hello  \t  world", 20, "hello world"},
		"truncate":   {"hello world", 8, "hello..."},
		"multitrunc": {"hello\nworld", 8, "hello..."},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := truncateFailure(tc.input, tc.maxLen)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestGoTestJSONParseFiltersGinkgoBootstrap(t *testing.T) {
	// Simulate go test output that includes both a Ginkgo bootstrap test and a real test
	input := `{"Time":"2025-01-01T12:00:00Z","Action":"run","Package":"./foo","Test":"TestFixtures"}
{"Time":"2025-01-01T12:00:00Z","Action":"pass","Package":"./foo","Test":"TestFixtures","Elapsed":0.001}
{"Time":"2025-01-01T12:00:00Z","Action":"run","Package":"./foo","Test":"TestRealTest"}
{"Time":"2025-01-01T12:00:00Z","Action":"pass","Package":"./foo","Test":"TestRealTest","Elapsed":0.1}`

	// Create parser with a location map that marks TestFixtures as a Ginkgo bootstrap
	parser := &GoTestJSON{
		LocationMap: map[string]TestLocation{
			"TestFixtures": {File: "fixtures_suite_test.go", Line: 10, IsGinkgoBootstrap: true},
			"TestRealTest": {File: "real_test.go", Line: 5, IsGinkgoBootstrap: false},
		},
	}

	results, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have 1 result - TestRealTest (TestFixtures should be filtered)
	if len(results) != 1 {
		t.Errorf("expected 1 test (bootstrap filtered), got %d", len(results))
		for _, r := range results {
			t.Logf("  - %s", r.Name)
		}
	}

	if len(results) > 0 && results[0].Name != "TestRealTest" {
		t.Errorf("expected TestRealTest, got %s", results[0].Name)
	}
}
