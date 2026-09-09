package lifecycle

import (
	"fmt"

	"github.com/flanksource/captain/pkg/api"
)

func constrainPermissionLayers(layers []api.SpecLayer) ([]api.SpecLayer, error) {
	constrained := append([]api.SpecLayer(nil), layers...)
	for i := range constrained {
		layer := constrained[i]
		if layer.Source == api.SpecLayerSourceRequest && layer.Name == "request" {
			continue
		}
		var err error
		constrained[i], err = api.ConstrainSpecLayerPermissions(layer)
		if err != nil {
			return nil, fmt.Errorf("project permission constraints: %w", err)
		}
	}
	return constrained, nil
}

// RestrictHostPermissions keeps the dashboard's approval default from widening
// authored plan/dontAsk modes, while broader modes still require host approval.
func RestrictHostPermissions(layers []api.SpecLayer) []api.SpecLayer {
	ordered := api.OrderSpecLayers(layers...)
	var baseline api.Spec
	for i := range ordered {
		layer := &ordered[i]
		if layer.Source == api.SpecLayerSourceRequest && layer.Name == "host dashboard" && layer.Spec.Permissions.Mode == api.PermissionDefault {
			switch baseline.Permissions.Mode {
			case api.PermissionPlan, api.PermissionDontAsk:
				layer.Spec.Permissions.Mode = baseline.Permissions.Mode
			}
		}
		baseline = baseline.Merge(layer.Spec)
	}
	return ordered
}
