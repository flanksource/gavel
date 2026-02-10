package testrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
)

func TestPathFiltering(t *testing.T) {
	wd, _ := os.Getwd()
	projectRoot := filepath.Join(wd, "..")

	tests := []struct {
		name          string
		startingPaths []string
		expectedPkgs  []string
		unexpectedPkg string
	}{
		{
			name:          "Filter to parsers directory only",
			startingPaths: []string{"./testrunner/parsers"},
			expectedPkgs:  []string{"./testrunner/parsers"},
			unexpectedPkg: "./claudehistory",
		},
		{
			name:          "Filter to claudehistory directory only",
			startingPaths: []string{"./claudehistory"},
			expectedPkgs:  []string{"./claudehistory"},
			unexpectedPkg: "./testrunner/parsers",
		},
		{
			name:          "Multiple starting paths",
			startingPaths: []string{"./testrunner/parsers", "./claudehistory"},
			expectedPkgs:  []string{"./testrunner/parsers", "./claudehistory"},
			unexpectedPkg: "./todos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry(projectRoot)
			goTestRunner := runners.NewGoTest(projectRoot)
			registry.Register(goTestRunner)

			orchestrator := &TestOrchestrator{
				RunOptions: RunOptions{
					WorkDir:       projectRoot,
					StartingPaths: tt.startingPaths,
				},
				registry: registry,
			}

			runner, ok := registry.Get(parsers.GoTest)
			if !ok {
				t.Fatalf("Failed to get GoTest runner from registry")
			}

			packages, err := orchestrator.discoverPackagesInPaths(runner, tt.startingPaths)
			if err != nil {
				t.Fatalf("discoverPackagesInPaths failed: %v", err)
			}

			for _, expectedPkg := range tt.expectedPkgs {
				found := false
				for _, pkg := range packages {
					if strings.HasPrefix(pkg, expectedPkg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected package prefix %q not found in discovered packages: %v", expectedPkg, packages)
				}
			}

			for _, pkg := range packages {
				if strings.HasPrefix(pkg, tt.unexpectedPkg) {
					t.Errorf("Unexpected package %q found in filtered results: %v", pkg, packages)
				}
			}

			t.Logf("Discovered packages: %v", packages)
		})
	}
}
