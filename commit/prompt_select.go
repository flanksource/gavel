package commit

import (
	"context"

	"github.com/flanksource/clicky"
)

type promptSelectOption struct {
	Index int
	Label string
}

func (o promptSelectOption) String() string {
	return o.Label
}

// promptSelectFunc routes a select prompt to clicky, which renders it on the TTY
// when attached and otherwise hands it to the installed interactive sink (the
// dashboard), inheriting any prompt.Scope on ctx. Tests override it.
var promptSelectFunc = func(ctx context.Context, options []promptSelectOption, title string) (promptSelectOption, bool) {
	return clicky.PromptSelectCtx(ctx, options, clicky.PromptSelectOptions[promptSelectOption]{
		Title: title,
	})
}

func promptSelectIndex(ctx context.Context, title string, items []string) (int, bool) {
	if len(items) == 0 {
		return -1, false
	}

	options := make([]promptSelectOption, len(items))
	for i, item := range items {
		options[i] = promptSelectOption{
			Index: i,
			Label: item,
		}
	}

	selected, ok := promptSelectFunc(ctx, options, title)
	if !ok || selected.Index < 0 || selected.Index >= len(items) {
		return -1, false
	}

	return selected.Index, true
}
