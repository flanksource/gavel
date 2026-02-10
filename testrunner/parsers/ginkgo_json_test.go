package parsers

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGinkgoJSONParserName(t *testing.T) {
	parser := NewGinkgoJSON()
	if parser.Name() != "ginkgo json" {
		t.Errorf("expected 'ginkgo json', got %s", parser.Name())
	}
}

func TestGinkgoJSONParsePassedTests(t *testing.T) {
	suites := []ginkgoSuiteReport{
		{
			SuitePath:        "/path/to/cmd",
			SuiteDescription: "CMD Suite",
			SpecReports: []ginkgoSpecReport{
				{
					ContainerHierarchyTexts: []string{"AST Edit CLI", "ast edit rename"},
					LeafNodeText:            "should rename a method",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/path/to/cmd/ast_edit_test.go",
						LineNumber: 73,
					},
					State:   "passed",
					RunTime: int64(10 * time.Millisecond),
				},
			},
		},
	}

	data, err := json.Marshal(suites)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	parser := NewGinkgoJSON()
	tests, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(tests) != 1 {
		t.Errorf("expected 1 test, got %d", len(tests))
	}

	test := tests[0]
	if test.Name != "should rename a method" {
		t.Errorf("expected name 'should rename a method', got %s", test.Name)
	}
	expectedSuite := []string{"AST Edit CLI", "ast edit rename"}
	if len(test.Suite) != len(expectedSuite) {
		t.Errorf("expected suite length %d, got %d", len(expectedSuite), len(test.Suite))
	}
	for i, expectedVal := range expectedSuite {
		if i >= len(test.Suite) {
			t.Errorf("expected suite[%d] '%s', but index out of range", i, expectedVal)
		} else if test.Suite[i] != expectedVal {
			t.Errorf("expected suite[%d] '%s', got '%s'", i, expectedVal, test.Suite[i])
		}
	}
	if test.File != "/path/to/cmd/ast_edit_test.go" {
		t.Errorf("expected file '/path/to/cmd/ast_edit_test.go', got %s", test.File)
	}
	if test.Line != 73 {
		t.Errorf("expected line 73, got %d", test.Line)
	}
	if test.Failed {
		t.Error("expected test to be passed")
	}
	if test.Skipped {
		t.Error("expected test to not be skipped")
	}
	if test.Framework != Ginkgo {
		t.Errorf("expected framework Ginkgo, got %v", test.Framework)
	}
}

