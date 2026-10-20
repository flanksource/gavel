package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/ai/prfix"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/prwatch"
)

func TestLoopSessionID_LastNonEmptyWins(t *testing.T) {
	tests := []struct {
		name string
		res  *captainai.LoopResult
		want string
	}{
		{"nil result", nil, ""},
		{"no iterations", &captainai.LoopResult{}, ""},
		{
			name: "single iteration reports its session",
			res: &captainai.LoopResult{Iterations: []*captainai.LoopIteration{
				{SessionID: "24eec7df-41e5-4e8f-a6f4-f5cee4299fae"},
			}},
			want: "24eec7df-41e5-4e8f-a6f4-f5cee4299fae",
		},
		{
			name: "iteration without a session id does not clear the known one",
			res: &captainai.LoopResult{Iterations: []*captainai.LoopIteration{
				{SessionID: "s1"},
				{},
			}},
			want: "s1",
		},
		{
			name: "a later session id supersedes an earlier one",
			res: &captainai.LoopResult{Iterations: []*captainai.LoopIteration{
				{SessionID: "s1"},
				{SessionID: "s2"},
			}},
			want: "s2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := loopSessionID(tc.res); got != tc.want {
				t.Errorf("loopSessionID = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHistoryOptionsForRun_PrefersTheRunsOwnSession(t *testing.T) {
	runStart := time.Date(2026, 7, 24, 17, 46, 42, 0, time.UTC)

	withSession := historyOptionsForRun(runStart, "24eec7df-41e5-4e8f-a6f4-f5cee4299fae")
	if withSession.SessionID != "24eec7df-41e5-4e8f-a6f4-f5cee4299fae" {
		t.Errorf("SessionID = %q, want the run's own session", withSession.SessionID)
	}
	if withSession.Last {
		t.Error("Last must be off once the session id pins the run: --last re-resolves by recency and can select another agent's session")
	}

	fallback := historyOptionsForRun(runStart, "")
	if !fallback.Last {
		t.Error("Last must stay on when the backend reports no session id")
	}
	if fallback.SessionID != "" {
		t.Errorf("SessionID = %q, want empty when the run reported none", fallback.SessionID)
	}

	for _, opts := range []struct {
		name string
		got  time.Time
	}{{"with session", withSession.Since}, {"fallback", fallback.Since}} {
		if want := runStart.Add(-2 * time.Second); !opts.got.Equal(want) {
			t.Errorf("%s: Since = %v, want %v (run start minus clock skew)", opts.name, opts.got, want)
		}
	}
}

func TestPRContextOf_CountsOnlyUnresolvedReviewThreads(t *testing.T) {
	result := &prwatch.PRWatchResult{
		PR: &model.PRInfo{
			Number:      42,
			Title:       "feat: thing",
			URL:         "https://github.com/o/r/pull/42",
			HeadRefName: "feat/thing",
		},
		Comments: []model.PRComment{
			{IsReviewThread: true, IsResolved: false, IsOutdated: false},
			{IsReviewThread: true, IsResolved: true, IsOutdated: false},
			{IsReviewThread: true, IsResolved: false, IsOutdated: true},
			{IsReviewThread: true, IsResolved: false, IsOutdated: false},
			// An issue comment or review body has IsResolved false structurally
			// — nothing can ever flip it — so counting it would tell the agent
			// there is work outstanding that it can never discharge.
			{IsReviewThread: false, IsResolved: false, IsOutdated: false},
		},
	}

	got := prContextOf(result, "Workflows:\n  ✗ Lint\n")
	if got.UnresolvedComments != 2 {
		t.Errorf("UnresolvedComments = %d, want 2 (resolved and outdated comments need no reply)", got.UnresolvedComments)
	}
	if got.Number != 42 || got.Branch != "feat/thing" || got.URL != result.PR.URL || got.Title != result.PR.Title {
		t.Errorf("PR identity not projected: %+v", got)
	}
	if !strings.Contains(got.StatusText, "✗ Lint") {
		t.Errorf("StatusText = %q, want the rendered snapshot", got.StatusText)
	}
}

func conflictingResult() *prwatch.PRWatchResult {
	return &prwatch.PRWatchResult{
		PR: &model.PRInfo{
			Number: 42, State: "OPEN", Mergeable: "CONFLICTING",
			URL:         "https://github.com/o/r/pull/42",
			BaseRefName: "main", HeadRefName: "feat/thing",
			HeadRefOID:        "0123456789abcdef0123456789abcdef01234567",
			HeadRepositoryURL: "https://github.com/o/r",
		},
		Conflicts: &github.MergeConflictReport{
			BaseRefName: "main", HeadRefName: "feat/thing",
			BaseOID: "fedcba9876543210fedcba9876543210fedcba98",
			Files:   []github.MergeConflict{{Path: "go.mod", Kind: "content"}},
		},
	}
}

func TestPRContextOf_ProjectsTheMergeConflicts(t *testing.T) {
	got := prContextOf(conflictingResult(), "status")
	if got.BaseBranch != "main" || got.BaseSHA != "fedcba9876543210fedcba9876543210fedcba98" {
		t.Errorf("BaseBranch, BaseSHA = %q, %q; want main and the base commit the rebase replays onto", got.BaseBranch, got.BaseSHA)
	}
	if len(got.Conflicts) != 1 || got.Conflicts[0] != (prfix.Conflict{Path: "go.mod", Kind: "content"}) {
		t.Errorf("Conflicts = %+v, want go.mod (content)", got.Conflicts)
	}

	// A CONFLICTING verdict git could not reproduce (stale, or no local checkout)
	// has no files to resolve, so there is no rebase to start and nothing to list.
	stale := conflictingResult()
	stale.Conflicts.Files, stale.Conflicts.Unavailable = nil, "GitHub's CONFLICTING verdict looks stale"
	if got := prContextOf(stale, "status"); len(got.Conflicts) != 0 {
		t.Errorf("Conflicts = %+v, want none for an unreproducible conflict", got.Conflicts)
	}
}

// --worktree is a spec layer, so it composes like any other setup: the worktree
// is branched from the PR head (not the caller's HEAD, which may be another
// branch entirely) and lives outside the repository.
func TestPRFixLayers_WorktreeBranchesFromThePRHeadOutsideTheRepo(t *testing.T) {
	repoRoot := t.TempDir()
	result := conflictingResult()
	setup := prFixWorktreeSetup(prFixWorktree{RepoRoot: repoRoot, PR: result.PR, CacheDir: t.TempDir()})

	layers, err := prFixLayers(prFixLayerOptions{
		Resolve: prfix.ResolveOptions{Base: api.Spec{Model: api.Model{Name: "agent:sonnet"}}, Dir: repoRoot, PR: prContextOf(result, "status")},
		Setup:   setup,
	})
	if err != nil {
		t.Fatal(err)
	}
	composed, err := api.ComposeSpecLayers(api.ResolveSpecOptions{Layers: layers})
	if err != nil {
		t.Fatal(err)
	}

	got := composed.Spec.Setup
	if got == nil || got.Checkout == nil || got.Checkout.Worktree == nil {
		t.Fatalf("composed spec has no worktree setup: %+v", got)
	}
	wt := got.Checkout.Worktree
	if got.Checkout.Path != repoRoot {
		t.Errorf("Checkout.Path = %q, want the repo root %q", got.Checkout.Path, repoRoot)
	}
	if wt.Mode != shell.WorktreeNew || wt.Base != result.PR.HeadRefOID {
		t.Errorf("worktree = %+v, want a new worktree based on the PR head %s", wt, result.PR.HeadRefOID)
	}
	if strings.HasPrefix(wt.Path, repoRoot) {
		t.Errorf("worktree path %q is inside the repo %q", wt.Path, repoRoot)
	}
	if !strings.Contains(wt.Prefix, "pr-42") {
		t.Errorf("worktree branch prefix %q does not name the PR", wt.Prefix)
	}
}

func hookNames(hooks []any) []string {
	names := make([]string, 0, len(hooks))
	for _, h := range hooks {
		if named, ok := h.(interface{ Name() string }); ok {
			names = append(names, named.Name())
		}
	}
	return names
}

// Order is load-bearing: setup relocates the run before the rebase starts (the
// commit hooks have no PreRun), so the rebase lands in the worktree; the rebase
// hook's post-turn push follows the commit pipeline's, and both precede the
// verifiers judging the turn.
func TestPRFixHooks_SetupThenCommitThenRebaseThenVerify(t *testing.T) {
	wf := &api.Workflow{
		Verify:  &api.Verify{Commands: []string{"true"}},
		Commits: []api.Commit{{On: api.CommitOnTurn, Mode: api.CommitModeCommit}},
	}
	hooks, err := prFixHooks(context.Background(), prFixRun{Workflow: wf, Tee: io.Discard, RepoRoot: t.TempDir(), Result: conflictingResult()})
	if err != nil {
		t.Fatal(err)
	}
	names := hookNames(hooks)
	if len(names) != 4 || names[0] != "setup" || names[1] != "commit:turn" || names[2] != prFixRebaseHookName {
		t.Fatalf("hook order = %q, want setup, commit:turn, %s, then the verifier", names, prFixRebaseHookName)
	}
	if _, ok := hooks[1].(agent.PreRun); ok {
		t.Errorf("%T has a PreRun that would run before the rebase starts", hooks[1])
	}
	if !hooks[2].(*prFixRebaseHook).target.Rebase {
		t.Error("a conflicting PR's push target must allow the rebase to rewrite the branch")
	}
	if _, ok := hooks[3].(*capverify.Plugin); !ok {
		t.Errorf("last hook = %T, want the workflow's verifier", hooks[3])
	}

	clean := conflictingResult()
	clean.Conflicts = nil
	hooks, err = prFixHooks(context.Background(), prFixRun{Workflow: wf, Tee: io.Discard, RepoRoot: t.TempDir(), Result: clean})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range hookNames(hooks) {
		if name == prFixRebaseHookName {
			t.Errorf("a PR with no conflicts must not start a rebase; hooks = %q", hookNames(hooks))
		}
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// conflictingCheckout is a repo on branch feat/thing whose go.mod conflicts
// with main's, plus the report prwatch would have produced for it.
func conflictingCheckout(t *testing.T) (string, *github.MergeConflictReport) {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "gavel@example.com")
	runGit(t, dir, "config", "user.name", "Gavel Test")
	write := func(content string) {
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("require xz v0.5.12\n")
	runGit(t, dir, "add", "go.mod")
	runGit(t, dir, "commit", "-m", "base")
	runGit(t, dir, "switch", "-c", "feat/thing")
	write("require xz v0.5.14\n")
	runGit(t, dir, "commit", "-am", "branch")
	runGit(t, dir, "switch", "main")
	write("require xz v0.5.20\n")
	runGit(t, dir, "commit", "-am", "main")
	runGit(t, dir, "switch", "feat/thing")
	return dir, &github.MergeConflictReport{
		BaseRefName: "main", HeadRefName: "feat/thing",
		BaseOID: runGit(t, dir, "rev-parse", "main"), HeadOID: runGit(t, dir, "rev-parse", "feat/thing"),
		Files: []github.MergeConflict{{Path: "go.mod", Kind: "content"}},
	}
}

// The rebase is reported on the run's own stream — in order with the agent's
// events and kept with the run's notices — not as a detached log line.
func TestPRFixRebaseHook_ReportsTheRebaseItStartsOnTheRunStream(t *testing.T) {
	dir, report := conflictingCheckout(t)
	hc := &agent.HookContext{Context: context.Background(), Response: &captainai.Response{Workspace: &api.Workspace{Cwd: dir}}}

	if err := (&prFixRebaseHook{report: report}).PreRun(hc); err != nil {
		t.Fatalf("PreRun: %v", err)
	}

	notices := hc.Workspace().Notices
	want := "[pre-run] rebasing feat/thing onto origin/main: stopped with 1 conflicted — go.mod"
	if len(notices) != 1 || notices[0].Text != want {
		t.Fatalf("notices = %+v, want exactly %q", notices, want)
	}
}

// publishedCheckout publishes the conflicting checkout's PR branch to a bare
// remote and returns the push target the run would hold for it.
func publishedCheckout(t *testing.T) (dir, remote string, report *github.MergeConflictReport, target *provider.RunTarget) {
	t.Helper()
	dir, report = conflictingCheckout(t)
	remote = filepath.Join(t.TempDir(), "source.git")
	runGit(t, dir, "init", "--bare", remote)
	runGit(t, dir, "push", remote, "feat/thing")
	return dir, remote, report, &provider.RunTarget{RepositoryURL: remote, Branch: "feat/thing", HeadSHA: report.HeadOID, Rebase: true}
}

// A rebase that finished in a turn leaves a clean tree, so the commit pipeline
// has nothing to push it with; the rebase hook's post-turn push publishes it,
// and pushes nothing while the rebase is still stopped.
func TestPRFixRebaseHook_PushesTheRebasedBranchOnceTheRebaseFinishes(t *testing.T) {
	dir, remote, report, target := publishedCheckout(t)
	hook := &prFixRebaseHook{report: report, target: target}
	hc := &agent.HookContext{Context: context.Background(), Response: &captainai.Response{Workspace: &api.Workspace{Cwd: dir}}}
	if err := hook.PreRun(hc); err != nil {
		t.Fatalf("PreRun: %v", err)
	}

	if err := hook.Post(hc, agent.PhaseTurn); err != nil {
		t.Fatalf("Post while stopped: %v", err)
	}
	if got := runGit(t, remote, "rev-parse", "feat/thing"); got != report.HeadOID {
		t.Fatalf("remote moved to %s while the rebase was still stopped", got)
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("require xz v0.5.20\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "go.mod")
	runGit(t, dir, "-c", "core.editor=true", "rebase", "--continue")
	rebased := runGit(t, dir, "rev-parse", "HEAD")

	if err := hook.Post(hc, agent.PhaseTurn); err != nil {
		t.Fatalf("Post after the rebase: %v", err)
	}
	if got := runGit(t, remote, "rev-parse", "feat/thing"); got != rebased {
		t.Errorf("remote feat/thing = %s, want the rebased head %s", got, rebased)
	}
	if target.HeadSHA != rebased {
		t.Errorf("target.HeadSHA = %s, want the lease moved to %s", target.HeadSHA, rebased)
	}
}

// The push is pinned to the head the run started from (checked, then leased): a push someone made to
// the PR branch meanwhile fails the turn instead of being overwritten.
func TestPRFixRebaseHook_RefusesToOverwriteAConcurrentPush(t *testing.T) {
	dir, remote, report, target := publishedCheckout(t)
	hook := &prFixRebaseHook{report: report, target: target}
	hc := &agent.HookContext{Context: context.Background(), Response: &captainai.Response{Workspace: &api.Workspace{Cwd: dir}}}
	other := filepath.Join(t.TempDir(), "other")
	runGit(t, dir, "clone", "--branch", "feat/thing", remote, other)
	runGit(t, other, "-c", "user.email=other@example.com", "-c", "user.name=Other", "commit", "--allow-empty", "-m", "someone else's push")
	runGit(t, other, "push", "origin", "feat/thing")
	theirs := runGit(t, other, "rev-parse", "HEAD")

	if err := hook.PreRun(hc); err != nil {
		t.Fatalf("PreRun: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("require xz v0.5.20\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "go.mod")
	runGit(t, dir, "-c", "core.editor=true", "rebase", "--continue")

	err := hook.Post(hc, agent.PhaseTurn)

	if err == nil || !strings.Contains(err.Error(), "advanced") {
		t.Fatalf("Post err = %v, want the advanced-branch refusal", err)
	}
	if got := runGit(t, remote, "rev-parse", "feat/thing"); got != theirs {
		t.Errorf("remote feat/thing = %s, want their push %s kept", got, theirs)
	}
}

// The original defect was a wiring one: captain's CmdVerifier swallowed a
// ten-minute check's output because nothing handed it a live sink. This asserts
// the sink actually reaches the verifier, which is where that would have shown.
func TestPRFixHooks_TeeVerifyOutput(t *testing.T) {
	var sink bytes.Buffer

	wf := &api.Workflow{Verify: &api.Verify{Commands: []string{"echo wired"}}}
	hooks, err := prFixHooks(context.Background(), prFixRun{Workflow: wf, Tee: &sink, Result: conflictingResult()})
	if err != nil {
		t.Fatalf("prFixHooks: %v", err)
	}

	var teed int
	for _, hook := range hooks {
		plugin, ok := hook.(*capverify.Plugin)
		if !ok {
			continue
		}
		cmd, ok := plugin.Verifier().(*capverify.CmdVerifier)
		if !ok {
			continue
		}
		if cmd.Output != io.Writer(&sink) {
			t.Errorf("verify hook %q did not get the tee", plugin.Name())
		}
		teed++
	}
	if teed != 1 {
		t.Fatalf("expected exactly one command verifier to be teed, got %d", teed)
	}
}

func TestApplyMaxIterations_FlagOverridesThePromptCap(t *testing.T) {
	spec := api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{MaxIterations: 3}}}
	if err := applyMaxIterations(&spec, 5); err != nil {
		t.Fatalf("applyMaxIterations err: %v", err)
	}
	if spec.Workflow.Verify.MaxIterations != 5 {
		t.Errorf("MaxIterations = %d, want the flag's 5", spec.Workflow.Verify.MaxIterations)
	}
}

func TestApplyMaxIterations_ZeroKeepsThePromptCap(t *testing.T) {
	spec := api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{MaxIterations: 3}}}
	if err := applyMaxIterations(&spec, 0); err != nil {
		t.Fatalf("applyMaxIterations err: %v", err)
	}
	if spec.Workflow.Verify.MaxIterations != 3 {
		t.Errorf("MaxIterations = %d, want the prompt's 3 left untouched", spec.Workflow.Verify.MaxIterations)
	}
}

func TestApplyMaxIterations_ErrorsWhenThereIsNoVerifyToCap(t *testing.T) {
	spec := api.Spec{}
	if err := applyMaxIterations(&spec, 5); err == nil {
		t.Fatal("expected an error: a spec without workflow.verify cannot honour the flag")
	}
}
