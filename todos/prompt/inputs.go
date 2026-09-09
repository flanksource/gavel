package prompt

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/todos/types"
	"github.com/invopop/jsonschema"
)

// RunInputs are the schema-driven inputs of a run-mode prompt: the test/lint
// gates the run must pass, in the same shape fixture `yaml test` / `yaml lint`
// blocks unmarshal onto. The dashboard renders these with a JsonSchemaForm
// (clicky PromptDialog) instead of hand-coded option fields; the collected
// values configure the run's fixture verifier.
type RunInputs struct {
	Test *testrunner.RunOptions `json:"test,omitempty" yaml:"test,omitempty" jsonschema_description:"Test engine options gating the run (fixture 'yaml test' keys)"`
	Lint *lint.Options          `json:"lint,omitempty" yaml:"lint,omitempty" jsonschema_description:"Lint engine options gating the run (fixture 'yaml lint' keys)"`
}

// InputSchema returns the JSON schema of a mode's prompt inputs for the
// dashboard's schema-driven run form, or nil for modes with no inputs
// (plan investigates; verify derives everything from the issue). Field names
// follow the structs' YAML tags — the same wire contract as fixture yaml
// blocks, so collected values unmarshal straight onto the engine options.
func InputSchema(mode types.RunMode) ([]byte, error) {
	if mode != types.ModeRun {
		return nil, nil
	}
	reflector := jsonschema.Reflector{
		FieldNameTag:               "yaml",
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             false,
	}
	schema := reflector.Reflect(&RunInputs{})
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshal run input schema: %w", err)
	}
	return raw, nil
}
