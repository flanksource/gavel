package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/cmd/gavel/choose"
	"github.com/flanksource/gavel/github"
)

// pushTargetKind tags how we arrived at the chosen branch to push to.
type pushTargetKind int

const (
	pushTargetNone        pushTargetKind = iota
	pushTargetBranchMatch                // open PR's head branch equals local branch
	pushTargetAncestorPR                 // open PR's head is ancestor of HEAD — fast-forward safe
	pushTargetNewPR                      // no suitable PR; will create one
)

type pushTarget struct {
	kind pushTargetKind
	pr   *github.PRListItem // populated for BranchMatch / AncestorPR
}

// pushFlags lets tests intercept the network/git side-effects without
// spinning up GitHub and git servers.
type pushDeps struct {
	searchPRs        func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error)
	defaultBranch    func(github.Options) (string, error)
	createPR         func(github.Options, github.CreatePRInput) (*github.CreatePRResult, error)
	isAncestor       func(workDir, ref, head string) bool
	gitPush          func(workDir, refspec string) error
	pickPR           func(header string, prs []github.PRListItem) (*github.PRListItem, error)
	generatePRPrompt func(ctx context.Context, agent clickyai.Agent, in prContentInput) (prContent, error)
}

func defaultPushDeps() pushDeps {
	return pushDeps{
		searchPRs:        github.SearchPRs,
		defaultBranch:    github.DefaultBranch,
		createPR:         github.CreatePR,
		isAncestor:       gitIsAncestor,
		gitPush:          runGitPush,
		pickPR:           choosePR,
		generatePRPrompt: generatePRContent,
	}
}

// pushAfterCommit runs the push/PR flow after a successful commit (or in
// dry-run, simulates it). Called once per `Run` invocation — if CommitAll
// produced several commits, we still push the final HEAD once.
func pushAfterCommit(ctx context.Context, opts Options, result *Result) error {
	if result == nil || len(result.Commits) == 0 {
		return nil
	}

	ghOpts := github.Options{WorkDir: opts.WorkDir}
	branch, err := gitCurrentBranch(opts.WorkDir)
	if err != nil {
		return fmt.Errorf("resolve current branch: %w", err)
	}
	if branch == "" {
		return fmt.Errorf("cannot push from detached HEAD")
	}

	deps := defaultPushDeps()
	target, candidates, err := decidePushTarget(ctx, ghOpts, branch, deps)
	if err != nil {
		return err
	}

	switch target.kind {
	case pushTargetBranchMatch:
		return executeExistingPRPush(opts, deps, target.pr, "branch-match")
	case pushTargetAncestorPR:
		if len(candidates) > 1 {
			picked, perr := deps.pickPR("Multiple open PRs could accept this push — select one:", candidates)
			if perr != nil {
				return fmt.Errorf("pick PR: %w", perr)
			}
			if picked == nil {
				return fmt.Errorf("push cancelled")
			}
			target.pr = picked
		}
		return executeExistingPRPush(opts, deps, target.pr, "ancestor-match")
	case pushTargetNewPR:
		return executeNewPRPush(ctx, opts, ghOpts, deps, branch, result)
	}
	return nil
}

func decidePushTarget(ctx context.Context, ghOpts github.Options, branch string, deps pushDeps) (pushTarget, []github.PRListItem, error) {
	prs, _, err := deps.searchPRs(ghOpts, github.PRSearchOptions{
		Author: "@me",
		State:  "open",
	})
	if err != nil {
		return pushTarget{}, nil, fmt.Errorf("search open PRs: %w", err)
	}

	for i := range prs {
		if prs[i].Source == branch {
			return pushTarget{kind: pushTargetBranchMatch, pr: &prs[i]}, nil, nil
		}
	}

	cands := findAncestorPRs(ghOpts.WorkDir, prs, deps.isAncestor)
	switch len(cands) {
	case 0:
		return pushTarget{kind: pushTargetNewPR}, nil, nil
	case 1:
		return pushTarget{kind: pushTargetAncestorPR, pr: &cands[0]}, cands, nil
	default:
		return pushTarget{kind: pushTargetAncestorPR}, cands, nil
	}
}

