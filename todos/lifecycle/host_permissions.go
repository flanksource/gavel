package lifecycle

import (
	"fmt"
	"maps"
	"slices"

	"github.com/flanksource/captain/pkg/api"
)

// ValidateRequestPermissions checks the explicit request against the preceding
// effective permission layers. Callers supply layers in final precedence order.
// WORKAROUND(runtime-permission-constraints): Captain does not yet expose permission constraints.
// Correct fix: move this policy into Captain constraints.permissions and remove this guard when P4 lands.
// Ref: approved lifecycle migration plan Part 12 R5, discussed with user 2026-09-05.
func ValidateRequestPermissions(layers []api.SpecLayer) error {
	var baseline api.Spec
	var request, modeLayer api.SpecLayer
	toolLayers := map[string]api.SpecLayer{}
	for _, layer := range api.OrderSpecLayers(layers...) {
		if layer.Source == api.SpecLayerSourceRequest && layer.Name == "request" {
			if request.Name != "" {
				return fmt.Errorf("permission layers contain more than one explicit request")
			}
			request = layer
			continue
		}
		permissions := layer.Spec.Permissions
		if permissions.Mode != "" && (layer.Name != "host dashboard" || permissions.Mode != baseline.Permissions.Mode) {
			modeLayer = layer
		}
		for tool, policy := range permissions.Tools {
			if policy != "" {
				toolLayers[tool] = layer
			}
		}
		baseline = baseline.Merge(api.Spec{Permissions: permissions})
	}
	if request.Name == "" {
		return nil
	}
	effective := baseline.Merge(api.Spec{Permissions: request.Spec.Permissions})
	for _, tool := range slices.Sorted(maps.Keys(request.Spec.Permissions.Tools)) {
		if baseline.Permissions.Tools[tool] == api.ToolPolicyDeny && effective.Permissions.Tools[tool] != api.ToolPolicyDeny {
			return fmt.Errorf("request layer %q cannot widen permissions.tools.%s from deny in %s layer %q to %q",
				request.Name, tool, toolLayers[tool].Source, toolLayers[tool].Name, effective.Permissions.Tools[tool])
		}
	}
	if baseline.Permissions.Mode == "" || baseline.Permissions.Mode == effective.Permissions.Mode {
		return nil
	}
	order := []api.PermissionMode{api.PermissionPlan, api.PermissionDefault, api.PermissionAcceptEdits, api.PermissionBypass}
	before, after := slices.Index(order, baseline.Permissions.Mode), slices.Index(order, effective.Permissions.Mode)
	if before < 0 || after < 0 {
		return fmt.Errorf("request layer %q cannot replace permissions.mode %q in %s layer %q with incomparable mode %q",
			request.Name, baseline.Permissions.Mode, modeLayer.Source, modeLayer.Name, effective.Permissions.Mode)
	}
	if after > before {
		return fmt.Errorf("request layer %q cannot widen permissions.mode from %q in %s layer %q to %q",
			request.Name, baseline.Permissions.Mode, modeLayer.Source, modeLayer.Name, effective.Permissions.Mode)
	}
	return nil
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
		baseline = baseline.Merge(api.Spec{Permissions: layer.Spec.Permissions})
	}
	return ordered
}
