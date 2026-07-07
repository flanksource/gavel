package todos

import "strings"

// verificationFixtureSection is the "## Verification" section that holds the
// TODO's executable test fixtures (fixtures.FixtureNode trees), parsed by
// ParseTODO/ParseTODOContent into TODO.Verification. It is intentionally
// distinct from "## Verification Result" (see verification.go), which holds an
// AI-scored verdict rather than the fixture source itself.
const verificationFixtureSection = "Verification"

// verificationFixtureInsertBefore keeps a newly-added Verification fixture
// section above the operational history sections, mirroring criteriaInsertBefore.
var verificationFixtureInsertBefore = []string{
	"Custom Validations", "Latest Failure", "Verification Result",
	"Attempts", "Failure History",
}

// ExtractVerificationFixture returns the raw markdown inside a TODO body's
// "## Verification" section (excluding the heading line itself), so the
// dashboard's FixtureEditor can be seeded with just the fixture source.
func ExtractVerificationFixture(body string) string {
	lines := sectionLines(body, verificationFixtureSection)
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// UpsertVerificationFixture replaces (or inserts) the "## Verification" section
// in body with fixture, the fixture markdown edited via the FixtureEditor.
func UpsertVerificationFixture(body, fixture string) string {
	var b strings.Builder
	b.WriteString("## " + verificationFixtureSection + "\n\n")
	b.WriteString(strings.TrimSpace(fixture))
	b.WriteString("\n")
	return ReplaceOrAppendSection(body, verificationFixtureSection, b.String(), verificationFixtureInsertBefore...)
}
