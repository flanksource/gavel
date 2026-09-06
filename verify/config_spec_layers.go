package verify

import (
	"maps"
	"slices"

	"github.com/flanksource/captain/pkg/api"
)

func validateConfigSpecLayers(cfg GavelConfig, path string) error {
	layers := []api.SpecLayer{
		{Name: path + " ai", Source: api.SpecLayerSourcePreset, Scope: api.SpecLayerGlobal, Spec: cfg.AI},
		api.PromptSpecLayer(path+" todos.run", cfg.Todos.Run.Spec),
		api.PromptSpecLayer(path+" todos.plan", cfg.Todos.Plan.Spec),
		api.PromptSpecLayer(path+" todos.triage", cfg.Todos.Triage.Spec),
		api.PromptSpecLayer(path+" todos.verify", cfg.Todos.Verify.Spec),
		{
			Name: path + " todos.timeout", Source: api.SpecLayerSourcePreset, Scope: api.SpecLayerContext,
			Constraints: api.RuntimeConstraints{Limits: api.RunLimits{Budget: api.Budget{Timeout: cfg.Todos.Timeout}}},
		},
	}
	for _, name := range slices.Sorted(maps.Keys(cfg.Todos.Steps)) {
		layers = append(layers, api.PromptSpecLayer(path+" todos.steps."+name, cfg.Todos.Steps[name].Spec))
	}
	return api.ValidateSpecLayers(layers...)
}
