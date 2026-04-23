package commit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptSelectIndexReturnsSelectedIndex(t *testing.T) {
	previous := promptSelectFunc
	promptSelectFunc = func(options []promptSelectOption, title string) (promptSelectOption, bool) {
		require.Equal(t, "Pick one", title)
		require.Len(t, options, 3)
		assert.Equal(t, "alpha", options[0].Label)
		assert.Equal(t, "beta", options[1].Label)
		assert.Equal(t, "gamma", options[2].Label)
		return options[1], true
	}
	defer func() {
		promptSelectFunc = previous
	}()

	index, ok := promptSelectIndex("Pick one", []string{"alpha", "beta", "gamma"})
	require.True(t, ok)
	assert.Equal(t, 1, index)
}

func TestPromptSelectIndexTreatsCancelledPromptAsNoSelection(t *testing.T) {
	previous := promptSelectFunc
	promptSelectFunc = func(options []promptSelectOption, title string) (promptSelectOption, bool) {
		return promptSelectOption{}, false
	}
	defer func() {
		promptSelectFunc = previous
	}()

	index, ok := promptSelectIndex("Pick one", []string{"alpha", "beta"})
	assert.False(t, ok)
	assert.Equal(t, -1, index)
}

func TestPromptSelectIndexRejectsOutOfRangeSelection(t *testing.T) {
	previous := promptSelectFunc
	promptSelectFunc = func(options []promptSelectOption, title string) (promptSelectOption, bool) {
		return promptSelectOption{Index: len(options)}, true
	}
	defer func() {
		promptSelectFunc = previous
	}()

	index, ok := promptSelectIndex("Pick one", []string{"alpha", "beta"})
	assert.False(t, ok)
	assert.Equal(t, -1, index)
}
