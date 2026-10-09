package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/prwatch"
)

func runAzurePRStatus(opts PRStatusOptions, target statusTarget, client provider.Client, workDir string) (any, error) {
	if opts.AIFix {
		return nil, fmt.Errorf("--ai-fix is not supported for Azure DevOps")
	}
	if opts.Worktree {
		return nil, fmt.Errorf("--worktree is not supported for Azure DevOps")
	}
	if len(opts.Comments) > 0 {
		return nil, fmt.Errorf("--comments is not supported for Azure DevOps")
	}
	interval, err := time.ParseDuration(opts.Interval)
	if err != nil || interval <= 0 {
		return nil, fmt.Errorf("invalid --interval %q: expected a positive duration", opts.Interval)
	}
	if opts.TailLogs < 0 {
		return nil, fmt.Errorf("--tail-logs must not be negative")
	}
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	number := target.PR
	if number == 0 {
		number, err = azureBranchPR(ctx, client, workDir)
		if err != nil {
			return nil, err
		}
	}
	result, code := prwatch.Run(prwatch.WatchOptions{Context: ctx, PRNumber: number, Interval: interval, Follow: opts.Follow, FailFast: opts.FailFast, Logs: opts.Logs, TailLogs: opts.TailLogs, Actions: opts.Actions,
		FetchSnapshot: func(ctx context.Context, watch prwatch.WatchOptions) (*prwatch.PRWatchResult, error) {
			snapshot, err := client.Status(ctx, watch.PRNumber, model.StatusOptions{Logs: watch.Logs, TailLogs: watch.TailLogs})
			if err != nil {
				return nil, err
			}
			return &prwatch.PRWatchResult{PR: snapshot.PR, Runs: snapshot.Runs}, nil
		},
	})
	exitCode = code
	if result == nil {
		return nil, nil
	}
	return result, nil
}

func azureBranchPR(ctx context.Context, client provider.Client, workDir string) (int, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("resolve current branch: %w", err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return 0, fmt.Errorf("azure PR status needs a PR number when HEAD is detached")
	}
	prs, err := client.OpenPRs(ctx)
	if err != nil {
		return 0, err
	}
	var numbers []int
	for _, pr := range prs {
		if pr.HeadRefName == branch {
			numbers = append(numbers, pr.Number)
		}
	}
	if len(numbers) != 1 {
		return 0, fmt.Errorf("expected one open Azure PR for branch %q, found %d; specify a PR number", branch, len(numbers))
	}
	return numbers[0], nil
}
