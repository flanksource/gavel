package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
)

// RestrictHostPermissions decides what the dashboard's own layer says before it
// is folded with the rest. The dashboard contributes `default` because it can
// broker an approval, but a stack that had already settled on a read-only or
// approval-free posture does not need brokering — so the host adopts that
// posture instead of overwriting it.
//
// This is not a constraint on the layers below: it is the host choosing its own
// value from what it can see, which is what any layer that must not widen an
// authored posture does. Every layer here still only defaults, and any layer
// after this one is free to name a different posture.
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
