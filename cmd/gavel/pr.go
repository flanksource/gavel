package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/prwatch"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Pull request commands — the gavel replacement for gh pr / gh run",
	Long: `Inspect and drive GitHub pull requests and their Actions checks.

Prefer these over raw gh for PR/CI work — gavel renders workflow steps,
conclusions, timing, and failing-step logs in one view, and can feed failures
into the AI. Reserve raw gh for actions gavel does not cover.

Subcommands:
  status   Show a PR's Actions status (replaces gh pr view / gh run view / gh run list)
  create   Cherry-pick a commit into a fresh worktree and open a PR (AI title/body/branch)
  list     List PRs, optionally with CI status or a live browser dashboard (--ui)

Examples:
  gavel pr status                 # current branch's PR + checks
  gavel pr status 123 --logs      # PR #123 with failing-job logs
  gavel pr status --follow        # block until checks complete
  gavel pr create <SHA>           # open a PR from one commit
  gavel pr list --ui              # live PR dashboard`,
}

type PRStatusOptions struct {
	Repo     string          `flag:"repo" short:"R" help:"GitHub repository (owner/repo)"`
	Follow   bool            `flag:"follow" help:"Keep watching until all checks complete"`
	Interval string          `flag:"interval" help:"Poll interval (e.g. 30s, 1m)" default:"30s"`
	Logs     bool            `flag:"logs" help:"Fetch and include failed job logs (uses extra GitHub API quota)"`
	TailLogs int             `flag:"tail-logs" help:"Number of failed log lines to show per step (only applies with --logs)" default:"100"`
	Comments []string        `flag:"comments" help:"Filter review comments by MatchItem patterns over comment ID and @author/@bot tokens, e.g. '1,2,!3,*,!@coderabbit'"`
	Actions  []string        `flag:"actions" help:"Filter workflow actions by MatchItem patterns over run ID, workflow ID, workflow YAML path, workflow name, and job/check name"`
	Args     []string        `args:"true"`
	Context  context.Context `json:"-"`

	AIFix         bool `flag:"ai-fix" help:"Fix failing checks/comments with AI: each turn is committed and pushed, then the PR status is re-polled until it is green"`
	AIFixMaxIters int  `flag:"ai-fix-max-iterations" help:"Max fix→push→re-poll rounds; 0 uses the pr.fix prompt's workflow.verify.maxIterations"`

	// Embedded: contributes --model, --backend, --api-key, --no-cache,
	// --budget, --debug, --max-tokens, --temperature, --permission-mode,
	// --edit, --allowed-tools, --disallowed-tools, --mcp, --hooks,
	// --skills, --skill-dir, --user, --project, --memory, --bare.
	// Defaults overlay from ~/.captain.yaml via captain configure.
	captaincli.AIRuntimeOptions
}

func (o PRStatusOptions) Help() api.Textable {
	return clicky.Text(`Show a PR's GitHub Actions status — checks, conclusions, timing, and failing-step logs.

Use this instead of gh pr view / gh run view / gh run list: one readable view of
every workflow step for the PR. With no argument it resolves the current branch's
PR (falling back to your most recent PR). Accepts a PR number, owner/repo + number,
or a full PR URL.

When the PR published gavel results, each failing shard shows the failing tests and
lint violations, plus a "Reproduce locally" block with the exact gavel commands that
re-run them here: gavel test --pr <n> / gavel lint --pr <n>.

Key flags:
  --follow          Poll until all checks finish (--interval sets the cadence, default 30s)
  --logs            Also fetch failing-job logs (--tail-logs lines per step; extra API quota)
  --comments LIST   Filter comments by MatchItem patterns over IDs and @author/@bot tokens
  --actions LIST    Filter actions by MatchItem patterns over run/workflow IDs, YAML path, workflow name, or job/check name
  --ai-fix          Fix failures/comments with AI, committing and pushing each turn, then
                    re-polling this same status until it is green (--ai-fix-max-iterations
                    caps the rounds). Configure it under pr.fix in .gavel.yaml.

Examples:
  gavel pr status                              # current branch's PR
  gavel pr status 123                          # PR #123 in this repo
  gavel pr status owner/repo 123               # PR #123 in another repo
  gavel pr status https://github.com/o/r/pull/1
  gavel pr status --follow                     # block until checks complete
  gavel pr status 123 --logs                   # include failing-job logs
  gavel pr status --comments '1,2,!3,*,!@coderabbit'
  gavel pr status --actions '.github/workflows/ci.yml,!deploy'
  gavel pr status --actions 'lint,Install Tests - windows-amd64'   # by job/check name
  gavel pr status --ai-fix                     # fix, push, and re-poll until checks pass`)
}

