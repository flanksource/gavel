package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

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
func renderedSpec(spec api.Spec, fixture string) (map[string]any, error) {
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
	return rendered, nil
}

func declaresFixture(spec api.Spec) bool {
	return spec.Workflow != nil && spec.Workflow.Verify != nil && strings.TrimSpace(spec.Workflow.Verify.Fixture) != ""
}
