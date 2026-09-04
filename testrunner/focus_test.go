package testrunner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestExtractNativeFocus(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		focus     string
		remaining []string
		wantErr   string
	}{
		{name: "go spaced", args: []string{"-run", "TestFoo", "-count=1"}, focus: "TestFoo", remaining: []string{"-count=1"}},
		{name: "go equals", args: []string{"-run=TestFoo"}, focus: "TestFoo"},
		{name: "ginkgo spaced", args: []string{"--focus", "does work", "--label-filter", "smoke"}, focus: "does work", remaining: []string{"--label-filter", "smoke"}},
		{name: "ginkgo equals", args: []string{"--focus=does work"}, focus: "does work"},
		{name: "raw args stay opaque", args: []string{"--label-filter", "-run"}, remaining: []string{"--label-filter", "-run"}},
		{name: "missing go pattern", args: []string{"-run"}, wantErr: "non-empty focus pattern"},
		{name: "empty ginkgo pattern", args: []string{"--focus="}, wantErr: "non-empty focus pattern"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			focus, remaining, err := extractNativeFocus(tt.args)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("extractNativeFocus(%v) error = %v, want %q", tt.args, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("extractNativeFocus(%v): %v", tt.args, err)
			}
			if focus != tt.focus || !reflect.DeepEqual(remaining, tt.remaining) {
				t.Fatalf("extractNativeFocus(%v) = (%q, %v), want (%q, %v)", tt.args, focus, remaining, tt.focus, tt.remaining)
			}
		})
	}
}

func TestResolveFrameworkArgsMapsFocusAndSkipsUnsupported(t *testing.T) {
	o := &TestOrchestrator{
		RunOptions: RunOptions{PassThroughArgs: []string{"--focus", "TestFoo"}},
		registry:   DefaultRegistry(t.TempDir()),
	}

	frameworks, argsByFramework, err := o.resolveFrameworkArgs(
		[]Framework{parsers.GoTest, parsers.Ginkgo, parsers.Jest},
		[]string{"-count=1"},
	)
	if err != nil {
		t.Fatalf("resolveFrameworkArgs: %v", err)
	}
	if !reflect.DeepEqual(frameworks, []Framework{parsers.GoTest, parsers.Ginkgo}) {
		t.Fatalf("frameworks = %v, want go test + ginkgo", frameworks)
	}
	if got, want := argsByFramework[parsers.GoTest], []string{"-count=1", "-run", "TestFoo"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("go test args = %v, want %v", got, want)
	}
	if got, want := argsByFramework[parsers.Ginkgo], []string{"-count=1", "--focus", "TestFoo"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ginkgo args = %v, want %v", got, want)
	}
	if _, ok := argsByFramework[parsers.Jest]; ok {
		t.Fatalf("jest should be excluded from a focused run: %v", argsByFramework)
	}
}

func TestResolveFrameworkArgsPassThroughRequiresSingleFramework(t *testing.T) {
	o := &TestOrchestrator{
		RunOptions: RunOptions{PassThroughArgs: []string{"--label-filter", "smoke"}},
		registry:   DefaultRegistry(t.TempDir()),
	}

	_, _, err := o.resolveFrameworkArgs([]Framework{parsers.GoTest, parsers.Ginkgo}, nil)
	if err == nil || !strings.Contains(err.Error(), "exactly one selected framework") {
		t.Fatalf("expected ambiguous pass-through error, got %v", err)
	}

	frameworks, argsByFramework, err := o.resolveFrameworkArgs([]Framework{parsers.Ginkgo}, []string{"--trace"})
	if err != nil {
		t.Fatalf("single-framework pass-through: %v", err)
	}
	if !reflect.DeepEqual(frameworks, []Framework{parsers.Ginkgo}) {
		t.Fatalf("frameworks = %v", frameworks)
	}
	if got, want := argsByFramework[parsers.Ginkgo], []string{"--trace", "--label-filter", "smoke"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ginkgo args = %v, want %v", got, want)
	}
}

func TestResolveFrameworkArgsErrorsWhenNoRunnerSupportsFocus(t *testing.T) {
	o := &TestOrchestrator{
		RunOptions: RunOptions{PassThroughArgs: []string{"-run=TestFoo"}},
		registry:   DefaultRegistry(t.TempDir()),
	}

	_, _, err := o.resolveFrameworkArgs([]Framework{parsers.Jest, parsers.Vitest}, nil)
	if err == nil || !strings.Contains(err.Error(), "not supported by any detected framework") {
		t.Fatalf("expected unsupported focus error, got %v", err)
	}
}

func TestApplyFailedFilterReturnsPerFrameworkFocusPatterns(t *testing.T) {
	snapshot := testui.Snapshot{Tests: []parsers.Test{
		{Name: "TestFoo", PackagePath: "./go", Framework: parsers.GoTest, Failed: true},
		{Name: "does work", PackagePath: "./ginkgo", Framework: parsers.Ginkgo, Failed: true},
	}}
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	path := filepath.Join(t.TempDir(), "failed.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}

	filtered, focusByFramework, err := applyFailedFilter(map[Framework][]string{
		parsers.GoTest: {"./go", "./passing"},
		parsers.Ginkgo: {"./ginkgo"},
	}, path)
	if err != nil {
		t.Fatalf("applyFailedFilter: %v", err)
	}
	if got, want := filtered[parsers.GoTest], []string{"./go"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("filtered go packages = %v, want %v", got, want)
	}
	if got, want := filtered[parsers.Ginkgo], []string{"./ginkgo"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("filtered ginkgo packages = %v, want %v", got, want)
	}
	if got, want := focusByFramework[parsers.GoTest], "^(TestFoo)$"; got != want {
		t.Fatalf("go focus = %q, want %q", got, want)
	}
	if got, want := focusByFramework[parsers.Ginkgo], "does work"; got != want {
		t.Fatalf("ginkgo focus = %q, want %q", got, want)
	}
}

func TestApplyFailedFilterRejectsFailedPackagesMissingFromDiscovery(t *testing.T) {
	snapshot := testui.Snapshot{Tests: []parsers.Test{{
		Name:        "fails in CI",
		PackagePath: "./ci-only",
		Framework:   parsers.Ginkgo,
		Failed:      true,
	}}}
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	failedPath := filepath.Join(t.TempDir(), "failed.json")
	if err := os.WriteFile(failedPath, data, 0o644); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}

	_, _, err = applyFailedFilter(map[Framework][]string{
		parsers.Ginkgo: {"./local"},
	}, failedPath)
	if err == nil || !strings.Contains(err.Error(), "did not match any detected packages") {
		t.Fatalf("expected unmatched failed-package error, got %v", err)
	}
}
