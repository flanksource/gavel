package fixtures

import (
	"encoding/json"
	"fmt"
)

// ParamsStepID is the id of the form step a fixture's `params:` lowers to.
// Its outputs are `{{.params.x}}` / CEL `params.x`, and it alone takes the
// values supplied up front (RunOptions.Params) before it asks.
const ParamsStepID = "params"

// FormKind is the kind name of the JSON-schema form step `params:` lowers to.
const FormKind = "form"

// ParamsStep is the synthetic first step a fixture's `params:` JSON Schema
// lowers to: a form titled after the fixture.
func ParamsStep(title string, schema json.RawMessage) (StepSpec, error) {
	config, err := json.Marshal(map[string]any{"title": title, "schema": schema})
	if err != nil {
		return StepSpec{}, fmt.Errorf("params: encode form config: %w", err)
	}
	return StepSpec{ID: ParamsStepID, Name: ParamsStepID, Kind: FormKind, Config: config}, nil
}

// lowerMarkdownParams prepends the `params:` form step to a markdown file's
// tree, so every row after it reads `{{.params.x}}` and is skipped when the
// form fails.
func lowerMarkdownParams(file *FixtureNode, frontMatter *FrontMatter, title, sourceDir string) error {
	if frontMatter == nil || frontMatter.Params == nil {
		return nil
	}
	schema, err := json.Marshal(frontMatter.Params)
	if err != nil {
		return fmt.Errorf("params: encode schema: %w", err)
	}
	spec, err := ParamsStep(title, schema)
	if err != nil {
		return err
	}
	test := &FixtureTest{Name: ParamsStepID, SourceDir: sourceDir, Step: &spec, FrontMatter: *frontMatter}
	node := &FixtureNode{Name: ParamsStepID, Type: TestNode, Test: test, Parent: file,
		Origin: &FixtureOrigin{File: file.Origin.File, Kind: "params"}}
	file.Children = append([]*FixtureNode{node}, file.Children...)
	return nil
}
