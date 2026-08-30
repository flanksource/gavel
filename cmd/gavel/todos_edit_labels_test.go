package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTodoEditLabels(t *testing.T) {
	t.Run("--label replaces the label set", func(t *testing.T) {
		changes, err := buildTodoEdit(todoEditFlags{Labels: &[]string{"bug", "ui"}})
		require.NoError(t, err)
		require.NotNil(t, changes.Content.Labels)
		assert.Equal(t, []string{"bug", "ui"}, *changes.Content.Labels)
	})

	t.Run("--label normalizes and drops blanks", func(t *testing.T) {
		changes, err := buildTodoEdit(todoEditFlags{Labels: &[]string{"  BUG ", "", "   "}})
		require.NoError(t, err)
		require.NotNil(t, changes.Content.Labels)
		assert.Equal(t, []string{"bug"}, *changes.Content.Labels)
	})

	// The whole point of the pointer: an empty non-nil slice must reach storage
	// as "clear them all", not be mistaken for "no label edit requested".
	t.Run("--clear-labels sets a non-nil empty slice", func(t *testing.T) {
		changes, err := buildTodoEdit(todoEditFlags{ClearLabels: true})
		require.NoError(t, err)
		require.NotNil(t, changes.Content.Labels)
		assert.Empty(t, *changes.Content.Labels)
		assert.False(t, changes.Content.IsEmpty(), "clearing labels is a real edit")
	})

	t.Run("--label and --clear-labels together is an error", func(t *testing.T) {
		_, err := buildTodoEdit(todoEditFlags{Labels: &[]string{"bug"}, ClearLabels: true})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--clear-labels")
	})

	t.Run("neither flag leaves labels untouched", func(t *testing.T) {
		_, err := buildTodoEdit(todoEditFlags{Title: ptr("Only a title")})
		require.NoError(t, err)

		changes, err := buildTodoEdit(todoEditFlags{Title: ptr("Only a title")})
		require.NoError(t, err)
		assert.Nil(t, changes.Content.Labels)
	})

	t.Run("the nothing-to-edit error names the label flags", func(t *testing.T) {
		_, err := buildTodoEdit(todoEditFlags{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--label")
		assert.Contains(t, err.Error(), "--clear-labels")
	})
}

func TestTodosLabelFlagsRegistered(t *testing.T) {
	for _, flag := range []string{"label", "clear-labels"} {
		assert.NotNil(t, todosEditCmd.Flags().Lookup(flag), "missing todos edit --%s", flag)
	}
	assert.NotNil(t, todosCreateCmd.Flags().Lookup("label"), "missing todos create --label")

	for _, flag := range []string{"color", "icon", "description", "global"} {
		assert.NotNil(t, todosLabelsSetCmd.Flags().Lookup(flag), "missing todos labels set --%s", flag)
	}
}
