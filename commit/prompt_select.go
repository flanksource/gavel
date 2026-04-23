package commit

import "github.com/flanksource/clicky"

type promptSelectOption struct {
	Index int
	Label string
}

func (o promptSelectOption) String() string {
	return o.Label
}

var promptSelectFunc = func(options []promptSelectOption, title string) (promptSelectOption, bool) {
	return clicky.PromptSelect(options, clicky.PromptSelectOptions[promptSelectOption]{
		Title: title,
	})
}

func promptSelectIndex(title string, items []string) (int, bool) {
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

	selected, ok := promptSelectFunc(options, title)
	if !ok || selected.Index < 0 || selected.Index >= len(items) {
		return -1, false
	}

	return selected.Index, true
}
