package testrunner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/testrunner/runners"
)

func TestRegistryDefaultFactory(t *testing.T) {
	tmpDir := t.TempDir()
	registry := DefaultRegistry(tmpDir)

	if registry == nil {
		t.Fatal("expected non-nil registry")
	}

	// Check that both runners are registered
	goTestRunner, ok := registry.Get(GoTest)
	if !ok {
		t.Error("expected GoTest runner to be registered")
	}
	if goTestRunner == nil {
		t.Error("expected non-nil GoTest runner")
	}

	ginkgoRunner, ok := registry.Get(Ginkgo)
	if !ok {
		t.Error("expected Ginkgo runner to be registered")
	}
	if ginkgoRunner == nil {
		t.Error("expected non-nil Ginkgo runner")
	}
}

func TestRegistryDetectAll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create go test files
	if err := os.WriteFile(filepath.Join(tmpDir, "example_test.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	registry := DefaultRegistry(tmpDir)
	frameworks, err := registry.DetectAll()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frameworks) == 0 {
		t.Error("expected at least one framework to be detected")
	}

	// Should detect GoTest at minimum
	found := false
	for _, fw := range frameworks {
		if fw == GoTest {
			found = true
		}
	}
	if !found {
		t.Error("expected GoTest to be detected")
	}
}

func TestRegistryGet(t *testing.T) {
	registry := DefaultRegistry("/tmp")

	runner, ok := registry.Get(GoTest)
	if !ok {
		t.Error("expected GoTest runner to be found")
	}
	if runner == nil {
		t.Error("expected non-nil runner")
	}
	if runner.Name() != GoTest {
		t.Errorf("expected runner name GoTest, got %v", runner.Name())
	}
}

func TestRegistryGetParser(t *testing.T) {
	registry := DefaultRegistry("/tmp")

	parser, ok := registry.GetParser("go test json")
	if !ok {
		t.Error("expected parser to be found")
	}
	if parser == nil {
		t.Error("expected non-nil parser")
	}
	if parser.Name() != "go test json" {
		t.Errorf("expected parser name 'go test json', got %s", parser.Name())
	}
}

func TestRegistryRegister(t *testing.T) {
	registry := NewRegistry("/tmp")

	// Register runners manually
	goTestRunner := runners.NewGoTest("/tmp")
	ginkgoRunner := runners.NewGinkgo("/tmp")

	registry.Register(goTestRunner)
	registry.Register(ginkgoRunner)

	// Verify both are registered
	_, ok := registry.Get(GoTest)
	if !ok {
		t.Error("expected GoTest runner to be registered")
	}

	_, ok = registry.Get(Ginkgo)
	if !ok {
		t.Error("expected Ginkgo runner to be registered")
	}
}
