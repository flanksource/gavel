package types

import "fmt"

func MergedPR(todos []*TODO) (*PR, error) {
	var merged *PR
	for _, todo := range todos {
		if todo == nil || todo.PR == nil {
			continue
		}
		if err := todo.PR.Validate(); err != nil {
			return nil, err
		}
		if merged != nil && (merged.Number != todo.PR.Number || merged.URL != todo.PR.URL) {
			return nil, fmt.Errorf("cannot merge TODOs linked to different PRs: %s and %s", merged.URL, todo.PR.URL)
		}
		if merged == nil {
			copy := *todo.PR
			merged = &copy
		}
	}
	return merged, nil
}
