package main

import (
	"testing"

	"github.com/flanksource/gavel/testrunner"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestPrepareRerunOptionsDisablesRecursiveDiscovery(t *testing.T) {
	base := testrunner.RunOptions{
		Lint:      true,
		Recursive: true,
	}

	got := prepareRerunOptions(base, testui.RerunRequest{
		PackagePaths: []string{"./git"},
		Framework:    "ginkgo",
	}, nil)

	if got.Lint {
		t.Fatalf("Lint should be disabled for reruns")
	}
	if got.Recursive {
		t.Fatalf("Recursive discovery should be disabled for explicit rerun packages")
	}
	if len(got.StartingPaths) != 1 || got.StartingPaths[0] != "./git" {
		t.Fatalf("unexpected rerun package paths: %#v", got.StartingPaths)
	}
}
