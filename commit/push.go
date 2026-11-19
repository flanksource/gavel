package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
)

// PushMode selects the commits --push publishes and seeds the PR from.
type PushMode string

const (
	// PushModeCommits publishes the commits ahead of the branch's upstream.
	PushModeCommits PushMode = ""
	// PushModeBranch publishes the whole branch: every commit ahead of the
	// default branch, so an already-pushed branch can still get its PR.
	PushModeBranch PushMode = "branch"
)

// pushTargetKind tags how we arrived at the chosen branch to push to.
type pushTargetKind int

const (
	pushTargetNone        pushTargetKind = iota
	pushTargetBranchMatch                // open PR's head branch equals local branch
	pushTargetAncestorPR                 // open PR's head is ancestor of HEAD — fast-forward safe
	pushTargetNewPR                      // create a new PR
)

type pushTarget struct {
	kind pushTargetKind
	pr   *github.PRListItem // populated for BranchMatch / AncestorPR
}

// pushFlags lets tests intercept the network/git side-effects without
// spinning up GitHub and git servers.
type pushDeps struct {
	searchPRs           func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error)
	defaultBranch       func(github.Options) (string, error)
	createPR            func(github.Options, model.CreatePRInput) (*model.CreatePRResult, error)
	isAncestor          func(workDir, ref, head string) bool
	gitPush             func(workDir, refspec string) error
	rebaseOnto          func(workDir, upstreamBranch string) error
	pickPR              func(ctx context.Context, header string, prs []github.PRListItem) (pushTarget, error)
	generatePRPrompt    func(ctx context.Context, in PRContentInput) (PRContent, error)
	aheadCommits        func(workDir, branch, defaultBase string) ([]CommitResult, error)
	confirmProtectedRef func(branch string) bool
	enableAutoMerge     func(github.Options, *model.CreatePRResult, string) error
}

func defaultPushDeps(ctx context.Context, opts Options) pushDeps {
	return pushDeps{
		searchPRs:     github.SearchPRs,
		defaultBranch: github.DefaultBranch,
		createPR:      github.CreatePR,
		isAncestor:    gitIsAncestor,
		gitPush: func(workDir, refspec string) error {
			return runGitPush(remoteRef{WorkDir: workDir, Remote: opts.target.remoteName(), Ref: refspec})
		},
		rebaseOnto: func(workDir, branch string) error {
			return rebaseOnto(ctx, remoteRef{WorkDir: workDir, Remote: opts.target.remoteName(), Ref: branch})
		},
		pickPR:           choosePR,
		generatePRPrompt: GeneratePRContent,
		aheadCommits:     loadAheadCommits,
		confirmProtectedRef: func(branch string) bool {
			return confirmProtectedBranchPush(ctx, branch)
		},
		enableAutoMerge: func(opts github.Options, pr *model.CreatePRResult, method string) error {
			return github.EnableAutoMerge(opts, pr.NodeID, method)
		},
	}
}

// protectedBranches are remote branches gavel will never push to without
// explicit user confirmation. These are the names commonly configured as
// branch-protected on GitHub; pushing directly to them bypasses PR review.
var protectedBranches = map[string]bool{
	"main":    true,
	"master":  true,
	"develop": true,
	"trunk":   true,
}

func isProtectedBranch(name string) bool {
	return protectedBranches[strings.ToLower(strings.TrimSpace(name))]
}

// confirmProtectedBranchPush prompts the user before pushing to a
// protected remote branch. Returns true if the user confirmed.
func confirmProtectedBranchPush(ctx context.Context, branch string) bool {
	header := fmt.Sprintf("Pushing to protected branch %q bypasses PR review. Proceed?", branch)
	idx, ok := promptSelectIndex(ctx, header, []string{
		fmt.Sprintf("No, cancel push to %s", branch),
		fmt.Sprintf("Yes, push directly to %s", branch),
	})
	if !ok {
		return false
	}
	return idx == 1
}

// pushDepsForTest, when non-nil, replaces defaultPushDeps() inside
// pushAfterCommit. Only set from tests that drive Run() end-to-end and
// need to swap GitHub/git side effects.
var pushDepsForTest *pushDeps

