package types

import (
	"slices"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/fixtures"
)

// resolveAIStepSpec resolves the grader runtime of an acceptance-criteria
// checklist step.
func resolveAIStepSpec(fixture fixtures.FixtureTest, opts fixtures.RunOptions, schema *checklistResponse) (api.ResolvedSpec, ai.AgentConfig, error) {
	return resolveAISpec(fixture, opts, "fixture.checklist", api.Prompt{
		User:   buildChecklistPrompt(fixture, fixtureRepoPath(fixture, opts), checklistItems(fixture), opts.Changed, opts.PreviousResults),
		Source: "fixtures.ai-step", Schema: schema,
	})
}

// resolveAISpec resolves the runtime of an AI step around prompt. A serialized
// grader or an explicit runner spec is authoritative. Authored standalone
// fixtures receive raw layers so saved defaults see the final model.
func resolveAISpec(fixture fixtures.FixtureTest, opts fixtures.RunOptions, layer string, prompt api.Prompt) (api.ResolvedSpec, ai.AgentConfig, error) {
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
		api.RequestSpecLayer(layer, api.Spec{Prompt: prompt}),
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
	return api.ResolvedSpec{Spec: composed.Spec, Trace: composed.Trace,
		Provenance: composed.Provenance, Warnings: warnings}, nil
}
