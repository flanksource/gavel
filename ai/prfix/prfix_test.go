package prfix

import (
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/verify"
)

func prContext() PRContext {
	return PRContext{
		Number:             42,
		Title:              "feat: thing",
		URL:                "https://github.com/o/r/pull/42",
		Branch:             "feat/thing",
		StatusText:         "Workflows:\n  ✗ Lint\n    ✗ Go Mod Tidy Check",
		UnresolvedComments: 3,
	}
}

func resolve(t *testing.T, override verify.PromptSpec, pr PRContext) api.Spec {
	t.Helper()
	layers, err := Layers(ResolveOptions{Base: api.Spec{Model: api.Model{Name: "agent:sonnet"}}, Prompt: override, Dir: "/repo", PR: pr})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := api.ResolveSpecLayers(api.ResolveSpecOptions{Layers: layers, RequireModel: true})
	if err != nil {
		t.Fatalf("ResolveSpec err: %v", err)
	}
	return resolved.Spec
}

func TestResolveSpecRendersPRIdentityIntoSystemPrompt(t *testing.T) {
	out := resolve(t, verify.PromptSpec{}, prContext()).Prompt.System
	for _, want := range []string{"/repo", "feat/thing", "#42", "(feat: thing)", "edit files in place"} {
		if !strings.Contains(out, want) {
			t.Errorf("system prompt missing %q; out=%q", want, out)
		}
	}
}

func TestResolveSpecOmitsTitleClauseWhenAbsent(t *testing.T) {
	pr := prContext()
	pr.Title = ""
	out := resolve(t, verify.PromptSpec{}, pr).Prompt.System
	if strings.Contains(out, "#42 (") {
		t.Errorf("system prompt should omit the title clause when there is no title: %q", out)
	}
}

func TestResolveSpecEmbedsStatusTextURLAndCommentCount(t *testing.T) {
	pr := prContext()
	out := resolve(t, verify.PromptSpec{}, pr).Prompt.User
	for _, want := range []string{pr.StatusText, pr.URL, "3 unresolved review comments"} {
		if !strings.Contains(out, want) {
			t.Errorf("user prompt missing %q; out=%q", want, out)
		}
	}
}

func TestResolveSpecOmitsCommentSentenceWhenNoneUnresolved(t *testing.T) {
	pr := prContext()
	pr.UnresolvedComments = 0
	out := resolve(t, verify.PromptSpec{}, pr).Prompt.User
	if strings.Contains(out, "unresolved review comments") {
		t.Errorf("user prompt should omit the comment sentence when the count is 0: %q", out)
	}
}

// The verify command is the loop's only definition of done, and --follow is what
// makes it one: a still-running check rollup exits 0, so polling without it would
// green-light the run before CI re-ran the freshly pushed commit.
func TestResolveSpecDeclaresAFollowingVerifyCommand(t *testing.T) {
	workflow := resolve(t, verify.PromptSpec{}, prContext()).Workflow
	if workflow == nil || workflow.Verify == nil {
		t.Fatalf("prompt declares no workflow.verify: %+v", workflow)
	}
	if len(workflow.Verify.Commands) == 0 {
		t.Fatal("workflow.verify.commands is empty; --ai-fix would have no definition of done")
	}
	cmd := workflow.Verify.Commands[0]
	for _, want := range []string{"gavel pr status", "--follow"} {
		if !strings.Contains(cmd, want) {
			t.Errorf("verify command %q missing %q", cmd, want)
		}
	}
	if strings.Contains(cmd, "--ai-fix") {
		t.Errorf("verify command must not re-enter --ai-fix: %q", cmd)
	}
	if workflow.Verify.MaxIterations <= 0 {
		t.Errorf("MaxIterations = %d, want a positive cap", workflow.Verify.MaxIterations)
	}
}

// Every turn's commit is pushed, so it must be a real commit: captain collapses a
// fixup chain only at the end of a run, and `fixup!` subjects must never reach a
// branch someone else pulls.
func TestResolveSpecCommitsRealCommitsPerTurn(t *testing.T) {
	workflow := resolve(t, verify.PromptSpec{}, prContext()).Workflow
	if workflow == nil || len(workflow.Commits) != 1 {
		t.Fatalf("want exactly one commit policy, got %+v", workflow)
	}
	commit := workflow.Commits[0]
	if commit.On != api.CommitOnTurn {
		t.Errorf("On = %q, want turn", commit.On)
	}
	if commit.Mode != api.CommitModeCommit {
		t.Errorf("Mode = %q, want commit (a pushed fixup chain never gets squashed)", commit.Mode)
	}
}

