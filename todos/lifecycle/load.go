package lifecycle

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/verify"
)

//go:embed todos.yaml
var defaultYAML string

// Default is the embedded todo lifecycle.
func Default() (lifecycle.Lifecycle, error) {
	return Parse([]byte(defaultYAML))
}

// Parse decodes a lifecycle document and validates its transitions.
func Parse(data []byte) (lifecycle.Lifecycle, error) {
	var def lifecycle.Lifecycle
	if err := verify.DecodeYAML(data, &def, verify.DecodeOptions{Source: "lifecycle"}); err != nil {
		return lifecycle.Lifecycle{}, fmt.Errorf("parse lifecycle: %w", err)
	}
	if err := Validate(def); err != nil {
		return lifecycle.Lifecycle{}, err
	}
	return lifecycle.Select(def)
}

// Load resolves the lifecycle a project runs: the embedded default with the
// project's `todos.lifecycle` override merged over it. The override is a file
// path inside workDir (relative to it, or absolute but still under it) or an
// inline definition; a step it names replaces
// the default step of that name wholesale, a step it adds is appended, and its
// subject declarations are added to the default's. A `verify` step must survive
// the merge.
func Load(workDir string) (lifecycle.Lifecycle, error) {
	trace, err := verify.LoadGavelConfigTrace(workDir)
	if err != nil {
		return lifecycle.Lifecycle{}, fmt.Errorf("load .gavel.yaml: %w", err)
	}
	def, err := Default()
	if err != nil {
		return def, err
	}
	for _, source := range trace.Sources {
		overlay, err := lifecycle.Read(source.Config.Todos.Lifecycle, filepath.Dir(source.Path))
		if err != nil {
			return def, err
		}
		def = lifecycle.Merge(def, overlay)
	}
	if err := Validate(def); err != nil {
		return def, err
	}
	def, err = lifecycle.Select(def)
	if err != nil {
		return def, err
	}
	_, err = lifecycle.New(def)
	return def, err
}

// LoadWith is Load for a config already in hand.
func LoadWith(override verify.LifecycleConfig, workDir string) (lifecycle.Lifecycle, error) {
	def, err := Default()
	if err != nil {
		return lifecycle.Lifecycle{}, err
	}
	if override.IsZero() {
		return def, nil
	}
	if err := override.Validate(); err != nil {
		return lifecycle.Lifecycle{}, err
	}
	overlay, err := lifecycle.Read(override, workDir)
	if err != nil {
		if override.File != "" {
			return lifecycle.Lifecycle{}, fmt.Errorf("todos.lifecycle file: %w", err)
		}
		return lifecycle.Lifecycle{}, err
	}
	merged := lifecycle.Merge(def, overlay)
	if err := Validate(merged); err != nil {
		return lifecycle.Lifecycle{}, fmt.Errorf("todos.lifecycle: %w", err)
	}
	return lifecycle.Select(merged)
}
