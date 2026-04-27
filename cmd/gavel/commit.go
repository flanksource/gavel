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
	CommitAll bool   `flag:"commit-all" short:"A" help:"Split the selected change set into commits grouped by directory"`
	MaxFiles  int    `flag:"max-files" help:"Max files per commit group before splitting further by subdirectory" default:"7"`
	MaxLines  int    `flag:"max-lines" help:"Max changed lines (adds+dels, excluding new files) per commit group before splitting further by subdirectory" default:"500"`
	Message   string `flag:"message" short:"m" help:"Explicit commit message (skips only the message-generation LLM call)"`
	Model     string `flag:"model" help:"Override LLM model from .gavel.yaml commit.model"`
	DryRun    bool   `flag:"dry-run" help:"Print the generated message without committing"`
	Force     bool   `flag:"force" help:"Skip pre-commit hooks"`
	NoCache   bool   `flag:"no-cache" help:"Bypass the LLM response cache at ~/.cache/clicky-ai.db"`
	Push      bool   `flag:"push" short:"p" help:"After committing, push to a matching open PR or open a new PR"`
	Precommit string `flag:"precommit" help:"Behavior for pre-commit gitignore and linked dependency checks: prompt|fail|skip|false"`
	Compat    string `flag:"compat" help:"Behavior for AI compatibility analysis and findings (default: skip): prompt|fail|skip|false"`
	WorkDir   string `flag:"work-dir" help:"Working directory"`
}

func (o CommitOptions) Help() string {
	return `Generate a conventional commit message via LLM and run pre-commit hooks.

Reads pre-commit hooks from .gavel.yaml under commit.hooks. Hooks run with
sh -c in the git root and abort the commit on non-zero exit. Pass --force
to skip hooks.

Before hooks run, staged files are checked against commit.gitignore patterns
(typically set in ~/.gavel.yaml). Matches trigger a per-file prompt to
(1) append the matched pattern to the repo .gitignore, (2) append the file's
folder, (3) append the exact file, (4) allow it via commit.allow in the repo's
.gavel.yaml, or (5) cancel. --precommit=fail|skip|false overrides the prompt;
non-TTY runs auto-escalate prompt -> fail.

Staged go.mod / go.work / package.json files are also scanned for local
references that escape the git root (go.mod replace, go.work use,
package.json file:/link:/portal: or ../ paths). Newly introduced or changed
violations relative to HEAD prompt to (1) unstage the manifest so the bad edit
is dropped from the commit, (2) ignore and keep it in this commit, or
(3) cancel. --precommit controls this check too, and skip|false disables both
the gitignore and linked-deps checks entirely.

AI analysis can also check for removed functionality and compatibility issues.
By default this is skipped. Use --compat=prompt|fail to enable it; skip|false
disables the compatibility AI checks entirely, and non-TTY runs auto-escalate
prompt -> fail.

The -A flag groups staged files by their top-level directory and recursively
splits any group that exceeds --max-files or --max-lines. An LLM still writes
the conventional commit message for each group.

Examples:
  gavel commit                          # LLM-generated message, staged changes
  gavel commit -A                       # one commit per directory, split when large
  gavel commit -A --max-files=3         # tighter file cap; triggers deeper splits
  gavel commit -A --max-lines=50        # tighter line cap; triggers deeper splits
  gavel commit -m "chore: bump dep"     # explicit message, still run compatibility analysis
  gavel commit --stage all --dry-run    # stage everything, print message
  gavel commit --force                  # skip hooks
  gavel commit --precommit=fail         # error on gitignore or linked-deps issues`
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
		WorkDir:       workDir,
		Stage:         opts.Stage,
		CommitAll:     opts.CommitAll,
		MaxFiles:      opts.MaxFiles,
		MaxLines:      opts.MaxLines,
		DryRun:        opts.DryRun,
		Force:         opts.Force,
		NoCache:       opts.NoCache,
		Model:         opts.Model,
		Message:       opts.Message,
		Push:          opts.Push,
		PrecommitMode: opts.Precommit,
		CompatMode:    opts.Compat,
		Config:        cfg.Commit,
	})

	if err != nil {
		if errors.Is(err, commitpkg.ErrNothingStaged) {
			fmt.Fprintln(os.Stderr, "nothing staged to commit")
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrGitIgnoreCancelled) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrLinkedDepsCancelled) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrCompatibilityCancelled) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		return result, err
	}
	return result, nil
}
