package main

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/prwatch"
)

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