// pushAfterCommit runs the push/PR flow after a successful commit (or in
// dry-run, simulates it). Called once per `Run` invocation — if CommitAll
// produced several commits, we still push the final HEAD once.
//
// When result.Commits is empty (i.e. --push was used with nothing staged),
// pushAfterCommit falls back to local commits ahead of upstream so there's
// still something to seed PR title/body generation with. If HEAD has no
// commits ahead of upstream either, returns ErrNothingToPush.
func pushAfterCommit(ctx context.Context, opts Options, result *Result) error {
	if err := opts.resolveTarget(); err != nil {
		return err
	}
	var deps pushDeps
	if pushDepsForTest != nil {
		deps = *pushDepsForTest
	} else {
		var err error
		deps, err = providerPushDeps(ctx, opts)
		if err != nil {
			return err
		}
	}
	return pushWithDeps(ctx, opts, result, deps)
}

func pushWithDeps(ctx context.Context, opts Options, result *Result, deps pushDeps) error {
	if result == nil {
		return nil
	}

	if err := opts.resolveTarget(); err != nil {
		return err
	}
	ghOpts, err := opts.pushGithubOptions()
	if err != nil {
		return err
	}
	branch, err := gitCurrentBranch(opts.WorkDir)
	if err != nil {
		return fmt.Errorf("resolve current branch: %w", err)
	}
	if branch == "" {
		return fmt.Errorf("cannot push from detached HEAD")
	}

	prSeed := result
	if len(result.Commits) == 0 || opts.PushMode == PushModeBranch || opts.Target != "" {
		base, baseErr := opts.pushBase(deps, ghOpts)
		if baseErr != nil {
			return fmt.Errorf("resolve default branch: %w", baseErr)
		}
		baseRef := ""
		if base != "" {
			baseRef = opts.target.remoteName() + "/" + base
		}
		// An empty branch makes the ahead base the default branch rather
		// than the branch's own upstream.
		aheadBranch := branch
		if opts.PushMode == PushModeBranch || opts.Target != "" {
			aheadBranch = ""
		}
		if opts.Target != "" && !validRef(opts.WorkDir, baseRef) {
			return fmt.Errorf("--target base %q is not available locally; fetch %s first", baseRef, opts.target.remoteName())
		}
		ahead, aErr := deps.aheadCommits(opts.WorkDir, aheadBranch, baseRef)
		if aErr != nil {
			return fmt.Errorf("read ahead commits: %w", aErr)
		}
		if opts.DryRun && !result.PushOnly {
			for _, c := range result.Commits {
				if c.Hash == "" {
					ahead = append(ahead, c)
				}
			}
		}
		if len(ahead) == 0 {
			return ErrNothingToPush
		}
		prSeed = &Result{Commits: ahead}
		if len(result.Commits) == 0 {
			result.Commits = ahead
			// In dry-run mode, runSingleCommit/runCommitAll would have printed
			// the commit preview; here we print it ourselves so the user sees
			// what's about to be pushed before the "would push ..." line.
			if opts.DryRun {
				printDryRunPreview(result)
			}
		}
	}

	if opts.PushBranch != "" {
		if err := pushHEADTo(opts, deps, opts.PushBranch); err != nil {
			return err
		}
		logger.Infof("Pushed %s to %s", branch, opts.PushBranch)
		return nil
	}

	if opts.target.head != "" {
		branch = opts.target.head
	}
	target, candidates, err := decidePushTarget(opts, ghOpts, branch, deps)
	if err != nil {
		return err
	}

	if target.kind == pushTargetAncestorPR && len(candidates) > 1 {
		target, err = deps.pickPR(ctx, "Multiple open PRs could accept this push — select one:", candidates)
		if err != nil {
			return fmt.Errorf("pick PR: %w", err)
		}
		if target.kind == pushTargetNone {
			return fmt.Errorf("push cancelled")
		}
	}

	if opts.AutoMerge && target.kind != pushTargetNewPR {
		logger.Warnf("--auto-merge only applies to newly opened PRs; pushing to an existing PR, auto-merge not changed")
	}

	switch target.kind {
	case pushTargetBranchMatch:
		return executeExistingPRPush(opts, deps, target.pr, "branch-match")
	case pushTargetAncestorPR:
		return executeExistingPRPush(opts, deps, target.pr, "ancestor-match")
	case pushTargetNewPR:
		return executeNewPRPush(ctx, opts, ghOpts, deps, branch, prSeed)
	}
	return nil
}

