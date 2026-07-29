package main

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func ptr[T any](value T) *T { return &value }

func TestBuildTodoEdit(t *testing.T) {
	t.Run("splits content and state flags", func(t *testing.T) {
		edit, state, err := buildTodoEdit(todoEditFlags{
			Title:    ptr("Fix parser panic"),
			Body:     ptr("body"),
			Status:   "completed",
			Priority: "high",
		})
		if err != nil {
			t.Fatalf("buildTodoEdit: %v", err)
		}
		if edit.Title == nil || *edit.Title != "Fix parser panic" {
			t.Fatalf("edit.Title = %v, want the title flag", edit.Title)
		}
		if edit.Body == nil || *edit.Body != "body" {
			t.Fatalf("edit.Body = %v, want the body flag", edit.Body)
		}
		if state.Status == nil || *state.Status != types.StatusCompleted {
			t.Fatalf("state.Status = %v, want completed", state.Status)
		}
		if state.Priority == nil || *state.Priority != types.PriorityHigh {
			t.Fatalf("state.Priority = %v, want high", state.Priority)
		}
	})

	t.Run("status alone is a valid edit", func(t *testing.T) {
		edit, state, err := buildTodoEdit(todoEditFlags{Status: "draft"})
		if err != nil {
			t.Fatalf("buildTodoEdit: %v", err)
		}
		if !edit.IsEmpty() {
			t.Fatal("edit should be empty when only --status is set")
		}
		if state.Status == nil || *state.Status != types.StatusDraft {
			t.Fatalf("state.Status = %v, want draft", state.Status)
		}
	})

	t.Run("rejects a projected status", func(t *testing.T) {
		_, _, err := buildTodoEdit(todoEditFlags{Status: "failed"})
		if err == nil {
			t.Fatal("buildTodoEdit(--status failed) = nil error, want rejection")
		}
		if !strings.Contains(err.Error(), "projected") {
			t.Fatalf("error = %q, want it to explain the status is projected", err)
		}
	})

	t.Run("rejects an unknown priority", func(t *testing.T) {
		if _, _, err := buildTodoEdit(todoEditFlags{Priority: "critical"}); err == nil {
			t.Fatal("buildTodoEdit(--priority critical) = nil error, want rejection")
		}
	})

	t.Run("rejects an empty title", func(t *testing.T) {
		if _, _, err := buildTodoEdit(todoEditFlags{Title: ptr("   ")}); err == nil {
			t.Fatal("buildTodoEdit(--title '   ') = nil error, want rejection")
		}
	})

	t.Run("rejects a no-op edit", func(t *testing.T) {
		_, _, err := buildTodoEdit(todoEditFlags{})
		if err == nil {
			t.Fatal("buildTodoEdit() = nil error, want rejection")
		}
		if !strings.Contains(err.Error(), "--status") {
			t.Fatalf("error = %q, want it to list --status among the editable flags", err)
		}
	})
}
