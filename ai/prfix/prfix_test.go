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
	spec, err := ResolveSpec(api.Spec{Model: api.Model{Name: "agent:sonnet"}}, override, "/repo", pr)
	if err != nil {
		t.Fatalf("ResolveSpec err: %v", err)
	}
	return spec
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
