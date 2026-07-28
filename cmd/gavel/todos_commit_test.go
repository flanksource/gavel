package main

import (
	"context"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/drivers"
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
	oldCommit, oldMode, oldDefaults := commitAfter, todosRunMode, todosDef
	t.Cleanup(func() {
		commitAfter, todosRunMode, todosDef = oldCommit, oldMode, oldDefaults
	})
	todosRunMode = types.ModeRun

	commitAfter = true
	cfg, err := newAgentRunConfig(context.Background(), drivers.Cli, t.TempDir(), nil, nil)
	if err != nil {
		t.Fatalf("newAgentRunConfig: %v", err)
	}
	if cfg.Workflow == nil || len(cfg.Workflow.Commits) != 1 {
		t.Fatalf("workflow commits = %+v, want one run policy", cfg.Workflow)
	}
	commit := cfg.Workflow.Commits[0]
	if commit.On != api.CommitOnRun || commit.Gates != api.CommitGatesFull {
		t.Fatalf("commit policy = %+v, want on=run gates=full", commit)
	}

	commitAfter = false
	cfg, err = newAgentRunConfig(context.Background(), drivers.Cli, t.TempDir(), nil, nil)
	if err != nil {
		t.Fatalf("newAgentRunConfig without commit: %v", err)
	}
	if cfg.Workflow != nil && len(cfg.Workflow.Commits) > 0 {
		t.Fatalf("workflow commits = %+v, want none", cfg.Workflow.Commits)
	}
}
