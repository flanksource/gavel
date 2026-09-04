package lifecycle

import (
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
)

// Class is the behaviour class a step runs as: what the commit and verify
// invariants key on. A verify step grades without an agent turn; a plan or
// triage envelope is a read-only pass; everything else implements.
func Class(step Step) types.RunMode {
	return classOf(step)
}

// StepDefaults is a step's spec folded from configuration alone — the prompt's
// frontmatter, the project's `todos.<step>` block and the host — with no todo
// and no request. It is what a run dialog seeds itself from: the runtime a step
// would resolve to before the operator has chosen anything, so a dialog
// defaulting to the account-wide mode cannot outrank the frontmatter it was
// supposed to defer to.
//
// The step's own declaration is left out: its placeholders read the todo the
// step runs against, and there is none here. A definition that pins a model on
// a step is therefore not reflected in the defaults, only in the run.
func (h *Host) StepDefaults(step Step) (api.Spec, error) {
	definition, err := h.promptFor(step)
	if err != nil {
		return api.Spec{}, err
	}
	var frontmatter []api.SpecLayer
	if definition.Class != types.ModeVerify {
		if frontmatter, _, err = PromptLayers(h.WorkDir, nil, definition); err != nil {
			return api.Spec{}, err
		}
	}
	resolved, err := ResolveLayers(LayerInput{
		Config: h.Config, Step: step.Name, Frontmatter: frontmatter, Host: h.Kind,
	})
	if err != nil {
		return api.Spec{}, fmt.Errorf("resolve defaults for step %s: %w", step.Name, err)
	}
	return resolved.Spec, nil
}
