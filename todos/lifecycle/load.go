package lifecycle

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/verify"
	"gopkg.in/yaml.v3"
)

//go:embed todos.yaml
var defaultYAML string

// Default is the embedded todo lifecycle.
func Default() (Lifecycle, error) {
	return Parse([]byte(defaultYAML))
}

// Parse decodes a lifecycle document. Unknown keys are errors: a misspelt
// `outcome:` that decoded to nothing would be a step with no transitions.
func Parse(data []byte) (Lifecycle, error) {
	var def Lifecycle
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&def); err != nil {
		return Lifecycle{}, fmt.Errorf("parse lifecycle: %w", err)
	}
	if err := def.Validate(); err != nil {
		return Lifecycle{}, err
	}
	return def, nil
}

// Load resolves the lifecycle a project runs: the embedded default with the
// project's `todos.lifecycle` override merged over it. The override is a file
// path (relative to workDir) or an inline definition; a step it names replaces
// the default step of that name wholesale, a step it adds is appended, and its
// subject declarations are added to the default's. A `verify` step must survive
// the merge.
func Load(workDir string) (Lifecycle, error) {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return Lifecycle{}, fmt.Errorf("load .gavel.yaml: %w", err)
	}
	return LoadWith(cfg.Todos.Lifecycle, workDir)
}

// LoadWith is Load for a config already in hand.
func LoadWith(override verify.LifecycleConfig, workDir string) (Lifecycle, error) {
	def, err := Default()
	if err != nil {
		return Lifecycle{}, err
	}
	if override.IsZero() {
		return def, nil
	}
	if err := override.Validate(); err != nil {
		return Lifecycle{}, err
	}
	overlay, err := overlayFrom(override, workDir)
	if err != nil {
		return Lifecycle{}, err
	}
	merged := Merge(def, overlay)
	if err := merged.Validate(); err != nil {
		return Lifecycle{}, fmt.Errorf("todos.lifecycle: %w", err)
	}
	return merged, nil
}

// overlay is a partial lifecycle: an override need not be a complete
// definition, so it is decoded without Validate.
func overlayFrom(override verify.LifecycleConfig, workDir string) (Lifecycle, error) {
	if override.File != "" {
		path := override.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(workDir, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return Lifecycle{}, fmt.Errorf("todos.lifecycle file: %w", err)
		}
		var overlay Lifecycle
		decoder := yaml.NewDecoder(strings.NewReader(string(data)))
		decoder.KnownFields(true)
		if err := decoder.Decode(&overlay); err != nil {
			return Lifecycle{}, fmt.Errorf("todos.lifecycle file %s: %w", override.File, err)
		}
		return overlay, nil
	}
	// The inline form arrived as generic YAML through the config; round-trip it
	// so it is decoded by the same strict decoder as a file.
	data, err := yaml.Marshal(map[string]any{
		"name": override.Name, "subject": override.Subject, "steps": override.Steps,
	})
	if err != nil {
		return Lifecycle{}, fmt.Errorf("todos.lifecycle: %w", err)
	}
	var overlay Lifecycle
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&overlay); err != nil {
		return Lifecycle{}, fmt.Errorf("todos.lifecycle: %w", err)
	}
	return overlay, nil
}

// Merge overlays a partial lifecycle onto a base. Steps merge by name and keep
// the base's order; new steps are appended in the overlay's order.
func Merge(base, overlay Lifecycle) Lifecycle {
	merged := Lifecycle{Name: base.Name, Subject: map[string]string{}}
	if overlay.Name != "" {
		merged.Name = overlay.Name
	}
	for name, typ := range base.Subject {
		merged.Subject[name] = typ
	}
	for name, typ := range overlay.Subject {
		merged.Subject[name] = typ
	}
	replaced := map[string]Step{}
	for _, step := range overlay.Steps {
		replaced[step.Name] = step
	}
	seen := map[string]bool{}
	for _, step := range base.Steps {
		if override, ok := replaced[step.Name]; ok {
			step = override
		}
		merged.Steps = append(merged.Steps, step)
		seen[step.Name] = true
	}
	for _, step := range overlay.Steps {
		if !seen[step.Name] {
			merged.Steps = append(merged.Steps, step)
		}
	}
	return merged
}
