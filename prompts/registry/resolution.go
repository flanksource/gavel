package registry

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

func resolvePromptSpec(opts ResolveOptions, desc prompts.Prompt, override verify.PromptSpec) (api.ResolvedSpec, error) {
	layers, err := override.Layers(verify.PromptResolveOptions{
		Base: opts.Trace.Merged.AI, DefaultPrompt: desc.Default, Data: opts.Data,
		Dir: opts.Trace.TargetDir, Name: desc.ConfigPath,
	})
	if err != nil {
		return api.ResolvedSpec{}, err
	}
	for i := range layers {
		layers[i].ID = promptLayerID(layers[i], desc, override.ResolvedFilePath(opts.Trace.TargetDir))
		if opts.Draft != "" && (layers[i].ID == "operation" || layers[i].ID == "inline") {
			layers[i].Name = "draft " + desc.ConfigPath
			layers[i].ID = "draft"
		}
	}
	options := api.ResolveSpecOptions{Layers: layers, Saved: opts.Saved, RequireModel: !opts.Preview, Normalize: opts.Normalize}
	var resolved api.ResolvedSpec
	if opts.Preview {
		composed, err := api.ComposeSpecLayers(options)
		if err != nil {
			return api.ResolvedSpec{}, err
		}
		resolved = api.ResolvedSpec{Spec: composed.Spec, Constraints: composed.Constraints, Trace: composed.Trace,
			Provenance: composed.Provenance, Warnings: composed.Warnings}
	} else {
		resolved, err = api.ResolveSpecLayers(options)
		if err != nil {
			return api.ResolvedSpec{}, err
		}
	}
	for path, field := range resolved.Provenance {
		field.Source, err = configFieldSource(opts.Trace, desc.ConfigPath, field.Source)
		if err != nil {
			return api.ResolvedSpec{}, err
		}
		resolved.Provenance[path] = field
	}
	return resolved, nil
}

func promptLayerID(layer api.SpecLayer, desc prompts.Prompt, file string) string {
	switch layer.Name {
	case ".gavel.yaml ai":
		return "ai"
	case "built-in " + desc.ConfigPath + " prompt":
		return "default"
	case "inline " + desc.ConfigPath + " prompt":
		return "inline"
	case ".gavel.yaml " + desc.ConfigPath:
		return "operation"
	case file:
		return "file"
	default:
		return layer.Name
	}
}

// The shared fold owns field selection. Config traces only refine its merged
// ai/operation sources to the file that authored that exact field.
func configFieldSource(trace verify.GavelConfigTrace, configPath string, source api.FieldSource) (api.FieldSource, error) {
	if source.Kind != api.FieldSourceLayer || (source.LayerID != "ai" && source.LayerID != "operation" && source.LayerID != "inline") {
		return source, nil
	}
	field := source.Key
	if source.LayerID == "inline" {
		field = "/prompt/user"
	}
	for i := len(trace.Sources) - 1; i >= 0; i-- {
		config := trace.Sources[i]
		spec := config.Config.AI
		key := "ai"
		if source.LayerID != "ai" {
			op, err := promptSpecAt(config.Config, configPath)
			if err != nil {
				return api.FieldSource{}, err
			}
			spec, key = op.Spec, configPath
		}
		if !spec.Fields().Has(field) {
			continue
		}
		source.Name = config.Path
		if source.LayerID == "inline" {
			source.Key = key + ".prompt.user#" + source.Key
		} else {
			source.Key = key + strings.ReplaceAll(source.Key, "/", ".")
		}
		return source, nil
	}
	if len(trace.Sources) > 0 {
		return api.FieldSource{}, fmt.Errorf("no authored config source for %s field %s", configPath, field)
	}
	return source, nil
}
