package commit

import (
	"context"
	"testing"
	"time"

	"github.com/flanksource/clicky/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptSelectIndexRoutesToInstalledManager proves the gavel commit chokepoint
// (promptSelectIndex) routes a scoped prompt to an installed interactive sink — the
// path that lets commit questions bubble to the dashboard — and maps the resolved
// answer back to the selected index. It uses the real promptSelectFunc so the whole
// gavel -> clicky -> prompt.Manager chain is exercised.
func TestPromptSelectIndexRoutesToInstalledManager(t *testing.T) {
	m := prompt.NewManager(prompt.NewMemory(prompt.MemoryConfig{}))
	prompt.SetDefault(m)
	t.Cleanup(func() { prompt.SetDefault(nil) })

	ctx := prompt.WithScope(context.Background(), prompt.Scope{
		Owner:  "todo-x",
		Labels: map[string]string{"session": "s1"},
	})
	got := make(chan int, 1)
	go func() {
		idx, _ := promptSelectIndex(ctx, "Pick", []string{"a", "b", "c"})
		got <- idx
	}()

	var id string
	for i := 0; i < 500; i++ {
		if ps := m.List(prompt.Filter{Owner: "todo-x"}); len(ps) == 1 {
			id = ps[0].ID
			break
		}
		time.Sleep(time.Millisecond)
	}
	require.NotEmpty(t, id, "scoped commit prompt never reached the manager")
	require.NoError(t, m.Resolve(id, prompt.Answer{Values: map[string]any{"choice": "2"}}))
	require.Equal(t, 2, <-got)
}

func TestPromptSelectIndexReturnsSelectedIndex(t *testing.T) {
	previous := promptSelectFunc
	promptSelectFunc = func(_ context.Context, options []promptSelectOption, title string) (promptSelectOption, bool) {
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

	index, ok := promptSelectIndex(context.Background(), "Pick one", []string{"alpha", "beta", "gamma"})
	require.True(t, ok)
	assert.Equal(t, 1, index)
}

func TestPromptSelectIndexTreatsCancelledPromptAsNoSelection(t *testing.T) {
	previous := promptSelectFunc
	promptSelectFunc = func(_ context.Context, options []promptSelectOption, title string) (promptSelectOption, bool) {
		return promptSelectOption{}, false
	}
	defer func() {
		promptSelectFunc = previous
	}()

	index, ok := promptSelectIndex(context.Background(), "Pick one", []string{"alpha", "beta"})
	assert.False(t, ok)
	assert.Equal(t, -1, index)
}

func TestPromptSelectIndexRejectsOutOfRangeSelection(t *testing.T) {
	previous := promptSelectFunc
	promptSelectFunc = func(_ context.Context, options []promptSelectOption, title string) (promptSelectOption, bool) {
		return promptSelectOption{Index: len(options)}, true
	}
	defer func() {
		promptSelectFunc = previous
	}()

	index, ok := promptSelectIndex(context.Background(), "Pick one", []string{"alpha", "beta"})
	assert.False(t, ok)
	assert.Equal(t, -1, index)
}
