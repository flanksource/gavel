package main

import (
	"testing"

	"github.com/flanksource/gavel/verify"
)

func TestBuildCommitOptionsTreeAliasEnablesInteractive(t *testing.T) {
	got := buildCommitOptions(CommitOptions{Tree: true}, "/repo", verify.GavelConfig{}, nil)
	if !got.Interactive {
		t.Fatalf("Tree alias did not enable interactive mode")
	}
}

func TestBuildCommitOptionsPassesInteractiveSummary(t *testing.T) {
	got := buildCommitOptions(CommitOptions{Interactive: true, Summary: true}, "/repo", verify.GavelConfig{}, nil)
	if !got.Interactive || !got.Summary {
		t.Fatalf("interactive summary flags were not propagated: %+v", got)
	}
}
