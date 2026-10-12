package ops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildEditLabels(t *testing.T) {
	t.Run("--label replaces the label set", func(t *testing.T) {
		changes, err := BuildEdit(EditFlags{Labels: &[]string{"bug", "ui"}})
		require.NoError(t, err)
		require.NotNil(t, changes.Content.Labels)
		assert.Equal(t, []string{"bug", "ui"}, *changes.Content.Labels)
	})

	t.Run("--label normalizes and drops blanks", func(t *testing.T) {
		changes, err := BuildEdit(EditFlags{Labels: &[]string{"  BUG ", "", "   "}})
		require.NoError(t, err)
		require.NotNil(t, changes.Content.Labels)
		assert.Equal(t, []string{"bug"}, *changes.Content.Labels)
	})

	// The whole point of the pointer: an empty non-nil slice must reach storage
	// as "clear them all", not be mistaken for "no label edit requested".
	t.Run("--clear-labels sets a non-nil empty slice", func(t *testing.T) {
		changes, err := BuildEdit(EditFlags{ClearLabels: true})
		require.NoError(t, err)
		require.NotNil(t, changes.Content.Labels)
		assert.Empty(t, *changes.Content.Labels)
		assert.False(t, changes.Content.IsEmpty(), "clearing labels is a real edit")
	})

	t.Run("--label and --clear-labels together is an error", func(t *testing.T) {
		_, err := BuildEdit(EditFlags{Labels: &[]string{"bug"}, ClearLabels: true})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--clear-labels")
	})

	t.Run("neither flag leaves labels untouched", func(t *testing.T) {
		changes, err := BuildEdit(EditFlags{Title: ptr("Only a title")})
		require.NoError(t, err)
		assert.Nil(t, changes.Content.Labels)
	})

	t.Run("the nothing-to-edit error names the label flags", func(t *testing.T) {
		_, err := BuildEdit(EditFlags{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--label")
		assert.Contains(t, err.Error(), "--clear-labels")
	})
}
