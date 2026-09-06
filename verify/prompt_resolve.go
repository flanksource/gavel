package verify

import (
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
)

type PromptResolveOptions struct {
	Base          api.Spec
	DefaultPrompt string
	FileSource    *string
	Data          map[string]any
	Dir           string
	Request       api.Spec
	Saved         *captainconfig.AIDefaults
	Name          string
	RequireModel  bool
	Normalize     func(api.Spec) (api.SpecNormalization, error)
}

func (s PromptSpec) Resolve(opts PromptResolveOptions) (api.ResolvedSpec, error) {
	layers, err := s.Layers(opts)
	if err != nil {
		return api.ResolvedSpec{}, err
	}
	resolved, err := api.ResolveSpecLayers(api.ResolveSpecOptions{Layers: layers, Saved: opts.Saved, RequireModel: opts.RequireModel, Normalize: opts.Normalize})
	if err != nil {
		return api.ResolvedSpec{}, fmt.Errorf("resolve %s prompt runtime: %w", opts.Name, err)
	}
	if opts.RequireModel {
		if err := resolved.Spec.Validate(); err != nil {
			return api.ResolvedSpec{}, fmt.Errorf("resolved prompt spec: %w", err)
		}
	}
	return resolved, nil
}

// Layers renders authored inputs without resolving their model or applying saved defaults.
func (s PromptSpec) Layers(opts PromptResolveOptions) ([]api.SpecLayer, error) {
	if opts.FileSource != nil && s.File == "" {
		return nil, fmt.Errorf("captured prompt file content requires a file reference")
	}
	if s.RuntimeProfile != "" {
		return nil, fmt.Errorf("runtime profiles require TODO lifecycle resolution")
	}
	var layers []api.SpecLayer
	if strings.TrimSpace(opts.Dir) != "" {
		var directory api.Spec
		directory.SetCwd(opts.Dir)
		layers = append(layers, api.SpecLayer{Name: "working directory", Source: api.SpecLayerSourcePreset, Scope: api.SpecLayerGlobal, Spec: directory})
	}
	layers = append(layers, api.SpecLayer{Name: ".gavel.yaml ai", Source: api.SpecLayerSourcePreset, Scope: api.SpecLayerGlobal, Spec: opts.Base})
	addPrompt := func(name, source string) error {
		rendered, err := RenderPromptSpec(source, opts.Data, PromptSpecOptions{Declared: true})
		if err != nil {
			return fmt.Errorf("render %s: %w", name, err)
		}
		if rendered.RuntimeProfile != "" {
			return fmt.Errorf("runtime profiles require TODO lifecycle resolution")
		}
		layers = append(layers, api.PromptSpecLayer(name, rendered.Spec))
		return nil
	}
	if err := addPrompt("built-in "+opts.Name+" prompt", opts.DefaultPrompt); err != nil {
		return nil, err
	}
	if s.File != "" {
		var source string
		if opts.FileSource != nil {
			source = *opts.FileSource
		} else {
			raw, err := os.ReadFile(s.resolvedFilePath(opts.Dir))
			if err != nil {
				return nil, fmt.Errorf("read prompt override file: %w", err)
			}
			source = string(raw)
		}
		if err := addPrompt(s.resolvedFilePath(opts.Dir), source); err != nil {
			return nil, err
		}
	}
	op := s.Spec
	if strings.TrimSpace(op.Prompt.User) != "" {
		if err := addPrompt("inline "+opts.Name+" prompt", op.Prompt.User); err != nil {
			return nil, err
		}
		op.Prompt.User = layers[len(layers)-1].Spec.Prompt.User
	}
	layers = append(layers, api.PromptSpecLayer(".gavel.yaml "+opts.Name, op))
	layers = append(layers, api.RequestSpecLayer("request", opts.Request))
	return layers, nil
}
