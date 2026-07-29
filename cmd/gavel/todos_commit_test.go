package main

import (
	"context"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
)

func TestTodosRunCommitFlagRegistered(t *testing.T) {
	flag := todosRunCmd.Flags().Lookup("commit")
	if flag == nil {
		t.Fatal("expected todos run --commit flag to be registered")
	}
	if flag.DefValue != "true" {
		t.Fatalf("expected --commit default true, got %q", flag.DefValue)
	}
}

func TestNewAgentRunConfigCommitPolicy(t *testing.T) {
	dir := isolatedTodosRun(t, "")

	commitAfter = true
	cfg, _, err := newAgentRunConfig(context.Background(), dir, nil, nil)
	if err != nil {
		t.Fatalf("newAgentRunConfig: %v", err)
	}
	if cfg.Spec.Workflow == nil || len(cfg.Spec.Workflow.Commits) != 1 {
		t.Fatalf("workflow commits = %+v, want one run policy", cfg.Spec.Workflow)
	}
	commit := cfg.Spec.Workflow.Commits[0]
	if commit.On != api.CommitOnRun || commit.Gates != api.CommitGatesFull {
		t.Fatalf("commit policy = %+v, want on=run gates=full", commit)
	}

	commitAfter = false
	cfg, _, err = newAgentRunConfig(context.Background(), dir, nil, nil)
	if err != nil {
		t.Fatalf("newAgentRunConfig without commit: %v", err)
	}
	if cfg.Spec.Workflow != nil && len(cfg.Spec.Workflow.Commits) > 0 {
		t.Fatalf("workflow commits = %+v, want none", cfg.Spec.Workflow.Commits)
	}
}

// TestPlanModeNeverCommits pins the mode invariant: a plan writes a document for
// review, so a `commits:` block inherited from the ai: base or todos config must
// not turn it into a committing run — regardless of --commit.
func TestPlanModeNeverCommits(t *testing.T) {
	dir := isolatedTodosRun(t, "ai:\n  workflow:\n    commits:\n      - on: run\n        gates: full\n")
	todosRunMode = types.ModePlan
	commitAfter = true

	cfg, _, err := newAgentRunConfig(context.Background(), dir, nil, nil)
	if err != nil {
		t.Fatalf("newAgentRunConfig: %v", err)
	}
	if cfg.Spec.Workflow != nil && len(cfg.Spec.Workflow.Commits) > 0 {
		t.Fatalf("plan workflow commits = %+v, want none", cfg.Spec.Workflow.Commits)
	}
}
