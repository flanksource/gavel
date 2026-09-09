package lifecycle

import (
	"context"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
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
// and no request. A run dialog displays these defaults while keeping its
// request limited to explicit operator overrides.
//
// The step's own declaration is left out: its placeholders read the todo the
// step runs against, and there is none here. A definition that pins a model on
// a step is therefore not reflected in the defaults, only in the run.
type StepDefaultResult struct {
	Spec           api.Spec
	RuntimeProfile string
	Trace          []api.SpecLayer
	Provenance     map[string]api.FieldProvenance
	Warnings       []string
}

func (h *Host) StepDefaults(ctx context.Context, step Step) (StepDefaultResult, error) {
	layers, err := h.StepLayers(ctx, step)
	if err != nil {
		return StepDefaultResult{}, err
	}
	composed, err := api.ComposeSpecLayers(api.ResolveSpecOptions{Layers: layers.Layers, Saved: h.savedDefaults()})
	if err != nil {
		return StepDefaultResult{}, fmt.Errorf("compose defaults for step %s: %w", step.Name, runtimeConfigurationError(err))
	}
	defaults := StepDefaultResult{Spec: composed.Spec, Trace: composed.Trace, Provenance: composed.Provenance, Warnings: composed.Warnings}
	if layers.Profile != nil {
		defaults.RuntimeProfile = layers.Profile.Profile.ID
	}
	return defaults, nil
}

// StepLayers returns authored defaults before saved settings and model normalization.
func (h *Host) StepLayers(ctx context.Context, step Step) (runtimeprofiles.LayerResult, error) {
	definition, err := h.promptFor(step)
	if err != nil {
		return runtimeprofiles.LayerResult{}, &ConfigurationError{Err: err}
	}
	var prompt PromptLayerResult
	if definition.Class != types.ModeVerify {
		if prompt, err = PromptLayers(h.WorkDir, nil, definition); err != nil {
			return runtimeprofiles.LayerResult{}, err
		}
	}
	layers, err := h.profileLayers(ctx, LayerInput{
		RuntimeProfile: h.profileSelection(step.Name, "", prompt.RuntimeProfile),
		Config:         h.Config, Step: step.Name, Frontmatter: prompt.Layers, Host: h.Kind,
	})
	if err != nil {
		return runtimeprofiles.LayerResult{}, fmt.Errorf("resolve defaults for step %s: %w", step.Name, err)
	}
	return layers, nil
}
