package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/github"
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
		PR: &github.PRInfo{
			Number:      42,
			Title:       "feat: thing",
			URL:         "https://github.com/o/r/pull/42",
			HeadRefName: "feat/thing",
		},
		Comments: []github.PRComment{
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

// The original defect was a wiring one: captain's CmdVerifier swallowed a
// ten-minute check's output because nothing handed it a live sink. This asserts
// the sink actually reaches the verifier, which is where that would have shown.
func TestPRFixHooks_TeeVerifyOutput(t *testing.T) {
	var sink bytes.Buffer

	wf := &api.Workflow{Verify: &api.Verify{Commands: []string{"echo wired"}}}
	hooks, err := prFixHooks(context.Background(), wf, nil, &sink)
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
