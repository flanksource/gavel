package todos

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestParseAcceptanceCriteriaReadsChecklistItems(t *testing.T) {
	body := strings.Join([]string{
		"Some description.",
		"",
		"## Acceptance Criteria",
		"",
		"- [ ] The issue's definition of done is met",
		"- [x] Export streams NDJSON for payloads over 10k rows",
		"",
		"## Verification",
		"- [ ] not-a-criterion-here",
	}, "\n")

	got := ParseAcceptanceCriteria(body)
	if len(got) != 2 {
		t.Fatalf("parsed %d criteria, want 2: %#v", len(got), got)
	}
	if got[0].Text != "The issue's definition of done is met" || got[0].Done {
		t.Errorf("first criterion = %#v, want unchecked with full text", got[0])
	}
	if got[1].Text != "Export streams NDJSON for payloads over 10k rows" || !got[1].Done {
		t.Errorf("second criterion = %#v, want checked", got[1])
	}
}

func TestCriteriaSectionRoundTrips(t *testing.T) {
	criteria := []types.AcceptanceCriterion{
		{Text: "New logic includes tests"},
		{Text: "Returns 400 on invalid input"},
		{Text: "Done item", Done: true},
	}
	section := RenderCriteriaSection(criteria)
	got := ParseAcceptanceCriteria(section)
	if len(got) != len(criteria) {
		t.Fatalf("round-trip parsed %d, want %d", len(got), len(criteria))
	}
	for i := range criteria {
		if got[i].Text != criteria[i].Text || got[i].Done != criteria[i].Done {
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
