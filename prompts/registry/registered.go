package registry

import (
	"fmt"

	"github.com/flanksource/gavel/ai/aifix"
	"github.com/flanksource/gavel/ai/prfix"
	"github.com/flanksource/gavel/commit"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/gavel/testrunner/outline"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
)

// All returns every overridable prompt in stable command-family order.
func All() []prompts.Prompt {
	var all []prompts.Prompt
	all = append(all, aifix.Prompts()...)
	all = append(all, prfix.Prompts()...)
	all = append(all, gavelgit.Prompts()...)
	all = append(all, commit.Prompts()...)
	all = append(all, todoprompt.Prompts()...)
	all = append(all, status.Prompts()...)
	all = append(all, outline.Prompts()...)
	return all
}

// Resolve expands every registered prompt against a merged config trace.
func Resolve(opts ResolveOptions) ([]ResolvedPrompt, error) {
	all := All()
	resolved := make([]ResolvedPrompt, 0, len(all))
	for _, desc := range all {
		item, err := ResolveOne(opts, desc)
		if err != nil {
			return nil, fmt.Errorf("resolve %s (%s): %w", desc.ID, desc.ConfigPath, err)
		}
		resolved = append(resolved, item)
	}
	return resolved, nil
}
