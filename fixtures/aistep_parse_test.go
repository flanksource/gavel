package fixtures

import "testing"

func TestExtractChecklist(t *testing.T) {
	items := ExtractChecklist("- [ ] first criterion\n- [x] second criterion\n- a plain bullet, not a task\n")
	if len(items) != 2 {
		t.Fatalf("got %d checklist items, want 2: %+v", len(items), items)
	}
	if items[0].Text != "first criterion" || items[0].Checked {
		t.Errorf("item 0 = %+v, want {first criterion false}", items[0])
	}
	if items[1].Text != "second criterion" || !items[1].Checked {
		t.Errorf("item 1 = %+v, want {second criterion true}", items[1])
	}
}

func TestParseAIFixture(t *testing.T) {
	content := "# Verify feature X\n\n" +
		"Some prose describing the change.\n\n" +
		"```prompt\nFocus on the new parser path.\n```\n\n" +
		"## Acceptance Criteria\n" +
		"- [ ] tests-added\n" +
		"- [x] No secret is logged\n\n" +
		"- cel: json.score >= 80\n"

	frontMatter := &FrontMatter{AI: &FixtureAIConfig{Model: "claude-code-sonnet"}}
	tree, err := ParseMarkdownContentWithTree("verify-x.md", content, ".", frontMatter)
	if err != nil {
		t.Fatalf("ParseMarkdownContentWithTree: %v", err)
	}

	var step *FixtureTest
	tree.Walk(func(node *FixtureNode) {
		if node.Test != nil && node.Test.IsAIStep() {
			step = node.Test
		}
	})
	if step == nil {
		t.Fatal("no AI step parsed from ai: front matter fixture")
	}

	if step.Name != "Verify feature X" {
		t.Errorf("Name = %q, want %q", step.Name, "Verify feature X")
	}
	if step.AIStep.Prompt != "Focus on the new parser path." {
		t.Errorf("Prompt = %q, want the prompt block body", step.AIStep.Prompt)
	}
	if len(step.AIStep.Criteria) != 2 {
		t.Fatalf("criteria = %d, want 2: %+v", len(step.AIStep.Criteria), step.AIStep.Criteria)
	}
	if step.AIStep.Criteria[0].Text != "tests-added" || step.AIStep.Criteria[1].Text != "No secret is logged" {
		t.Errorf("criteria texts = %+v", step.AIStep.Criteria)
	}
	if step.Expected.CEL != "json.score >= 80" {
		t.Errorf("Expected.CEL = %q, want %q", step.Expected.CEL, "json.score >= 80")
	}
	if step.AIStep.Description == "" {
		t.Error("description should capture the prose paragraph")
	}
}
