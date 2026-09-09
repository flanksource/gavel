package verify

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/merge"
)

// Merge layers one .gavel.yaml config value over another: the override's set
// fields win, its unset fields leave the base intact, and nested structures merge
// field by field so naming one setting never erases its siblings.
//
// It is structural — derived from the struct definitions rather than restated per
// type — so adding a field to any config type needs no merge code, and cannot be
// silently forgotten by a merge function that was not updated. configPolicy names
// the handful of rules structure cannot express.
func Merge[T any](base, override T) T {
	return merge.Apply(base, override, configPolicy())
}

// MergeGavelConfig layers override onto base. Config resolution calls it once per
// .gavel.yaml found, lowest precedence first: built-in defaults, the user's home
// config, the git root's, then the target directory's.
func MergeGavelConfig(base, override GavelConfig) GavelConfig {
	return Merge(base, override)
}

// configPolicy is the merge policy for every .gavel.yaml type. It is the spec
// policy — pointers to scalars mean "explicitly set", so `enabled: false` in a
// repo config turns something off — plus PromptSpec, the one config type whose
// layering is a domain rule the structure cannot express.
//
// The lists that accumulate across layers instead of being replaced say so on the
// field, with a `merge:"append"` tag.
func configPolicy() merge.Policy {
	return api.MergePolicy().With(merge.Policy{Merger: []any{PromptSpec{}}})
}
