package main

import (
	"testing"

	"github.com/flanksource/gavel/verify"
)

func TestBuildCommitOptionsTreeAliasEnablesInteractive(t *testing.T) {
	got := buildCommitOptions(CommitOptions{Tree: true}, "/repo", verify.GavelConfig{})
	if !got.Interactive {
		t.Fatalf("Tree alias did not enable interactive mode")
	}
}
