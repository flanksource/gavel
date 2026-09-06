package lifecycle

import (
	"context"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// Class is the behaviour class a step runs as: what the commit and verify
// invariants key on. A verify step grades without an agent turn; a plan or
// triage envelope is a read-only pass; everything else implements.
func Class(step Step) types.RunMode {
	return classOf(step)
}

// StepDefaults is a step's spec folded from configuration alone — the prompt's
// frontmatter, the project's `todos.<step>` block and the host — with no todo
// and no request. A run dialog displays these defaults while keeping its
// request limited to explicit operator overrides.
//
// The step's own declaration is left out: its placeholders read the todo the
// step runs against, and there is none here. A definition that pins a model on
// a step is therefore not reflected in the defaults, only in the run.
func (h *Host) StepDefaults(ctx context.Context, step Step) (verify.PromptSpec, error) {
	definition, err := h.promptFor(step)
	if err != nil {
		return verify.PromptSpec{}, &ConfigurationError{Err: err}
	}
	var prompt PromptLayerResult
	if definition.Class != types.ModeVerify {
		if prompt, err = PromptLayers(h.WorkDir, nil, definition); err != nil {
			return verify.PromptSpec{}, err
		}
	}
	layers, err := h.profileLayers(ctx, LayerInput{
		RuntimeProfile: h.profileSelection(step.Name, "", prompt.RuntimeProfile),
		Config:         h.Config, Step: step.Name, Frontmatter: prompt.Layers, Host: h.Kind,
	})
	if err != nil {
		return verify.PromptSpec{}, fmt.Errorf("resolve defaults for step %s: %w", step.Name, err)
	}
	composed, err := api.ComposeSpecLayers(layers.Layers...)
	if err != nil {
		return verify.PromptSpec{}, fmt.Errorf("compose defaults for step %s: %w", step.Name, err)
	}
	defaults := verify.PromptSpec{Spec: composed.Spec}
	if layers.Profile != nil {
		defaults.RuntimeProfile = layers.Profile.Profile.ID
	}
	return defaults, nil
}
