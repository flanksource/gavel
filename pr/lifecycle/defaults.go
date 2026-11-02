package lifecycle

import (
	_ "embed"
	"fmt"
	"slices"

	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/verify"
)

//go:embed pr.yaml
var defaultDocument []byte

func Default() (lifecycle.Lifecycle, error) {
	var def lifecycle.Lifecycle
	if err := verify.DecodeYAML(defaultDocument, &def, verify.DecodeOptions{Source: "built-in PR lifecycle"}); err != nil {
		return def, err
	}
	_, err := lifecycle.New(def)
	return def, err
}

func Validate(def lifecycle.Lifecycle) error {
	if err := def.Validate(); err != nil {
		return err
	}
	for _, step := range def.Steps {
		if len(step.Tools) > 0 {
			return fmt.Errorf("PR step %s: TODO tool sets are unsupported; configure Captain tool preferences in spec", step.Name)
		}
		for _, outcome := range step.Outcomes {
			if !slices.Contains([]string{"succeeded", "failed", "cancelled"}, outcome.Status) {
				return fmt.Errorf("PR step %s: outcome status %q must be succeeded, failed or cancelled", step.Name, outcome.Status)
			}
		}
	}
	return nil
}
