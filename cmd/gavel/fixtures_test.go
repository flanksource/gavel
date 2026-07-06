package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/flanksource/gavel/fixtures"
)

func TestFixturesHelpIncludesArgumentAndFlagReference(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"fixtures"})
	if err != nil {
		t.Fatalf("find fixtures command: %v", err)
	}

	help := fixturesHelp(cmd).ANSI()
	for _, want := range []string{
		"USAGE",
		"ARGUMENTS",
		"fixture-files",
		"daemon:",
		"{{.port}}",
		"ai:",
		"verify:",
		"claude-code-sonnet",
		"Focus on the new parser path",
		"Parser errors are actionable",
		"json.score >= 80",
		"maxTokens",
		"maxConcurrent",
		"cacheTTL",
		"noCache",
		"FORMAT 4: TEST / LINT STEPS",
		"```yaml test",
		"```lint",
		"test-timeout: 2m",
		"Common test keys",
		"Common lint keys",
		"gavel test --help",
		"gavel lint --help",
		"show-passed",
		"show-failed",
		"test, lint (or yaml test/yaml lint)",
		"expected count, count",
		"expected matches, matches",
		"query",
		"expectations",
		"ansi.has_cursor_hide",
		"ansi.final_text",
		"has_color(s)",
		"OUTPUT OPTIONS",
		"gavel fixtures outline tests.md",
		"Parse and outline without running fixtures",
		"--show-passed",
		"--show-stdout",
		"--show-stderr",
		"--update-golden",
		"--schema",
		"FLAGS",
		"--update-golden",
		"--show-passed",
		"--show-stdout",
		"--show-stderr",
		"--schema",
		"GLOBAL FLAGS",
		"--cwd",
		"-v, --loglevel",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("fixtures help missing %q:\n%s", want, help)
		}
	}
}

