package main

import (
	"strings"
	"testing"
)

func TestPRStatusFailFastRequiresFollow(t *testing.T) {
	t.Run("without --follow it is inert, so it is an error", func(t *testing.T) {
		err := PRStatusOptions{FailFast: true}.validate()
		if err == nil {
			t.Fatal("expected an error")
		}
		if !strings.Contains(err.Error(), "--follow") {
			t.Errorf("error must name the missing flag, got %q", err)
		}
	})

	t.Run("with --follow it is accepted", func(t *testing.T) {
		if err := (PRStatusOptions{FailFast: true, Follow: true}).validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("--worktree without --ai-fix has nothing to isolate", func(t *testing.T) {
		err := PRStatusOptions{Worktree: true}.validate()
		if err == nil || !strings.Contains(err.Error(), "--ai-fix") {
			t.Fatalf("want an error naming --ai-fix, got %v", err)
		}
		if err := (PRStatusOptions{Worktree: true, AIFix: true}).validate(); err != nil {
			t.Fatalf("--worktree with --ai-fix: unexpected error: %v", err)
		}
	})

	t.Run("neither flag is fine", func(t *testing.T) {
		if err := (PRStatusOptions{}).validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
