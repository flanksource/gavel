package fixtures

import "testing"

// TestParseRunnerSteps verifies that `yaml test` / `yaml lint` (and the bare
// `test` / `lint`) fences are parsed into RunnerStep fixtures carrying the raw
// YAML body, named after their parent heading.
func TestParseRunnerSteps(t *testing.T) {
	content := "# Suite\n\n" +
		"## Run tests\n\n" +
		"```yaml test\npaths: [./testrunner/...]\nshow-passed: true\n```\n\n" +
		"## Lint\n\n" +
		"```lint\nlinters: [golangci-lint]\n```\n"

	tree, err := ParseMarkdownContentWithTree("suite.md", content, ".", &FrontMatter{})
	if err != nil {
		t.Fatalf("ParseMarkdownContentWithTree: %v", err)
	}

	steps := map[string]*RunnerStepSpec{}
	tree.Walk(func(node *FixtureNode) {
		if node.Test != nil && node.Test.IsRunnerStep() {
			steps[node.Test.Name] = node.Test.RunnerStep
		}
	})

	if len(steps) != 2 {
		t.Fatalf("got %d runner steps, want 2: %+v", len(steps), steps)
	}

	testStep, ok := steps["Run tests"]
	if !ok {
		t.Fatalf("no runner step named %q; got %v", "Run tests", keysOf(steps))
	}
	if testStep.Kind != RunnerKindTest {
		t.Errorf("Kind = %q, want %q", testStep.Kind, RunnerKindTest)
	}
	if want := "paths: [./testrunner/...]\nshow-passed: true"; testStep.Config != want+"\n" && testStep.Config != want {
		t.Errorf("test Config = %q, want it to carry the yaml body", testStep.Config)
	}

	lintStep, ok := steps["Lint"]
	if !ok {
		t.Fatalf("no runner step named %q", "Lint")
	}
	if lintStep.Kind != RunnerKindLint {
		t.Errorf("Kind = %q, want %q", lintStep.Kind, RunnerKindLint)
	}
}

func keysOf(m map[string]*RunnerStepSpec) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
