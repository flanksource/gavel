package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

type CommitOptions struct {
	Stage     string `flag:"stage" help:"Which changes to commit: staged|unstaged|all" default:"staged"`
	CommitAll bool   `flag:"commit-all" short:"A" help:"Split the selected change set into multiple AI-planned commits"`
	Max       int    `flag:"max" help:"Maximum number of commit groups (0 = unlimited)" default:"0"`
	Message   string `flag:"message" short:"m" help:"Explicit commit message (skips LLM)"`
	Model     string `flag:"model" help:"Override LLM model from .gavel.yaml commit.model"`
	DryRun    bool   `flag:"dry-run" help:"Print the generated message without committing"`
	Force     bool   `flag:"force" help:"Skip pre-commit hooks"`
	NoCache   bool   `flag:"no-cache" help:"Bypass the LLM response cache at ~/.cache/clicky-ai.db"`
	Push      bool   `flag:"push" short:"p" help:"After committing, push to a matching open PR or open a new PR"`
	WorkDir   string `flag:"work-dir" help:"Working directory"`
}

func (o CommitOptions) Help() string {
	return `Generate a conventional commit message via LLM and run pre-commit hooks.

Reads pre-commit hooks from .gavel.yaml under commit.hooks. Hooks run with
sh -c in the git root and abort the commit on non-zero exit. Pass --force
to skip hooks.

Examples:
  gavel commit                          # LLM-generated message, staged changes
  gavel commit -A                       # split staged changes into multiple commits
  gavel commit -A --max=5               # split into at most 5 commits
  gavel commit -m "chore: bump dep"     # explicit message, skip LLM
  gavel commit --stage all --dry-run    # stage everything, print message
  gavel commit --force                  # skip hooks`
}

func init() {
	clicky.AddNamedCommand("commit", rootCmd, CommitOptions{}, runCommit)
}

func runCommit(opts CommitOptions) (any, error) {
	workDir := opts.WorkDir
	if workDir == "" {
		wd, err := getWorkingDir()
		if err != nil {
			return nil, err
		}
		workDir = wd
	}
	if root := repomap.FindGitRoot(workDir); root != "" {
		workDir = root
	}

	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		logger.Warnf("Failed to load .gavel.yaml: %v", err)
	}

	result, err := commitpkg.Run(context.Background(), commitpkg.Options{
		WorkDir:   workDir,
		Stage:     opts.Stage,
		CommitAll: opts.CommitAll,
		Max:       opts.Max,
		DryRun:    opts.DryRun,
		Force:     opts.Force,
		NoCache:   opts.NoCache,
		Model:     opts.Model,
		Message:   opts.Message,
		Push:      opts.Push,
		Config:    cfg.Commit,
	})

	if err != nil {
		if errors.Is(err, commitpkg.ErrNothingStaged) {
			fmt.Fprintln(os.Stderr, "nothing staged to commit")
			exitCode = 1
			return nil, nil
		}
		return result, err
	}
	return result, nil
}
