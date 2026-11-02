package testui

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

func TestParseRouteRequestDefaultsLintGroupingToLinterRuleFile(t *testing.T) {
	req := httptest.NewRequest("GET", "/lint", nil)
	parsed, ok := parseRouteRequest(req)
	if !ok {
		t.Fatalf("parseRouteRequest returned ok=false")
	}
	if parsed.LintFilters.Grouping != "linter-rule-file" {
		t.Fatalf("grouping = %q, want linter-rule-file", parsed.LintFilters.Grouping)
	}
}

// The history tab is a page route (reloading it must serve the page) with no
// export rendering.
func TestHistoryTabIsPageOnly(t *testing.T) {
	handler := NewServer().Handler()
	for path, want := range map[string]int{"/history": 200, "/history.json": 404} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Accept", "text/html")
		handler.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("GET %s = %d, want %d", path, rec.Code, want)
		}
	}
}

// The files tab is a page route for any file path, including one whose
// extension reads as an export format, and has no export rendering.
func TestFilesTabIsPageOnly(t *testing.T) {
	handler := NewServer().Handler()
	for path, want := range map[string]int{
		"/files": 200,
		"/files/rules/br/CopyBook-Fees/XMLData.xml": 200,
		"/files/docs/README.md":                     200,
		"/files/testdata/report.json":               200,
		"/files.json":                               404,
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path+"?dir=/work", nil)
		req.Header.Set("Accept", "text/html")
		handler.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("GET %s = %d, want %d", path, rec.Code, want)
		}
	}
}

// TestParseTestFiltersMatchesTheUIStatusParam mirrors src/routes.ts: no
// status param is the failing-only default, `all` is no status filter.
func TestParseTestFiltersMatchesTheUIStatusParam(t *testing.T) {
	for _, tc := range []struct {
		query string
		want  []string
	}{
		{query: "", want: []string{"!passed", "!skipped"}},
		{query: "?status=all", want: nil},
		{query: "?status=failed,!skipped", want: []string{"failed", "!skipped"}},
	} {
		got := parseTestFilters(httptest.NewRequest("GET", "/tests.json"+tc.query, nil)).Status
		if strings.Join(got, ",") != strings.Join(tc.want, ",") || (got == nil) != (tc.want == nil) {
			t.Errorf("status for %q = %#v, want %#v", tc.query, got, tc.want)
		}
	}
}

func TestBuildLintByLinterRuleFileGroupsRulesBeforeFolders(t *testing.T) {
	results := []*linters.LinterResult{{
		Linter:  "golangci-lint",
		WorkDir: "/tmp/work",
		Success: false,
		Violations: []models.Violation{
			testViolation("pkg/foo.go", 4, "errcheck"),
			testViolation("pkg/foo.go", 9, "errcheck"),
			testViolation("pkg/sub/bar.go", 3, "unused"),
		},
	}}

	nodes := buildLintByLinterRuleFile(results, lintRouteFilters{Grouping: "linter-rule-file"})
	if len(nodes) != 1 {
		t.Fatalf("top-level nodes = %d, want 1", len(nodes))
	}

	linter := nodes[0]
	if linter.Kind != "linter" {
		t.Fatalf("linter kind = %q, want linter", linter.Kind)
	}
	if got := names(linter.Children); !containsName(got, "errcheck (2)") || !containsName(got, "unused (1)") {
		t.Fatalf("rule children = %v, want errcheck and unused groups", got)
	}

	errcheck := mustChildByName(t, linter.Children, "errcheck (2)")
	if errcheck.Kind != "lint-rule-group" {
		t.Fatalf("rule group kind = %q, want lint-rule-group", errcheck.Kind)
	}
	pkg := mustChildByName(t, errcheck.Children, "pkg (2)")
	if pkg.Kind != "lint-folder" {
		t.Fatalf("pkg kind = %q, want lint-folder", pkg.Kind)
	}
	foo := mustChildByName(t, pkg.Children, "foo.go (2)")
	if foo.Kind != "lint-file" || foo.File != "pkg/foo.go" {
		t.Fatalf("foo node = %#v", foo)
	}
}

// One fix at three places across two files is one violation: counted once
// under its rule, listing each place (relative to the work dir) beneath it.
func TestBuildLintByLinterRuleFileListsEveryLocationUnderOneViolation(t *testing.T) {
	v := testViolation("rules/a/XMLData.xml", 4, "plan-warning")
	v.Locations = []models.ViolationLocation{
		{File: "/tmp/work/rules/a/XMLData.xml", Line: 4, Message: "statement a1"},
		{File: "rules/a/XMLData.xml", Line: 12, Message: "statement a2"},
		{File: "rules/b/XMLData.xml", Line: 9, Message: "statement b1"},
	}
	results := []*linters.LinterResult{{Linter: "oipa-transaction", WorkDir: "/tmp/work", Violations: []models.Violation{v}}}

	linter := buildLintByLinterRuleFile(results, lintRouteFilters{Grouping: "linter-rule-file"})[0]
	rule := mustChildByName(t, linter.Children, "plan-warning (1)")
	node := rule
	for _, name := range []string{"rules (1)", "a (1)", "XMLData.xml (1)", "rules/a/XMLData.xml:4 violation (3 locations)"} {
		node = mustChildByName(t, node.Children, name)
	}
	file := node
	want := []string{"rules/a/XMLData.xml:4 statement a1", "rules/a/XMLData.xml:12 statement a2", "rules/b/XMLData.xml:9 statement b1"}
	if got := names(file.Children); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("location children = %v, want %v", got, want)
	}
	if file.Children[2].Kind != "lint-location" || file.Children[2].File != "rules/b/XMLData.xml" || file.Children[2].Line != 9 {
		t.Fatalf("location node = %#v, want a lint-location on rules/b/XMLData.xml:9", file.Children[2])
	}
}