func TestGinkgoJSONParseFailedTests(t *testing.T) {
	suites := []ginkgoSuiteReport{
		{
			SuitePath:        "/path/to/cmd",
			SuiteDescription: "CMD Suite",
			SpecReports: []ginkgoSpecReport{
				{
					ContainerHierarchyTexts: []string{"Generate TypeScript"},
					LeafNodeText:            "should generate files",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/path/to/cmd/generate_test.go",
						LineNumber: 49,
					},
					State:   "failed",
					RunTime: int64(50 * time.Millisecond),
					Failure: &ginkgoFailure{
						Message: "Expected error to occur",
						Location: ginkgoLocation{
							FileName:   "/path/to/cmd/generate_test.go",
							LineNumber: 69,
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(suites)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	parser := NewGinkgoJSON()
	tests, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(tests) != 1 {
		t.Errorf("expected 1 test, got %d", len(tests))
	}

	test := tests[0]
	if !test.Failed {
		t.Error("expected test to be failed")
	}
	if test.Skipped {
		t.Error("expected test to not be skipped")
	}
	if !strings.Contains(test.Message, "Expected error to occur") {
		t.Errorf("expected message to contain failure text, got %s", test.Message)
	}
	// File and Line should be LeafNodeLocation (where test is defined), not Failure.Location
	// This is correct for table-driven tests where each entry has unique location
	if test.File != "/path/to/cmd/generate_test.go" {
		t.Errorf("expected file to be leaf node location, got %s", test.File)
	}
	if test.Line != 49 {
		t.Errorf("expected line to be leaf node line 49, got %d", test.Line)
	}
}

func TestGinkgoJSONParseSkippedTests(t *testing.T) {
	suites := []ginkgoSuiteReport{
		{
			SuitePath:        "/path/to/pkg",
			SuiteDescription: "Test Suite",
			SpecReports: []ginkgoSpecReport{
				{
					ContainerHierarchyTexts: []string{"Some Tests"},
					LeafNodeText:            "should skip this test",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/path/to/pkg/test_test.go",
						LineNumber: 100,
					},
					State:   "skipped",
					RunTime: 0,
				},
			},
		},
	}

	data, err := json.Marshal(suites)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	parser := NewGinkgoJSON()
	tests, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(tests) != 1 {
		t.Errorf("expected 1 test, got %d", len(tests))
	}

	test := tests[0]
	if test.Failed {
		t.Error("expected test to not be failed")
	}
	if !test.Skipped {
		t.Error("expected test to be skipped")
	}
}

func TestGinkgoJSONParseMultipleTests(t *testing.T) {
	suites := []ginkgoSuiteReport{
		{
			SuitePath:        "/path/to/cmd",
			SuiteDescription: "CMD Suite",
			SpecReports: []ginkgoSpecReport{
				{
					ContainerHierarchyTexts: []string{"AST Edit"},
					LeafNodeText:            "test 1",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/path/to/cmd/test1_test.go",
						LineNumber: 10,
					},
					State:   "passed",
					RunTime: 5000000,
				},
				{
					ContainerHierarchyTexts: []string{"AST Edit"},
					LeafNodeText:            "test 2",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/path/to/cmd/test2_test.go",
						LineNumber: 20,
					},
					State:   "failed",
					RunTime: 10000000,
					Failure: &ginkgoFailure{
						Message: "Test failed",
						Location: ginkgoLocation{
							FileName:   "/path/to/cmd/test2_test.go",
							LineNumber: 25,
						},
					},
				},
				{
					ContainerHierarchyTexts: []string{"AST Edit"},
					LeafNodeText:            "test 3",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/path/to/cmd/test3_test.go",
						LineNumber: 30,
					},
					State:   "skipped",
					RunTime: 0,
				},
			},
		},
	}

	data, err := json.Marshal(suites)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	parser := NewGinkgoJSON()
	tests, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(tests) != 3 {
		t.Errorf("expected 3 tests, got %d", len(tests))
	}

	// Check test 1 (passed)
	if tests[0].Failed || tests[0].Skipped {
		t.Error("expected test 1 to be passed")
	}

	// Check test 2 (failed)
	if !tests[1].Failed || tests[1].Skipped {
		t.Error("expected test 2 to be failed")
	}

	// Check test 3 (skipped)
	if !tests[2].Skipped || tests[2].Failed {
		t.Error("expected test 3 to be skipped")
	}
}

func TestGinkgoJSONParseContainerHierarchy(t *testing.T) {
	suites := []ginkgoSuiteReport{
		{
			SuitePath:        "/path/to/cmd",
			SuiteDescription: "CMD Suite",
			SpecReports: []ginkgoSpecReport{
				{
					ContainerHierarchyTexts: []string{"Outer Describe", "Inner Context"},
					LeafNodeText:            "nested test",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/path/to/cmd/nested_test.go",
						LineNumber: 50,
					},
					State:   "passed",
					RunTime: 1000000,
				},
			},
		},
	}

	data, err := json.Marshal(suites)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	parser := NewGinkgoJSON()
	tests, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	test := tests[0]
	expectedSuite := []string{"Outer Describe", "Inner Context"}
	if len(test.Suite) != len(expectedSuite) {
		t.Errorf("expected suite length %d, got %d", len(expectedSuite), len(test.Suite))
	}
	for i, expectedVal := range expectedSuite {
		if i >= len(test.Suite) {
			t.Errorf("expected suite[%d] '%s', but index out of range", i, expectedVal)
		} else if test.Suite[i] != expectedVal {
			t.Errorf("expected suite[%d] '%s', got '%s'", i, expectedVal, test.Suite[i])
		}
	}
}

func TestGinkgoJSONParsePackageExtraction(t *testing.T) {
	suites := []ginkgoSuiteReport{
		{
			SuitePath:        "/home/user/project/pkg/testrunner",
			SuiteDescription: "TestRunner Suite",
			SpecReports: []ginkgoSpecReport{
				{
					ContainerHierarchyTexts: []string{"Test"},
					LeafNodeText:            "should work",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/home/user/project/pkg/testrunner/runner_test.go",
						LineNumber: 10,
					},
					State:   "passed",
					RunTime: 1000000,
				},
			},
		},
	}

	data, err := json.Marshal(suites)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	parser := NewGinkgoJSON()
	tests, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	test := tests[0]
	if test.Package != "testrunner" {
		t.Errorf("expected package 'testrunner', got %s", test.Package)
	}
}

func TestGinkgoJSONParseEmptySuite(t *testing.T) {
	suites := []ginkgoSuiteReport{
		{
			SuitePath:        "/path/to/cmd",
			SuiteDescription: "Empty Suite",
			SpecReports:      []ginkgoSpecReport{},
		},
	}

	data, err := json.Marshal(suites)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	parser := NewGinkgoJSON()
	tests, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(tests) != 0 {
		t.Errorf("expected 0 tests, got %d", len(tests))
	}
}

func TestGinkgoJSONParseDuration(t *testing.T) {
	suites := []ginkgoSuiteReport{
		{
			SuitePath:        "/path/to/cmd",
			SuiteDescription: "Test Suite",
			SpecReports: []ginkgoSpecReport{
				{
					ContainerHierarchyTexts: []string{"Tests"},
					LeafNodeText:            "duration test",
					LeafNodeLocation: ginkgoLocation{
						FileName:   "/path/to/cmd/test_test.go",
						LineNumber: 10,
					},
					State:   "passed",
					RunTime: 50_000_000, // 50 milliseconds in nanoseconds
				},
			},
		},
	}

	data, err := json.Marshal(suites)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	parser := NewGinkgoJSON()
	tests, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	test := tests[0]
	expectedDuration := time.Duration(50_000_000)
	if test.Duration != expectedDuration {
		t.Errorf("expected duration %v, got %v", expectedDuration, test.Duration)
	}
}
