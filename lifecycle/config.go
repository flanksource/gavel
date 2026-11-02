package lifecycle

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

func Read(config verify.LifecycleConfig, dir string) (Lifecycle, error) {
	var overlay Lifecycle
	if err := config.Validate(); err != nil {
		return overlay, err
	}
	var data []byte
	var err error
	source := "lifecycle"
	if config.File != "" {
		path, pathErr := utils.ResolveWithin(dir, config.File)
		if pathErr != nil {
			return overlay, pathErr
		}
		data, err = os.ReadFile(path)
		source = path
	} else {
		data, err = json.Marshal(config)
	}
	if err != nil {
		return overlay, err
	}
	if err := verify.DecodeYAML(data, &overlay, verify.DecodeOptions{Source: source}); err != nil {
		return overlay, fmt.Errorf("decode lifecycle: %w", err)
	}
	seen := map[string]bool{}
	for _, step := range overlay.Steps {
		if seen[step.Name] {
			return overlay, fmt.Errorf("step %q is declared twice", step.Name)
		}
		seen[step.Name] = true
	}
	return overlay, nil
}
