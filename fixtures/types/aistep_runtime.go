package types

import (
	"slices"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/fixtures"
)

// A serialized grader or an explicit runner spec is authoritative. Authored
// standalone fixtures receive raw layers so saved defaults see the final model.
func resolveAIStepSpec(fixture fixtures.FixtureTest, opts fixtures.RunOptions, schema *checklistResponse) (api.ResolvedSpec, ai.AgentConfig, error) {
	var options api.ResolveSpecOptions
	snapshot := false
	switch {
	case fixture.AI != nil && fixture.AI.Spec != nil:
		snapshot = true
		options.Layers = []api.SpecLayer{api.PromptSpecLayer("fixture.ai.spec", *fixture.AI.Spec)}
	case opts.Spec != nil:
		snapshot = true
		options.Layers = []api.SpecLayer{api.PromptSpecLayer("fixture.runner.spec", *opts.Spec)}
	case opts.Runtime != nil:
		options = *opts.Runtime
		options.Layers = slices.Clone(opts.Runtime.Layers)
	}
	for i := range options.Layers {
		options.Layers[i].Spec = fixtures.GraderSpec(options.Layers[i].Spec)
	}
	options.Layers = append(options.Layers,
		api.RequestSpecLayer("fixture.ai", fixture.AI.SpecOverride()),
		api.RequestSpecLayer("fixture.checklist", api.Spec{Prompt: api.Prompt{
			User:   buildChecklistPrompt(fixture, fixtureRepoPath(fixture, opts), checklistItems(fixture), opts.Changed),
			Source: "fixtures.ai-step", Schema: schema,
		}}),
	)
	var resolved api.ResolvedSpec
	var err error
	if snapshot {
		resolved, err = composeAIStepSnapshot(options)
	} else {
		options.RequireModel = true
		resolved, err = api.ResolveSpecLayers(options)
	}
	if err != nil {
		return api.ResolvedSpec{}, ai.AgentConfig{}, err
	}
	return resolved, fixture.AI.ToAgentConfig(resolved.Spec), nil
}

func composeAIStepSnapshot(options api.ResolveSpecOptions) (api.ResolvedSpec, error) {
	composed, err := api.ComposeSpecLayers(options)
	if err != nil {
		return api.ResolvedSpec{}, err
	}
	if err := composed.Spec.Validate(); err != nil {
		return api.ResolvedSpec{}, err
	}
	validation := api.Spec{}.Merge(composed.Spec)
	validation.Model, err = api.ResolveModel(validation.Model)
	if err != nil {
		return api.ResolvedSpec{}, err
	}
	warnings, err := api.ValidateRuntimeSpec(validation)
	if err != nil {
		return api.ResolvedSpec{}, err
	}
	return api.ResolvedSpec{Spec: composed.Spec, Trace: composed.Trace, Constraints: composed.Constraints,
		Provenance: composed.Provenance, Warnings: warnings}, nil
}