func decidePushTarget(opts Options, ghOpts github.Options, branch string, deps pushDeps) (pushTarget, []github.PRListItem, error) {
	prs, _, err := deps.searchPRs(ghOpts, github.PRSearchOptions{
		Author: "@me",
		State:  "open",
	})
	if err != nil {
		return pushTarget{}, nil, fmt.Errorf("search open PRs: %w", err)
	}
	if opts.target.base != "" {
		matching := make(github.PRSearchResults, 0, len(prs))
		for _, pr := range prs {
			if pr.Target == opts.target.base {
				matching = append(matching, pr)
			}
		}
		prs = matching
	}

	for i := range prs {
		if prs[i].Source == branch {
			return pushTarget{kind: pushTargetBranchMatch, pr: &prs[i]}, nil, nil
		}
	}

	if opts.target.head != "" {
		return pushTarget{kind: pushTargetNewPR}, nil, nil
	}
	opts.WorkDir = ghOpts.WorkDir
	cands := findAncestorPRs(opts, prs, deps.isAncestor)
	switch len(cands) {
	case 0:
		return pushTarget{kind: pushTargetNewPR}, nil, nil
	case 1:
		return pushTarget{kind: pushTargetAncestorPR, pr: &cands[0]}, cands, nil
	default:
		return pushTarget{kind: pushTargetAncestorPR}, cands, nil
	}
}

func findAncestorPRs(opts Options, prs github.PRSearchResults, isAncestor func(string, string, string) bool) []github.PRListItem {
	var out []github.PRListItem
	for _, pr := range prs {
		if pr.Source == "" {
			continue
		}
		// The PR head must exist locally on the selected remote; if it doesn't,
		// skip rather than failing — that PR isn't a clean target.
		ref := opts.target.remoteName() + "/" + pr.Source
		if isAncestor(opts.WorkDir, ref, "HEAD") {
			out = append(out, pr)
		}
	}
	return out
}

func executeExistingPRPush(opts Options, deps pushDeps, pr *github.PRListItem, reason string) error {
	if opts.DryRun {
		fmt.Fprintf(dryRunOutput, "would push HEAD:%s (%s → PR #%d %s; remote %s)\n", pr.Source, reason, pr.Number, pr.URL, opts.target.remoteName())
		return nil
	}
	if err := pushHEADTo(opts, deps, pr.Source); err != nil {
		return err
	}
	logger.Infof("Pushed to PR #%d (%s): %s", pr.Number, pr.Source, pr.URL)
	printExistingPRSummary(pr)
	return nil
}

// pushHEADTo rebases HEAD onto origin/<branch> and pushes it there, confirming
// first when branch is protected.
func pushHEADTo(opts Options, deps pushDeps, branch string) error {
	refspec := "HEAD:" + branch
	if opts.DryRun {
		fmt.Fprintf(dryRunOutput, "would push %s (remote %s)\n", refspec, opts.target.remoteName())
		return nil
	}
	if isProtectedBranch(branch) && !deps.confirmProtectedRef(branch) {
		return fmt.Errorf("push to protected branch %q cancelled", branch)
	}
	if err := deps.rebaseOnto(opts.WorkDir, branch); err != nil {
		return err
	}
	if err := deps.gitPush(opts.WorkDir, refspec); err != nil {
		return fmt.Errorf("git push %s: %w", refspec, err)
	}
	return nil
}

