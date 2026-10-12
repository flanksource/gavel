package bulk

import (
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

// MergeFlags are the parameters of the `merge` bulk action.
//
// Every one of them is optional, which is a requirement and not a convenience:
// the dashboard's selection toolbar dispatches an action with no parameters
// unless they are a closed set of values, so a merge started from a checked set
// of rows arrives here empty. The survivor therefore defaults to the first
// reference in the request — the order the selection was sent in — rather than
// to a flag a UI has no way to supply.
type MergeFlags struct {
	Into   string `flag:"into" help:"Short id of the TODO to merge into; defaults to the first one selected"`
	DryRun bool   `flag:"dry-run" help:"Report the proposed merge without writing anything"`
	Model  string `flag:"model" help:"Override the model for this merge, as the compact mode:model:effort form"`
	Effort string `flag:"effort" help:"Reasoning effort" enum:"low,medium,high,xhigh"`
}

func (MergeFlags) ClickyActionFlags() {}

// Spec projects the caller's overrides onto the top resolution layer. An empty
// field means "whatever .gavel.yaml already says", not a default asserted here.
func (f MergeFlags) Spec() api.Spec {
	spec := api.Spec{}
	if model := strings.TrimSpace(f.Model); model != "" {
		spec.Name = model
	}
	if effort := strings.ToLower(strings.TrimSpace(f.Effort)); effort != "" {
		spec.Effort = api.Effort(effort)
	}
	return spec
}
