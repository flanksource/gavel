package lifecycle

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
)

type runtimeFields struct {
	Name    string
	Paths   []string
	Replace bool
}

func recordRuntimeFields(provenance map[string]api.FieldProvenance, fields runtimeFields) map[string]api.FieldProvenance {
	if provenance == nil {
		provenance = map[string]api.FieldProvenance{}
	}
	for _, path := range fields.Paths {
		source := api.FieldSource{Kind: api.FieldSourceContext, Name: fields.Name, Key: path}
		previous, exists := provenance[path]
		if !exists || fields.Replace {
			previous = api.FieldProvenance{Source: source}
		} else {
			previous.NormalizedBy = &source
		}
		provenance[path] = previous
	}
	return provenance
}

// Runtime deadlines are applied after saved defaults and restrictive limits.
func prepareRuntimeSpec(resolved *api.ResolvedSpec, class types.RunMode) (time.Duration, error) {
	before := resolved.Spec.Budget.Timeout
	timeout, err := ApplyTimeout(&resolved.Spec)
	if err != nil {
		return 0, err
	}
	if before != resolved.Spec.Budget.Timeout {
		resolved.Provenance = recordRuntimeFields(resolved.Provenance, runtimeFields{
			Name: "lifecycle runtime", Paths: []string{"/budget/timeout"},
		})
	}
	if class != types.ModeRun {
		if resolved.Spec.Workflow != nil {
			workflow := *resolved.Spec.Workflow
			resolved.Spec.Workflow = &workflow
		}
		for path := range resolved.Provenance {
			if path == "/workflow/commits" || strings.HasPrefix(path, "/workflow/commits/") {
				delete(resolved.Provenance, path)
			}
		}
	}
	ApplyClassInvariants(&resolved.Spec, class)
	return timeout, nil
}

func validateRequestEffort(model api.Model) error {
	if model.Effort == "" {
		return nil
	}
	expanded, err := model.Expand()
	if err != nil {
		return err
	}
	if expanded.Effort != model.Effort {
		return fmt.Errorf("effort %q conflicts with model %q, which pins effort %q; drop one", model.Effort, model.Name, expanded.Effort)
	}
	return nil
}