func executeNewPRPush(ctx context.Context, opts Options, ghOpts github.Options, deps pushDeps, branch string, result *Result) (err error) {
	base, baseErr := opts.pushBase(deps, ghOpts)
	if baseErr != nil {
		return fmt.Errorf("resolve default branch: %w", baseErr)
	}

	prIn := PRContentInput{
		Commits: commitInputsFromResults(result.Commits), Options: opts,
	}
	content, err := deps.generatePRPrompt(ctx, prIn)
	if err != nil {
		return fmt.Errorf("generate PR title/body: %w", err)
	}

	// When the user is on a protected branch (e.g. main/master), pushing
	// HEAD:main bypasses review entirely. Use the AI-suggested branch
	// name as the head ref instead so the new PR is opened from a fresh
	// topic branch.
	headBranch := branch
	if opts.target.head != "" {
		headBranch = opts.target.head
	} else if isProtectedBranch(branch) || branch == base {
		if content.Branch == "" {
			return fmt.Errorf("on protected branch %q, but AI did not suggest a branch name", branch)
		}
		headBranch = content.Branch
	}
	if headBranch == base {
		return fmt.Errorf("PR topic branch %q must differ from base", headBranch)
	}
	refspec := "HEAD:" + headBranch

	if opts.DryRun {
		fmt.Fprintf(dryRunOutput, "would push %s and open PR against %s (remote %s)\n", refspec, base, opts.target.remoteName())
		fmt.Fprintf(dryRunOutput, "title: %s\n", content.Title)
		if content.Body != "" {
			fmt.Fprintf(dryRunOutput, "body:\n%s\n", content.Body)
		}
		if opts.AutoMerge {
			fmt.Fprintf(dryRunOutput, "would enable auto-merge (%s) on the new PR\n", opts.MergeType)
		}
		return nil
	}

	// Guard against pushing directly to a protected remote branch. This
	// only fires when the LLM-suggested branch (or the user's current
	// branch on the non-protected path) collides with main/master/etc.
	if isProtectedBranch(headBranch) {
		if !deps.confirmProtectedRef(headBranch) {
			return fmt.Errorf("push to protected branch %q cancelled", headBranch)
		}
	}

	// Skip the rebase when we're pushing to a brand-new branch derived
	// from HEAD: there's nothing remote to reconcile against.
	if headBranch == branch || opts.Target != "" {
		if err := deps.rebaseOnto(opts.WorkDir, base); err != nil {
			return err
		}
	}

	if err := deps.gitPush(opts.WorkDir, refspec); err != nil {
		return fmt.Errorf("git push %s: %w", refspec, err)
	}

	created, err := deps.createPR(ghOpts, model.CreatePRInput{
		Title: content.Title,
		Body:  content.Body,
		Head:  headBranch,
		Base:  base,
	})
	if err != nil {
		return fmt.Errorf("create PR: %w", err)
	}
	logger.Infof("Opened PR #%d against %s: %s", created.Number, created.Base, created.URL)

	if opts.AutoMerge {
		if err := deps.enableAutoMerge(ghOpts, created, opts.MergeType); err != nil {
			return fmt.Errorf("enable auto-merge on PR #%d: %w", created.Number, err)
		}
		logger.Infof("Enabled auto-merge (%s) on PR #%d", opts.MergeType, created.Number)
	}

	printNewPRSummary(created, content)
	return nil
}

// printNewPRSummary renders the just-opened PR's title, URL, and body on
// stdout. Replaces the trailing commit re-print: the user already saw
// "Committed <hash> ..." per commit, what they actually want at the end
// is the PR they just opened.
func printNewPRSummary(created *model.CreatePRResult, content PRContent) {
	if created == nil {
		return
	}
	t := clicky.Text(fmt.Sprintf("PR #%d", created.Number), "font-bold text-green-600").
		Space().Append(content.Title, "font-bold").NewLine().
		Append(created.URL, "text-muted").NewLine()
	if body := strings.TrimSpace(content.Body); body != "" {
		t = t.NewLine().Append(body, "")
	}
	fmt.Println(t.ANSI())
}

// printExistingPRSummary renders the target PR's title and URL after a
// successful push to an existing PR. PRListItem has no body field, so
// only title + URL are shown.
func printExistingPRSummary(pr *github.PRListItem) {
	if pr == nil {
		return
	}
	t := clicky.Text(fmt.Sprintf("PR #%d", pr.Number), "font-bold text-green-600").
		Space().Append(pr.Title, "font-bold").NewLine().
		Append(pr.URL, "text-muted").NewLine()
	fmt.Println(t.ANSI())
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

func runGitPush(opts remoteRef) error {
	cmd := exec.Command("git", "push", opts.Remote, opts.Ref)
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// --- PR picker ---

func choosePR(ctx context.Context, header string, prs []github.PRListItem) (pushTarget, error) {
	items := make([]string, len(prs)+1)
	for i, pr := range prs {
		items[i] = fmt.Sprintf("#%d  %s  (%s → %s)", pr.Number, pr.Title, pr.Source, pr.Target)
	}
	items[len(prs)] = "Create new PR"
	index, ok := promptSelectIndex(ctx, header, items)
	if !ok {
		return pushTarget{kind: pushTargetNone}, nil
	}
	if index == len(prs) {
		return pushTarget{kind: pushTargetNewPR}, nil
	}
	return pushTarget{kind: pushTargetAncestorPR, pr: &prs[index]}, nil
}