func TestBuildLintByFileLinterRuleCollapsesCheckoutPrefix(t *testing.T) {
	results := []*linters.LinterResult{{
		Linter:  "betterleaks",
		WorkDir: "/tmp/work",
		Success: false,
		Violations: []models.Violation{
			testViolation(".shell/checkout/abc123/README.md", 4, "readme-rule"),
			testViolation(".shell/checkout/abc123/chart/app/conf/cycle/cycle.ps1", 8, "cycle-rule"),
			testViolation(".shell/checkout/abc123/chart/app/ci/fargate-values.yaml", 12, "yaml-rule"),
		},
	}}

	nodes := buildLintByFileLinterRule(results, lintRouteFilters{Grouping: "file-linter-rule"})
	if len(nodes) != 1 {
		t.Fatalf("top-level nodes = %d, want 1", len(nodes))
	}

	checkout := nodes[0]
	if checkout.Kind != "lint-folder" {
		t.Fatalf("top-level kind = %q, want lint-folder", checkout.Kind)
	}
	if checkout.Name != ".shell/checkout/abc123 (3)" {
		t.Fatalf("top-level name = %q", checkout.Name)
	}

	if got := names(checkout.Children); !containsName(got, "chart (2)") || !containsName(got, "README.md (1)") {
		t.Fatalf("checkout children = %v, want chart folder and README file", got)
	}

	chart := mustChildByName(t, checkout.Children, "chart (2)")
	app := mustChildByName(t, chart.Children, "app (2)")
	conf := mustChildByName(t, app.Children, "conf (1)")
	cycle := mustChildByName(t, conf.Children, "cycle (1)")
	ps1 := mustChildByName(t, cycle.Children, "cycle.ps1 (1)")
	if ps1.Kind != "lint-file" || ps1.File != ".shell/checkout/abc123/chart/app/conf/cycle/cycle.ps1" {
		t.Fatalf("cycle file node = %#v", ps1)
	}
}

func TestBuildLintByFileLinterRulePromotesFilelessViolations(t *testing.T) {
	results := []*linters.LinterResult{{
		Linter:  "betterleaks",
		WorkDir: "/tmp/work",
		Success: false,
		Violations: []models.Violation{
			testViolation(".shell/checkout/abc123/README.md", 4, "readme-rule"),
			testViolation("", 0, "global-rule"),
		},
	}}

	nodes := buildLintByFileLinterRule(results, lintRouteFilters{Grouping: "file-linter-rule"})
	if len(nodes) != 2 {
		t.Fatalf("top-level nodes = %d, want 2", len(nodes))
	}

	for _, node := range nodes {
		if strings.Contains(node.Name, "(no file)") || node.File == "(no file)" {
			t.Fatalf("unexpected no-file node: %#v", node)
		}
	}

	linter := mustChildByName(t, nodes, "betterleaks (1)")
	if linter.Kind != "linter" {
		t.Fatalf("promoted node kind = %q, want linter", linter.Kind)
	}
	rule := mustChildByName(t, linter.Children, "global-rule (1)")
	if rule.Kind != "lint-rule" || rule.File != "" {
		t.Fatalf("promoted rule node = %#v", rule)
	}
}

func TestBuildLintByLinterFileOmitsNoFileNode(t *testing.T) {
	results := []*linters.LinterResult{{
		Linter:  "betterleaks",
		WorkDir: "/tmp/work",
		Success: false,
		Violations: []models.Violation{
			testViolation(".shell/checkout/abc123/README.md", 4, "readme-rule"),
			testViolation("", 0, "global-rule"),
		},
	}}

	nodes := buildLintByLinterFile(results, lintRouteFilters{Grouping: "linter-file"})
	if len(nodes) != 1 {
		t.Fatalf("top-level nodes = %d, want 1", len(nodes))
	}

	linter := nodes[0]
	if linter.Name != "betterleaks (2)" {
		t.Fatalf("linter name = %q, want betterleaks (2)", linter.Name)
	}
	if len(linter.Children) != 1 {
		t.Fatalf("linter children = %d, want 1 visible file child", len(linter.Children))
	}
	if linter.Children[0].Name != ".shell/checkout/abc123/README.md (1)" {
		t.Fatalf("file child name = %q, want .shell/checkout/abc123/README.md (1)", linter.Children[0].Name)
	}
	for _, node := range linter.Children {
		if strings.Contains(node.Name, "(no file)") || node.File == "(no file)" {
			t.Fatalf("unexpected no-file child: %#v", node)
		}
	}
}

func testViolation(file string, line int, rule string) models.Violation {
	msg := "violation"
	return models.Violation{
		File:     file,
		Line:     line,
		Message:  &msg,
		Severity: models.SeverityError,
		Rule:     &models.Rule{Method: rule},
	}
}

func mustChildByName(t *testing.T, nodes []*ViewNode, want string) *ViewNode {
	t.Helper()
	for _, node := range nodes {
		if node.Name == want {
			return node
		}
	}
	t.Fatalf("node %q not found in %v", want, names(nodes))
	return nil
}

func names(nodes []*ViewNode) []string {
	out := make([]string, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, node.Name)
	}
	return out
}

func containsName(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
