package main

import (
	"os"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
)

// defaultAIRuntimeOptions mirrors the boolean defaults clicky sets on
// LintOptions' embedded captaincli.AIRuntimeOptions via the `default:"true"`
// flag tags. Call sites that do NOT receive parsed CLI flags (e.g. the
// interactive commit prompt) use this to keep behaviour in sync with
// `gavel lint --ai-fix`. Captain owns the canonical defaults; if it changes
// any of them this helper must be revisited.
func defaultAIRuntimeOptions() captaincli.AIRuntimeOptions {
	// The zero value enables everything: the ambient-context toggles are now
	// negative flags (No*), all defaulting to false.
	return captaincli.AIRuntimeOptions{}
}

type aiFixRequestOptions struct {
	Runtime captaincli.AIRuntimeOptions
	Layers  []api.SpecLayer
	Saved   captainconfig.Config
	Dir     string
}

func buildAIFixRequest(options aiFixRequestOptions) (captaincli.AIRuntimeResolved, error) {
	layers := []api.SpecLayer{api.PromptSpecLayer("repair host", api.Spec{Permissions: api.Permissions{Mode: api.PermissionAcceptEdits}})}
	layers = append(layers, options.Layers...)
	return options.Runtime.Resolve(captaincli.AIRuntimeResolveOptions{
		Layers: layers, Saved: options.Saved, Cwd: options.Dir, RequireModel: true,
	})
}

func newAIFixRenderer() *captaincli.EventRenderer {
	return captaincli.NewEventRenderer(os.Stderr)
}
