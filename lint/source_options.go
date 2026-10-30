package lint

import (
	"fmt"

	"github.com/flanksource/gavel/linters"
)

func validateTags(registry *linters.Registry, names, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	for _, name := range names {
		linter, ok := registry.Get(name)
		if !ok {
			return fmt.Errorf("no linter registered for %s", name)
		}
		if _, ok := linter.(linters.TagsMapper); ok {
			return nil
		}
	}
	return fmt.Errorf("--tags is not supported by the selected linters")
}
