package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

// renderedSpec projects the spec a run was dispatched with into Captain's
// `rendered_spec` jsonb column, stamping the issue's fixture as the workflow's
// definition of done. Nothing else puts the fixture on the dispatched spec, and
// nothing in Captain executes it — `api.Verify.Fixture` is declared for the
// schema and run only by gavel — so this is the durable record of what a later
// `todos check` replays.
//
// Setup.Env is `json:"-"` and deliberately does not round-trip: it is a
// shell.Prepare output, re-derived on the next run. The persisted spec says
// where the work landed, not the environment assembled for it.
func renderedSpec(spec api.Spec, fixture string) (map[string]any, error) {
	if strings.TrimSpace(fixture) != "" {
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
