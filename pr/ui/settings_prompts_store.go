package ui

import (
	"os"
	"path/filepath"
	"strings"

	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	promptregistry "github.com/flanksource/gavel/prompts/registry"
	"github.com/flanksource/gavel/verify"
)

// persistPromptOverride writes the validated text by source: inline stores the
// parsed api.Spec structurally in the layer's .gavel.yaml; file writes the
// referenced .prompt file with the config pointing at it. Called only after
// validation, so it never persists bad text. cfg is taken by pointer because ov
// points into it: mutating *ov and then saving *cfg must reflect the same
// struct, not a copy.
func persistPromptOverride(cfg *verify.GavelConfig, ov *verify.PromptSpec, write promptWrite) error {
	switch write.Source {
	case "inline":
		spec, err := promptTextToSpec(write.Text)
		if err != nil {
			return err
		}
		if bodyMatchesDefault(spec.Prompt.User, write.Default) {
			spec.Prompt.User = ""
		}
		*ov = verify.PromptSpec{Spec: spec}
	case "file":
		target := write.Path
		if !filepath.IsAbs(target) {
			target = filepath.Join(write.Dir, target)
		}
		if err := os.WriteFile(target, []byte(write.Text), 0o644); err != nil {
			return err
		}
		*ov = verify.PromptSpec{File: write.Path}
	}
	return verify.SaveGavelConfig(write.Dir, *cfg)
}

// bodyMatchesDefault reports whether body is the built-in default's body, in
// which case the runtime's TemplateSource keeps rendering the default.
func bodyMatchesDefault(body, def string) bool {
	_, defaultBody, _, err := promptregistry.ParsePromptSource(def)
	if err != nil {
		return false
	}
	return strings.TrimSpace(body) == strings.TrimSpace(defaultBody)
}

// promptSpecRaw returns the full .prompt document for one layer's override: the
// built-in default when unset, the file's contents for a file override, or the
// inline spec rendered as the document it runs as for an inline override.
func promptSpecRaw(ov *verify.PromptSpec, dir, def string) (string, error) {
	switch {
	case ov.IsEmpty():
		return def, nil
	case ov.File != "":
		return ov.TemplateSource(dir, def)
	default:
		return inlinePromptText(ov.Spec, def)
	}
}

// inlinePromptText renders an inline override as the .prompt document the
// runtime effectively runs: its spec keys (model, budget, effort, prompt.system,
// …) laid over the built-in default's frontmatter, with its own body — or, when
// it sets none, the default body TemplateSource falls back to.
func inlinePromptText(spec api.Spec, def string) (string, error) {
	body := spec.Prompt.User
	fm := specToMap(spec)
	if p, ok := fm["prompt"].(map[string]any); ok {
		delete(p, "user")
		if len(p) == 0 {
			delete(fm, "prompt")
		} else {
			fm["prompt"] = p
		}
	}
	if m, ok := fm["model"].(string); ok && m == "" {
		delete(fm, "model")
	}
	if body == "" {
		defaultDoc, err := dotprompt.Parse(def)
		if err != nil {
			// A default with templated frontmatter cannot be parsed before it is
			// rendered; the view keeps the override's keys over the default body.
			_, defaultBody, _, parseErr := promptregistry.ParsePromptSource(def)
			if parseErr != nil {
				return "", parseErr
			}
			body = defaultBody
		} else {
			body = defaultDoc.Body
			fm = overlayFrontmatter(defaultDoc.Frontmatter, fm)
		}
	}
	if len(fm) == 0 {
		return body, nil
	}
	return (&dotprompt.Document{Frontmatter: fm, Body: body}).String()
}

// overlayFrontmatter lays an override's keys over a base frontmatter, merging
// the prompt block one level deep so prompt.system does not erase prompt.schema.
func overlayFrontmatter(base, over map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(over))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range over {
		basePrompt, baseOK := merged[key].(map[string]any)
		overPrompt, overOK := value.(map[string]any)
		if key == "prompt" && baseOK && overOK {
			prompt := make(map[string]any, len(basePrompt)+len(overPrompt))
			for k, v := range basePrompt {
				prompt[k] = v
			}
			for k, v := range overPrompt {
				prompt[k] = v
			}
			merged[key] = prompt
			continue
		}
		merged[key] = value
	}
	return merged
}

// promptTextToSpec parses a validated .prompt document into the api.Spec stored
// as an inline override: frontmatter → spec, body → prompt.user when the
// frontmatter does not set it.
func promptTextToSpec(text string) (api.Spec, error) {
	doc, err := dotprompt.Parse(text)
	if err != nil {
		return api.Spec{}, err
	}
	spec := doc.Spec
	if spec.Prompt.User == "" {
		spec.Prompt.User = doc.Body
	}
	return spec, nil
}
