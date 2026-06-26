package todos

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestCriteriaSchemaToleratesObjectItems(t *testing.T) {
	// A model that ignores the string item type and returns objects must not
	// fail the whole response.
	raw := `{
		"applicable_checks": ["tests-added", {"id": "definition-of-done"}],
		"custom_criteria": [
			{"criterion": "Streams NDJSON for payloads over 10k rows"},
			"Returns 400 on invalid input",
			{"text": "  "}
		]
	}`
	var schema criteriaSchema
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		t.Fatalf("unmarshal should tolerate object items: %v", err)
	}

	got := criteriaFromSchema(&schema)
	want := []types.AcceptanceCriterion{
		{CheckID: "tests-added", Text: "New/modified logic includes corresponding test additions"},
		{CheckID: "definition-of-done", Text: "The original issue's definition of done is fully met by the change"},
		{Text: "Streams NDJSON for payloads over 10k rows"},
		{Text: "Returns 400 on invalid input"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d criteria, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("criterion %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestParseAcceptanceCriteriaClassifiesStaticAndCustom(t *testing.T) {
	body := strings.Join([]string{
		"Some description.",
		"",
		"## Acceptance Criteria",
		"",
		"- [ ] definition-of-done: The issue's definition of done is met",
		"- [x] Export streams NDJSON for payloads over 10k rows",
		"",
		"## Verification",
		"- [ ] not-a-criterion-here",
	}, "\n")

	got := ParseAcceptanceCriteria(body)
	if len(got) != 2 {
		t.Fatalf("parsed %d criteria, want 2: %#v", len(got), got)
	}
	if got[0].CheckID != "definition-of-done" {
		t.Errorf("first criterion CheckID = %q, want definition-of-done", got[0].CheckID)
	}
	if got[0].Text != "The issue's definition of done is met" {
		t.Errorf("first criterion Text = %q (id prefix should be stripped)", got[0].Text)
	}
	if got[1].CheckID != "" || !got[1].Done {
		t.Errorf("second criterion should be custom and done: %#v", got[1])
	}
}

func TestCriteriaSectionRoundTrips(t *testing.T) {
	criteria := []types.AcceptanceCriterion{
		{CheckID: "tests-added", Text: "New logic includes tests"},
		{Text: "Returns 400 on invalid input"},
		{Text: "Done item", Done: true},
	}
	section := RenderCriteriaSection(criteria)
	got := ParseAcceptanceCriteria(section)
	if len(got) != len(criteria) {
		t.Fatalf("round-trip parsed %d, want %d", len(got), len(criteria))
	}
	for i := range criteria {
		if got[i] != criteria[i] {
			t.Errorf("criterion %d round-trip = %#v, want %#v", i, got[i], criteria[i])
		}
	}
}

func TestUpsertCriteriaSectionIsIdempotent(t *testing.T) {
	body := "Issue description.\n\n## Verification\n- run tests\n"
	criteria := []types.AcceptanceCriterion{{Text: "Works"}}

	once := UpsertCriteriaSection(body, criteria)
	if !strings.Contains(once, "## Acceptance Criteria") {
		t.Fatalf("criteria section not added:\n%s", once)
	}
	// Inserted above the existing Verification section.
	if strings.Index(once, "## Acceptance Criteria") > strings.Index(once, "## Verification") {
		t.Errorf("criteria should precede Verification:\n%s", once)
	}

	twice := UpsertCriteriaSection(once, []types.AcceptanceCriterion{{Text: "Works"}, {Text: "And more"}})
	if strings.Count(twice, "## Acceptance Criteria") != 1 {
		t.Errorf("upsert must not duplicate the section:\n%s", twice)
	}
	if !strings.Contains(twice, "And more") {
		t.Errorf("upsert must replace contents:\n%s", twice)
	}
}
