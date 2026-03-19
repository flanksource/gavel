package main

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
)

func TestPrioritySymbol(t *testing.T) {
	assert.Equal(t, "⚠ HIGH", prioritySymbol(types.PriorityHigh))
	assert.Equal(t, "◉ MED", prioritySymbol(types.PriorityMedium))
	assert.Equal(t, "○ LOW", prioritySymbol(types.PriorityLow))
	assert.Equal(t, "", prioritySymbol(""))
}

func TestStatusSymbol(t *testing.T) {
	assert.Equal(t, "✗ FAIL", statusSymbol(types.StatusFailed))
	assert.Equal(t, "● PEND", statusSymbol(types.StatusPending))
	assert.Equal(t, "→ RUN", statusSymbol(types.StatusInProgress))
	assert.Equal(t, "✓ DONE", statusSymbol(types.StatusCompleted))
	assert.Equal(t, "⊘ SKIP", statusSymbol(types.StatusSkipped))
	assert.Equal(t, "", statusSymbol(""))
}

func TestHumanDuration(t *testing.T) {
	assert.Equal(t, "30s ago", humanDuration(30*time.Second))
	assert.Equal(t, "5m ago", humanDuration(5*time.Minute))
	assert.Equal(t, "3h ago", humanDuration(3*time.Hour))
	assert.Equal(t, "2d ago", humanDuration(48*time.Hour))
}

func TestFormatTODOItem(t *testing.T) {
	todo := &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{
			Title:    "Fix auth",
			Priority: types.PriorityHigh,
			Status:   types.StatusFailed,
			Path:     types.StringOrSlice{"src/auth.go"},
		},
	}

	result := formatTODOItem(todo)
	assert.Contains(t, result, "Fix auth")
	assert.Contains(t, result, "⚠ HIGH")
	assert.Contains(t, result, "✗ FAIL")
	assert.Contains(t, result, "src/auth.go")
}

func TestFormatTODOItem_FallbackToFilename(t *testing.T) {
	todo := &types.TODO{
		FilePath: "/path/to/fix-auth.md",
		TODOFrontmatter: types.TODOFrontmatter{
			Status: types.StatusPending,
		},
	}

	result := formatTODOItem(todo)
	assert.Contains(t, result, "FixAuth")
}

func TestBuildDetailOptions(t *testing.T) {
	lastRun := time.Now().Add(-2 * time.Hour)
	todo := &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{
			Title:    "Fix bug",
			Priority: types.PriorityMedium,
			Status:   types.StatusFailed,
			Attempts: 2,
			LastRun:  &lastRun,
			Language: types.LanguageGo,
			Branch:   "feature/fix",
			PR: &types.PR{
				Number:        42,
				CommentAuthor: "reviewer",
			},
			Path: types.StringOrSlice{"src/handler.go"},
		},
		Implementation: "Fix the handler logic\nby updating the parser",
	}

	detail := buildDetailOptions(todo)
	assert.Contains(t, detail, "Fix bug")
	assert.Contains(t, detail, "medium")
	assert.Contains(t, detail, "failed")
	assert.Contains(t, detail, "#42")
	assert.Contains(t, detail, "reviewer")
	assert.Contains(t, detail, "src/handler.go")
	assert.Contains(t, detail, "Fix the handler logic")
}
