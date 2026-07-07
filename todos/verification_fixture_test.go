package todos

import (
	"strings"
	"testing"
)

func TestExtractVerificationFixtureReturnsSectionBody(t *testing.T) {
	body := strings.Join([]string{
		"Some description.",
		"",
		"## Verification",
		"",
		"```test",
		"query: SELECT 1",
		"```",
		"",
		"## Custom Validations",
		"- not part of verification",
	}, "\n")

	got := ExtractVerificationFixture(body)
	want := strings.Join([]string{
		"```test",
		"query: SELECT 1",
		"```",
	}, "\n")
	if got != want {
		t.Errorf("ExtractVerificationFixture = %q, want %q", got, want)
	}
}

func TestExtractVerificationFixtureEmptyWhenSectionAbsent(t *testing.T) {
	if got := ExtractVerificationFixture("Just a description, no sections."); got != "" {
		t.Errorf("ExtractVerificationFixture = %q, want empty", got)
	}
}

func TestUpsertVerificationFixtureInsertsBeforeOperationalSections(t *testing.T) {
	body := "Issue description.\n\n## Attempts\n\n| # |\n"
	once := UpsertVerificationFixture(body, "```test\nquery: SELECT 1\n```")
	if !strings.Contains(once, "## Verification") {
		t.Fatalf("verification section not added:\n%s", once)
	}
	if strings.Index(once, "## Verification") > strings.Index(once, "## Attempts") {
		t.Errorf("verification should precede Attempts:\n%s", once)
	}
	if !strings.Contains(once, "query: SELECT 1") {
		t.Errorf("verification section missing fixture content:\n%s", once)
	}
}

func TestUpsertVerificationFixtureReplacesInPlace(t *testing.T) {
	body := "Issue description.\n\n## Verification\n\n```test\nquery: SELECT 1\n```\n\n## Attempts\n\n| # |\n"

	updated := UpsertVerificationFixture(body, "```test\nquery: SELECT 2\n```")
	if strings.Count(updated, "## Verification\n") != 1 {
		t.Fatalf("upsert must not duplicate the section:\n%s", updated)
	}
	if strings.Contains(updated, "SELECT 1") {
		t.Errorf("upsert must replace prior contents:\n%s", updated)
	}
	if !strings.Contains(updated, "SELECT 2") {
		t.Errorf("upsert must contain new contents:\n%s", updated)
	}
	if !strings.Contains(updated, "## Attempts") {
		t.Errorf("upsert must preserve unrelated sections:\n%s", updated)
	}
}

func TestVerificationFixtureRoundTrips(t *testing.T) {
	fixture := "```test\nquery: SELECT 1\n```"
	body := UpsertVerificationFixture("Issue description.", fixture)
	got := ExtractVerificationFixture(body)
	if got != fixture {
		t.Errorf("round-trip = %q, want %q", got, fixture)
	}
}