func TestFixturesSchemaDocumentIncludesRunnerSchemas(t *testing.T) {
	doc, err := fixturesSchemaDocument()
	if err != nil {
		t.Fatalf("fixturesSchemaDocument: %v", err)
	}
	if doc["source"] != "gavel fixtures --schema" {
		t.Fatalf("unexpected schema source: %#v", doc["source"])
	}
	help := doc["help"].(map[string]any)
	if help["command"] != "gavel fixtures --help" {
		t.Fatalf("unexpected help command: %#v", help["command"])
	}
	helpText := help["text"].(string)
	for _, want := range []string{"FORMAT 4: TEST / LINT STEPS", "CEL VALIDATION", "CWD RESOLUTION"} {
		if !strings.Contains(helpText, want) {
			t.Fatalf("schema help text missing %q:\n%s", want, helpText)
		}
	}
	helpSections := help["sections"].([]map[string]any)
	if len(helpSections) == 0 {
		t.Fatalf("schema help should include structured sections: %#v", help)
	}

	fences := doc["fences"].(map[string]any)
	testSchema := fences["test"].(map[string]any)["schema"].(fixtureJSONSchema)
	lintSchema := fences["lint"].(map[string]any)["schema"].(fixtureJSONSchema)

	for _, want := range []string{"paths", "framework", "test-timeout", "show-passed", "show-failed"} {
		if !slices.Contains(sortedSchemaPropertyNames(testSchema), want) {
			t.Fatalf("test schema missing %q: %#v", want, sortedSchemaPropertyNames(testSchema))
		}
	}
	for _, want := range []string{"files", "linters", "changed", "show-failed"} {
		if !slices.Contains(sortedSchemaPropertyNames(lintSchema), want) {
			t.Fatalf("lint schema missing %q: %#v", want, sortedSchemaPropertyNames(lintSchema))
		}
	}
	for _, unwanted := range []string{"lint", "lint-timeout", "sync-todos", "todos-dir", "todo-template"} {
		if slices.Contains(sortedSchemaPropertyNames(testSchema), unwanted) {
			t.Fatalf("test schema should hide %q: %#v", unwanted, sortedSchemaPropertyNames(testSchema))
		}
	}
	for _, unwanted := range []string{
		"ai-fix",
		"ai-fix-max-iterations",
		"allowed-tools",
		"api-key",
		"backend",
		"bare",
		"budget",
		"disallowed-tools",
		"edit",
		"effort",
		"fix",
		"group-by",
		"max-tokens",
		"max-turns",
		"model",
		"no-cache",
		"no-hooks",
		"no-mcp",
		"no-memory",
		"no-project",
		"no-skills",
		"no-user",
		"permission-mode",
		"resume",
		"skill-dir",
		"summary",
		"summary-limit",
		"sync-todos",
		"temperature",
		"triage",
		"yes",
	} {
		if slices.Contains(sortedSchemaPropertyNames(lintSchema), unwanted) {
			t.Fatalf("lint schema should hide %q: %#v", unwanted, sortedSchemaPropertyNames(lintSchema))
		}
	}

	testProps := testSchema["properties"].(map[string]any)
	framework := testProps["framework"].(fixtureSchemaProperty)
	if framework["x-array-display"] != "filter-pills" {
		t.Fatalf("framework should render as filter pills: %#v", framework)
	}
	assertSchemaHelp(t, framework, "Common test keys", "fixtures --help")
	frameworkItems := framework["items"].(fixtureSchemaProperty)
	if !slices.Contains(stringSliceFromAny(frameworkItems["enum"]), "go test") {
		t.Fatalf("framework enum missing go test: %#v", frameworkItems["enum"])
	}

	lintProps := lintSchema["properties"].(map[string]any)
	linters := lintProps["linters"].(fixtureSchemaProperty)
	if linters["x-array-display"] != "filter-pills" {
		t.Fatalf("linters should render as filter pills: %#v", linters)
	}
	assertSchemaHelp(t, linters, "Common lint keys", "fixtures --help")
	linterItems := linters["items"].(fixtureSchemaProperty)
	if !slices.Contains(stringSliceFromAny(linterItems["enum"]), "oxlint") {
		t.Fatalf("linter enum missing oxlint: %#v", linterItems["enum"])
	}

	execSchema := fences["exec"].(map[string]any)["schema"].(fixtureJSONSchema)
	for _, want := range []string{"content", "exitCode", "stdout", "stderr", "output", "error", "format", "count", "timeout", "cel", "properties"} {
		if !slices.Contains(sortedSchemaPropertyNames(execSchema), want) {
			t.Fatalf("exec schema missing %q: %#v", want, sortedSchemaPropertyNames(execSchema))
		}
	}
	execProps := execSchema["properties"].(map[string]any)
	assertSchemaHelp(t, execProps["cel"].(fixtureSchemaProperty), "CEL validation", "fixtures --help")

	frontmatterSchema := doc["frontmatter"].(fixtureJSONSchema)
	frontmatterProps := frontmatterSchema["properties"].(map[string]any)
	assertSchemaHelp(t, frontmatterProps["cwd"].(fixtureSchemaProperty), "CWD resolution", "fixtures --help")

	var buf bytes.Buffer
	if err := writeFixturesSchema(&buf); err != nil {
		t.Fatalf("writeFixturesSchema: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("schema output is not valid JSON: %v\n%s", err, buf.String())
	}
}

func assertSchemaHelp(t *testing.T, prop fixtureSchemaProperty, section, source string) {
	t.Helper()
	help, ok := prop["x-help"].(map[string]any)
	if !ok {
		t.Fatalf("schema property missing x-help: %#v", prop)
	}
	if help["section"] != section {
		t.Fatalf("x-help.section = %#v, want %q in %#v", help["section"], section, help)
	}
	if help["source"] != source {
		t.Fatalf("x-help.source = %#v, want %q in %#v", help["source"], source, help)
	}
}

func stringSliceFromAny(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, item.(string))
		}
		return out
	default:
		return nil
	}
}

func TestFixturesOutlineCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"fixtures", "outline"})
	if err != nil {
		t.Fatalf("find fixtures outline command: %v", err)
	}
	if cmd == nil || cmd.Name() != "outline" {
		t.Fatalf("expected fixtures outline command, got %#v", cmd)
	}
}

func TestRunFixturesOutlineReturnsReport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "outline.fixture.md")
	if err := os.WriteFile(path, []byte(`# Suite

| Name | Command | CEL |
|------|---------|-----|
| ok | echo ok | stdout.contains("ok") |
`), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := runFixturesOutline(fixturesOutlineOptions{Paths: []string{path}})
	if err != nil {
		t.Fatalf("runFixturesOutline: %v", err)
	}
	report, ok := got.(*fixtures.OutlineReport)
	if !ok {
		t.Fatalf("runFixturesOutline returned %T, want *fixtures.OutlineReport", got)
	}
	if report.Files != 1 || report.Fixtures != 1 || report.Counts["exec"] != 1 {
		t.Fatalf("unexpected outline report: %+v", report)
	}
}
