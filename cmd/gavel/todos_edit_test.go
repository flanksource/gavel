package main

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func ptr[T any](value T) *T { return &value }

func TestBuildTodoEdit(t *testing.T) {
	t.Run("splits content and state flags", func(t *testing.T) {
		changes, err := buildTodoEdit(todoEditFlags{
			Title:        ptr("Fix parser panic"),
			Body:         ptr("body"),
			Plan:         ptr("plan"),
			Verification: ptr("verification"),
			Status:       "completed",
			Priority:     "high",
		})
		if err != nil {
			t.Fatalf("buildTodoEdit: %v", err)
		}
		if changes.Content.Title == nil || *changes.Content.Title != "Fix parser panic" {
			t.Fatalf("edit.Title = %v, want the title flag", changes.Content.Title)
		}
		if changes.Content.Body == nil || *changes.Content.Body != "body" {
			t.Fatalf("edit.Body = %v, want the body flag", changes.Content.Body)
		}
		if changes.Content.Verification == nil || *changes.Content.Verification != "verification" {
			t.Fatalf("edit.Verification = %v, want the verification flag", changes.Content.Verification)
		}
		if changes.Plan == nil || *changes.Plan != "plan" {
			t.Fatalf("plan = %v, want the plan flag", changes.Plan)
		}
		if changes.State.Status == nil || *changes.State.Status != types.StatusCompleted {
			t.Fatalf("state.Status = %v, want completed", changes.State.Status)
		}
		if changes.State.Priority == nil || *changes.State.Priority != types.PriorityHigh {
			t.Fatalf("state.Priority = %v, want high", changes.State.Priority)
		}
	})

	t.Run("status alone is a valid edit", func(t *testing.T) {
		changes, err := buildTodoEdit(todoEditFlags{Status: "draft"})
		if err != nil {
			t.Fatalf("buildTodoEdit: %v", err)
		}
		if !changes.Content.IsEmpty() {
			t.Fatal("edit should be empty when only --status is set")
		}
		if changes.State.Status == nil || *changes.State.Status != types.StatusDraft {
			t.Fatalf("state.Status = %v, want draft", changes.State.Status)
		}
	})

	t.Run("rejects a projected status", func(t *testing.T) {
		_, err := buildTodoEdit(todoEditFlags{Status: "failed"})
		if err == nil {
			t.Fatal("buildTodoEdit(--status failed) = nil error, want rejection")
		}
		if !strings.Contains(err.Error(), "projected") {
			t.Fatalf("error = %q, want it to explain the status is projected", err)
		}
	})

	t.Run("rejects an unknown priority", func(t *testing.T) {
		if _, err := buildTodoEdit(todoEditFlags{Priority: "critical"}); err == nil {
			t.Fatal("buildTodoEdit(--priority critical) = nil error, want rejection")
		}
	})

	t.Run("rejects an empty title", func(t *testing.T) {
		if _, err := buildTodoEdit(todoEditFlags{Title: ptr("   ")}); err == nil {
			t.Fatal("buildTodoEdit(--title '   ') = nil error, want rejection")
		}
	})

	t.Run("rejects a no-op edit", func(t *testing.T) {
		_, err := buildTodoEdit(todoEditFlags{})
		if err == nil {
			t.Fatal("buildTodoEdit() = nil error, want rejection")
		}
		if !strings.Contains(err.Error(), "--plan") || !strings.Contains(err.Error(), "--verification") || !strings.Contains(err.Error(), "--status") {
			t.Fatalf("error = %q, want it to list plan, verification, and status among the editable flags", err)
		}
	})

	t.Run("rejects an empty plan", func(t *testing.T) {
		if _, err := buildTodoEdit(todoEditFlags{Plan: ptr("   ")}); err == nil {
			t.Fatal("buildTodoEdit(--plan '   ') = nil error, want rejection")
		}
	})
}