func TestResolveSpecOverrideReplacesVerifyCommands(t *testing.T) {
	override := verify.PromptSpec{Spec: api.Spec{Workflow: &api.Workflow{
		Verify: &api.Verify{Commands: []string{"gavel pr status --follow"}, MaxIterations: 7},
	}}}
	workflow := resolve(t, override, prContext()).Workflow
	if got := workflow.Verify.Commands; len(got) != 1 || got[0] != "gavel pr status --follow" {
		t.Errorf("Commands = %v, want the override's single command", got)
	}
	if workflow.Verify.MaxIterations != 7 {
		t.Errorf("MaxIterations = %d, want the override's 7", workflow.Verify.MaxIterations)
	}
}

func TestResolveSpecOverrideKeepsDefaultPromptAndBudget(t *testing.T) {
	override := verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{MaxTurns: 5}}}
	spec := resolve(t, override, prContext())
	if spec.Budget.MaxTurns != 5 {
		t.Errorf("MaxTurns = %d, want the override's 5", spec.Budget.MaxTurns)
	}
	if !strings.Contains(spec.Prompt.User, "Workflows:") {
		t.Errorf("an override that names no prompt must keep the default body; got %q", spec.Prompt.User)
	}
	if spec.Workflow == nil || spec.Workflow.Verify == nil || len(spec.Workflow.Verify.Commands) == 0 {
		t.Error("an override that names no workflow must keep the default verify commands")
	}
}

// The verify command runs from wherever the agent works — a --worktree run sits on
// a scratch branch no PR is open for — so it must name the PR rather than resolve
// it from the current branch.
func TestResolveSpecVerifyCommandTargetsThePRURL(t *testing.T) {
	pr := prContext()
	commands := resolve(t, verify.PromptSpec{}, pr).Workflow.Verify.Commands
	last := commands[len(commands)-1]
	if !strings.Contains(last, "gavel pr status "+pr.URL) {
		t.Errorf("verify command %q does not target %s", last, pr.URL)
	}
}

func conflictingPR() PRContext {
	pr := prContext()
	pr.BaseBranch = "main"
	pr.Conflicts = []Conflict{{Path: "go.mod", Kind: "content"}, {Path: "cmd/app/main.go", Kind: "modify/delete"}}
	return pr
}

func TestResolveSpecHighlightsMergeConflictsFirst(t *testing.T) {
	pr := conflictingPR()
	out := resolve(t, verify.PromptSpec{}, pr).Prompt.User
	section := strings.Index(out, "## Merge conflicts")
	if section < 0 {
		t.Fatalf("user prompt has no merge-conflict section: %q", out)
	}
	if status := strings.Index(out, pr.StatusText); status < section {
		t.Errorf("the conflict section must come before the status snapshot; out=%q", out)
	}
	for _, want := range []string{"`go.mod` (content)", "`cmd/app/main.go` (modify/delete)", "git merge origin/main", "`feat/thing`", "git add"} {
		if !strings.Contains(out, want) {
			t.Errorf("conflict section missing %q; out=%q", want, out)
		}
	}
}

func TestResolveSpecOmitsMergeConflictSectionWithoutConflicts(t *testing.T) {
	out := resolve(t, verify.PromptSpec{}, prContext()).Prompt.User
	if strings.Contains(out, "Merge conflicts") {
		t.Errorf("user prompt should have no conflict section for a mergeable PR: %q", out)
	}
}

// The unmerged-path and leftover-marker checks are captain cmd verifiers declared
// ahead of the PR poll, so an unresolved merge fails the turn on a sub-second git
// check with the offending files as feedback, instead of on a CI poll.
func TestResolveSpecVerifiesConflictResolutionBeforePolling(t *testing.T) {
	commands := resolve(t, verify.PromptSpec{}, conflictingPR()).Workflow.Verify.Commands
	want := []string{"git diff --name-only --diff-filter=U --exit-code", "git diff --check HEAD"}
	if len(commands) != 3 || commands[0] != want[0] || commands[1] != want[1] {
		t.Fatalf("Commands = %q, want %q ahead of the pr status poll", commands, want)
	}
	if !strings.HasPrefix(commands[2], "gavel pr status ") {
		t.Errorf("last command = %q, want the pr status poll", commands[2])
	}

	clean := resolve(t, verify.PromptSpec{}, prContext()).Workflow.Verify.Commands
	if len(clean) != 1 {
		t.Errorf("a mergeable PR should only poll pr status; got %q", clean)
	}
}

func TestPromptsRegistersPRFixOnce(t *testing.T) {
	got := Prompts()
	if len(got) != 1 {
		t.Fatalf("Prompts() = %d entries, want 1", len(got))
	}
	if got[0].ID != got[0].ConfigPath {
		t.Errorf("ID %q and ConfigPath %q must match for reflection-based resolution", got[0].ID, got[0].ConfigPath)
	}
	if got[0].Default == "" {
		t.Error("Default prompt body is empty; the go:embed did not take")
	}
}
