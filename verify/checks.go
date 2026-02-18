package verify

type Check struct {
	ID          string
	Category    string
	Description string
}

var AllChecks = []Check{
	// Completeness
	{ID: "tests-added", Category: "completeness", Description: "New/modified logic includes corresponding test additions"},
	{ID: "edge-cases-covered", Category: "completeness", Description: "Tests cover boundary values, empty inputs, nil/null, zero"},
	{ID: "error-paths-tested", Category: "completeness", Description: "Tests exercise error/failure paths, not just happy path"},
	{ID: "breaking-changes-noted", Category: "completeness", Description: "Breaking API/behavioral changes are documented"},
	{ID: "todos-resolved", Category: "completeness", Description: "No new TODO/FIXME/HACK without linked tracking issue"},
	{ID: "config-changes-documented", Category: "completeness", Description: "New env vars, flags, or config changes documented"},
	{ID: "migration-included", Category: "completeness", Description: "DB schema changes include a migration file"},

	// Code Quality
	{ID: "no-code-duplication", Category: "code-quality", Description: "No copy-pasted logic that should be extracted"},
	{ID: "single-responsibility", Category: "code-quality", Description: "Each function does one thing"},
	{ID: "no-commented-out-code", Category: "code-quality", Description: "No committed blocks of commented-out code"},
	{ID: "no-premature-optimization", Category: "code-quality", Description: "Code favors clarity over optimization without need"},
	{ID: "consistent-error-strategy", Category: "code-quality", Description: "Error handling follows project's established pattern"},

	// Testing
	{ID: "test-assertions-meaningful", Category: "testing", Description: "Tests assert specific expected values, not just no-error"},
	{ID: "no-flaky-patterns", Category: "testing", Description: "Tests don't depend on timing or non-deterministic ordering"},
	{ID: "mocking-minimal", Category: "testing", Description: "Prefer real implementations; mocks only for external deps"},
	{ID: "test-isolation", Category: "testing", Description: "Each test case is independent"},
	{ID: "no-test-logic-duplication", Category: "testing", Description: "Repeated setup uses table-driven tests or fixtures"},
	{ID: "negative-cases-tested", Category: "testing", Description: "Tests include invalid input, not-found, permission denied"},
	{ID: "regression-test-added", Category: "testing", Description: "Bug fixes include a test that catches the original bug"},

	// Naming
	{ID: "naming-consistent", Category: "naming", Description: "Identifiers follow project naming conventions and domain vocabulary"},
	{ID: "boolean-names-positive", Category: "naming", Description: "Booleans use positive phrasing (isValid not isNotInvalid)"},
	{ID: "abbreviations-avoided", Category: "naming", Description: "Names avoid unclear abbreviations (count not cnt)"},

	// Security
	{ID: "no-hardcoded-secrets", Category: "security", Description: "No API keys, passwords, tokens committed"},
	{ID: "input-validated", Category: "security", Description: "User input validated and sanitized"},
	{ID: "no-sql-injection", Category: "security", Description: "Queries use parameterized statements"},
	{ID: "no-command-injection", Category: "security", Description: "Shell commands don't interpolate unsanitized input"},
	{ID: "auth-checks-present", Category: "security", Description: "Operations requiring auth include access control"},
	{ID: "sensitive-data-not-logged", Category: "security", Description: "Passwords, tokens, PII not in logs"},
	{ID: "permissions-least-privilege", Category: "security", Description: "New permissions follow least privilege"},

	// Correctness
	{ID: "no-logic-errors", Category: "correctness", Description: "No off-by-one, incorrect conditions, or wrong operators"},
	{ID: "null-safety", Category: "correctness", Description: "Guards against nil/null dereferences"},
	{ID: "no-infinite-loops", Category: "correctness", Description: "Loops have correct termination conditions"},
	{ID: "boundary-conditions-handled", Category: "correctness", Description: "Guards against overflow, underflow, division by zero"},

	// Performance
	{ID: "no-n-plus-one-queries", Category: "performance", Description: "No query-per-item in loops"},
	{ID: "no-unbounded-growth", Category: "performance", Description: "Caches and buffers have size limits"},
	{ID: "pagination-used", Category: "performance", Description: "Large result sets use pagination or streaming"},
	{ID: "no-blocking-in-hot-path", Category: "performance", Description: "Critical paths don't block on heavy I/O"},
}

var AllCategories = []string{
	"completeness", "code-quality", "testing", "naming",
	"security", "correctness", "performance",
}

var RatingDimensions = []string{
	"duplication", "naming_consistency", "security", "test_coverage",
}

func EnabledChecks(cfg ChecksConfig) []Check {
	disabled := make(map[string]bool, len(cfg.Disabled))
	for _, id := range cfg.Disabled {
		disabled[id] = true
	}
	disabledCats := make(map[string]bool, len(cfg.DisabledCategories))
	for _, cat := range cfg.DisabledCategories {
		disabledCats[cat] = true
	}

	var out []Check
	for _, c := range AllChecks {
		if disabled[c.ID] || disabledCats[c.Category] {
			continue
		}
		out = append(out, c)
	}
	return out
}

func ChecksByCategory(checks []Check) map[string][]Check {
	out := make(map[string][]Check)
	for _, c := range checks {
		out[c.Category] = append(out[c.Category], c)
	}
	return out
}