func runPRStatus(opts PRStatusOptions) (any, error) {
	repo, prNumber, err := parseStatusArgs(opts.Args)
	if err != nil {
		return nil, err
	}

	var ghOpts github.Options
	switch {
	case repo != "":
		ghOpts.Repo = repo
	case opts.Repo != "":
		ghOpts.Repo = opts.Repo
	default:
		workDir, err := getWorkingDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		ghOpts.WorkDir = workDir
	}

	if prNumber == 0 {
		prNumber = resolveOrFallbackPR(ghOpts)
	}

	interval, err := time.ParseDuration(opts.Interval)
	if err != nil {
		return nil, fmt.Errorf("invalid --interval %q: %w", opts.Interval, err)
	}
	watchOpts := prwatch.WatchOptions{
		Options:  ghOpts,
		PRNumber: prNumber,
		Interval: interval,
		Follow:   opts.Follow,
		Logs:     opts.Logs,
		TailLogs: opts.TailLogs,
		Comments: opts.Comments,
		Actions:  opts.Actions,
	}

	result, code := prwatch.Run(watchOpts)
	exitCode = code
	if result == nil {
		return nil, nil
	}
	if opts.Logs && !result.HasFailedRun() {
		logger.Infof("--logs had no effect: no failed jobs found (logs are only shown for failed jobs)")
	}
	if opts.AIFix {
		// Print the status block first so the user sees the input the AI is
		// about to act on before live ai-fix events start streaming to
		// stderr. Returning nil suppresses clicky's trailing print of the
		// same data — the captain history rendered after the run is the
		// meaningful tail in this mode.
		clicky.MustPrint(result, clicky.FormatOptions{})

		ctx := opts.Context
		if ctx == nil {
			ctx = commonsContext.NewContext(context.Background())
		}
		if aiErr := runPRStatusAIFix(ctx, opts, result); aiErr != nil {
			return nil, fmt.Errorf("ai-fix: %w", aiErr)
		}
		// The exit code captured above describes the pre-fix PR, which the fix
		// has just invalidated. --ai-fix's definition of done is its own
		// workflow.verify.commands gate, and reaching here means that gate
		// passed; a failure would have returned an error above.
		exitCode = 0
		return nil, nil
	}

	return result, nil
}

func parseStatusArgs(args []string) (repo string, prNumber int, err error) {
	switch len(args) {
	case 0:
		return "", 0, nil
	case 1:
		return parsePRRef(args[0])
	case 2:
		prNumber, err = strconv.Atoi(args[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid PR number %q: %w", args[1], err)
		}
		repo, err = resolveRepoArg(args[0])
		if err != nil {
			return "", 0, err
		}
		return repo, prNumber, nil
	}
	return "", 0, fmt.Errorf("too many arguments")
}

func resolveOrFallbackPR(ghOpts github.Options) int {
	_, err := github.FetchPR(ghOpts, 0)
	if err == nil {
		return 0
	}

	logger.Infof("No PR found for current branch, checking most recent PR...")
	// Search open PRs first. Searching every state picks the most recent PR
	// regardless of age, so a repo with no open work resolves to something
	// merged months ago and reports its long-stale checks as if they were live.
	for _, state := range []string{"open", "all"} {
		results, _, searchErr := github.SearchPRs(ghOpts, github.PRSearchOptions{
			Author: "@me",
			State:  state,
			Limit:  1,
		})
		if searchErr != nil || len(results) == 0 {
			continue
		}
		logger.Infof("Using most recent %s PR #%d: %s", state, results[0].Number, results[0].Title)
		return results[0].Number
	}

	return 0
}

func init() {
	rootCmd.AddCommand(prCmd)
	clicky.AddNamedCommand("status", prCmd, PRStatusOptions{}, runPRStatus)
}