func findAncestorPRs(workDir string, prs github.PRSearchResults, isAncestor func(string, string, string) bool) []github.PRListItem {
	var out []github.PRListItem
	for _, pr := range prs {
		if pr.Source == "" {
			continue
		}
		// The PR head must exist locally as origin/<branch>; if it doesn't,
		// skip rather than failing — that PR isn't a clean target.
		ref := "origin/" + pr.Source
		if isAncestor(workDir, ref, "HEAD") {
			out = append(out, pr)
		}
	}
	return out
}

func executeExistingPRPush(opts Options, deps pushDeps, pr *github.PRListItem, reason string) error {
	refspec := "HEAD:" + pr.Source
	if opts.DryRun {
		fmt.Fprintf(dryRunOutput, "would push %s (%s → PR #%d %s)\n", refspec, reason, pr.Number, pr.URL)
		return nil
	}
	if err := deps.gitPush(opts.WorkDir, refspec); err != nil {
		return fmt.Errorf("git push %s: %w", refspec, err)
	}
	logger.Infof("Pushed to PR #%d (%s): %s", pr.Number, pr.Source, pr.URL)
	return nil
}

func executeNewPRPush(ctx context.Context, opts Options, ghOpts github.Options, deps pushDeps, branch string, result *Result) error {
	refspec := "HEAD:" + branch

	agent, err := buildAgent(opts)
	if err != nil {
		return fmt.Errorf("build AI agent for PR content: %w", err)
	}

	prIn := prContentInput{commits: result.Commits}
	content, err := deps.generatePRPrompt(ctx, agent, prIn)
	if err != nil {
		return fmt.Errorf("generate PR title/body: %w", err)
	}

	if opts.DryRun {
		base, _ := deps.defaultBranch(ghOpts)
		fmt.Fprintf(dryRunOutput, "would push %s and open PR against %s\n", refspec, base)
		fmt.Fprintf(dryRunOutput, "title: %s\n", content.Title)
		if content.Body != "" {
			fmt.Fprintf(dryRunOutput, "body:\n%s\n", content.Body)
		}
		return nil
	}

	if err := deps.gitPush(opts.WorkDir, refspec); err != nil {
		return fmt.Errorf("git push %s: %w", refspec, err)
	}

	created, err := deps.createPR(ghOpts, github.CreatePRInput{
		Title: content.Title,
		Body:  content.Body,
		Head:  branch,
	})
	if err != nil {
		return fmt.Errorf("create PR: %w", err)
	}
	logger.Infof("Opened PR #%d against %s: %s", created.Number, created.Base, created.URL)
	fmt.Fprintln(dryRunOutput, created.URL)
	return nil
}

// --- git helpers ---

func gitCurrentBranch(workDir string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch --show-current: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func gitIsAncestor(workDir, ref, head string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", ref, head)
	if workDir != "" {
		cmd.Dir = workDir
	}
	err := cmd.Run()
	if err == nil {
		return true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false
	}
	logger.V(2).Infof("git merge-base --is-ancestor %s %s: %v", ref, head, err)
	return false
}

func runGitPush(workDir, refspec string) error {
	cmd := exec.Command("git", "push", "origin", refspec)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// --- PR picker (bubbletea single-select wrapper) ---

func choosePR(header string, prs []github.PRListItem) (*github.PRListItem, error) {
	items := make([]string, len(prs))
	for i, pr := range prs {
		items[i] = fmt.Sprintf("#%d  %s  (%s → %s)", pr.Number, pr.Title, pr.Source, pr.Target)
	}
	indices, err := choose.Run(items, choose.WithHeader(header), choose.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(indices) == 0 {
		return nil, nil
	}
	return &prs[indices[0]], nil
}
