package main

import (
	"strings"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
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

func TestBuildPRStatusSystemPrompt_MentionsPRNumberTitleAndBranch(t *testing.T) {
	result := &prwatch.PRWatchResult{
		PR: &github.PRInfo{
			Number:      42,
			Title:       "feat: thing",
			HeadRefName: "feat/thing",
		},
	}
	out := buildPRStatusSystemPrompt(result)
	for _, want := range []string{"#42", "\"feat: thing\"", "feat/thing", "edit files in place"} {
		if !strings.Contains(out, want) {
			t.Errorf("system prompt missing %q; out=%q", want, out)
		}
	}
}

func TestBuildPRStatusSystemPrompt_OmitsBranchWhenAbsent(t *testing.T) {
	result := &prwatch.PRWatchResult{PR: &github.PRInfo{Number: 7, Title: "x"}}
	out := buildPRStatusSystemPrompt(result)
	if strings.Contains(out, "HEAD branch:") {
		t.Errorf("system prompt should omit branch when empty: %q", out)
	}
}

func TestBuildPRStatusPrompt_EmbedsStatusTextAndURL(t *testing.T) {
	result := &prwatch.PRWatchResult{PR: &github.PRInfo{
		Number: 1,
		URL:    "https://github.com/o/r/pull/1",
	}}
	status := "Workflows:\n  ✗ Lint\n    ✗ Go Mod Tidy Check"
	out := buildPRStatusPrompt(result, status)
	if !strings.Contains(out, status) {
		t.Errorf("prompt missing rendered status; out=%q", out)
	}
	if !strings.Contains(out, "https://github.com/o/r/pull/1") {
		t.Errorf("prompt missing PR URL; out=%q", out)
	}
	if !strings.Contains(out, "```") {
		t.Errorf("prompt should fence the status in a code block; out=%q", out)
	}
}

func TestBuildPRStatusPrompt_AppendsNewlineWhenMissing(t *testing.T) {
	result := &prwatch.PRWatchResult{PR: &github.PRInfo{Number: 1}}
	out := buildPRStatusPrompt(result, "no trailing newline")
	if !strings.Contains(out, "no trailing newline\n```") {
		t.Errorf("status block should always end with a newline before the closing fence; out=%q", out)
	}
}
