package runtime

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
)

type renderedSpecOptions struct {
	Spec           api.Spec
	Fixture        string
	RuntimeProfile *runtimeprofiles.Resolution
	SpecTrace      []api.SpecLayer
	Previous       map[string]any
}

// renderedSpec projects the spec a run was dispatched with into Captain's
// `rendered_spec` jsonb column. A spec that declares no fixture of its own gets
// the issue's stamped on as the workflow's definition of done, so the durable
// record names what a later `todos check` replays; a spec that already declares
// one — the lifecycle step expanded `{{subject.verification.document}}` into
// it, or a caller supplied its own — is recorded as dispatched. Overwriting it
// would persist a document the run never executed.
//
// Setup.Env is `json:"-"` and deliberately does not round-trip: it is a
// shell.Prepare output, re-derived on the next run. The persisted spec says
// where the work landed, not the environment assembled for it.
func renderedSpec(options renderedSpecOptions) (map[string]any, error) {
	spec, fixture := options.Spec, options.Fixture
	if strings.TrimSpace(fixture) != "" && !declaresFixture(spec) {
		// Merged, not assigned: Workflow is a pointer shared with the live runner,
		// and merge clones both sides rather than writing through it.
		spec = spec.Merge(api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{Fixture: fixture}}})
	}
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal rendered spec: %w", err)
	}
	var rendered map[string]any
	if err := json.Unmarshal(data, &rendered); err != nil {
		return nil, fmt.Errorf("decode rendered spec: %w", err)
	}
	provenance, err := renderedSpecProvenance(options)
	if err != nil {
		return nil, err
	}
	maps.Copy(rendered, provenance)
	return rendered, nil
}

func renderedSpecProvenance(options renderedSpecOptions) (map[string]any, error) {
	values := map[string]any{}
	for _, key := range []string{"runtimeProfile", "specTrace"} {
		if value, ok := options.Previous[key]; ok {
			values[key] = value
		}
	}
	if profile := options.RuntimeProfile; profile != nil {
		values["runtimeProfile"] = struct {
			Profile runtimeprofiles.Profile  `json:"profile"`
			Presets []runtimeprofiles.Preset `json:"presets"`
		}{Profile: profile.Profile, Presets: profile.Presets}
	}
	if len(options.SpecTrace) > 0 {
		values["specTrace"] = options.SpecTrace
	}
	data, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("marshal rendered spec provenance: %w", err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("decode rendered spec provenance: %w", err)
	}
	return metadata, nil
}

func declaresFixture(spec api.Spec) bool {
	return spec.Workflow != nil && spec.Workflow.Verify != nil && strings.TrimSpace(spec.Workflow.Verify.Fixture) != ""
}
