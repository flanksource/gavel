package ui

import (
	"testing"

	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
)

// TestAddTodoStatusBuckets pins the status→bucket mapping shared by
// summarizeTodos and countProjectTodos. Every known status must land in exactly
// one detail bucket, and Open must cover everything that is not terminal.
func TestAddTodoStatusBuckets(t *testing.T) {
	tests := []struct {
		status types.Status
		want   todoCounts
	}{
		{types.StatusDraft, todoCounts{Total: 3, Open: 3, Draft: 3}},
		{types.StatusPending, todoCounts{Total: 3, Open: 3, Pending: 3}},
		{types.StatusInProgress, todoCounts{Total: 3, Open: 3, InProgress: 3}},
		{types.StatusReview, todoCounts{Total: 3, Open: 3, Review: 3}},
		{types.StatusAsk, todoCounts{Total: 3, Open: 3, Ask: 3}},
		{types.StatusFailed, todoCounts{Total: 3, Open: 3, Failed: 3}},
		// Unverified has no bucket of its own and falls through to Pending.
		{types.StatusUnverified, todoCounts{Total: 3, Open: 3, Pending: 3}},
		{types.StatusVerified, todoCounts{Total: 3, Open: 3, Verified: 3}},
		{types.StatusSkipped, todoCounts{Total: 3, Open: 3, Skipped: 3}},
		// Completed is the only terminal status: counted, but never Open.
		{types.StatusCompleted, todoCounts{Total: 3, Completed: 3}},
	}
	for _, test := range tests {
		t.Run(string(test.status), func(t *testing.T) {
			var counts todoCounts
			addTodoStatus(&counts, test.status, 3)
			assert.Equal(t, test.want, counts)
		})
	}

	covered := map[types.Status]bool{}
	for _, test := range tests {
		covered[test.status] = true
	}
	for _, status := range types.KnownStatuses() {
		assert.True(t, covered[status], "status %q has no bucket expectation", status)
	}
}

// TestAddTodoStatusMatchesSummarizeTodos keeps the aggregate fold and the
// per-item walk on the same mapping: folding a count map must equal summarizing
// the equivalent materialized list.
func TestAddTodoStatusMatchesSummarizeTodos(t *testing.T) {
	byStatus := map[types.Status]int{
		types.StatusPending:   4,
		types.StatusReview:    2,
		types.StatusCompleted: 7,
		types.StatusFailed:    1,
	}
	var items types.TODOS
	var folded todoCounts
	for status, n := range byStatus {
		addTodoStatus(&folded, status, n)
		for i := 0; i < n; i++ {
			items = append(items, &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: status}})
		}
	}
	assert.Equal(t, summarizeTodos(items), folded)
}
