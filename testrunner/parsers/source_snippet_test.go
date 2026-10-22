package parsers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const snippetFixture = `kind: Fixture
steps:
  - id: build
    kind: exec
  - id: render
    kind: exec
    exec: ./render
`

func writeSnippetFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "render.fixture.yaml")
	if err := os.WriteFile(path, []byte(snippetFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPretty_FailedTestShowsItsSourceLines(t *testing.T) {
	path := writeSnippetFixture(t)
	rendered := Test{Name: "render", Failed: true, File: path, Line: 5, EndLine: 7}.Pretty().String()
	for _, want := range []string{"- id: render", "kind: exec", "exec: ./render"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered failure should contain source line %q, got:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "id: build") {
		t.Errorf("rendered failure should hold only lines 5-7, got:\n%s", rendered)
	}
}

func TestPretty_ResolvesSourceAgainstWorkDir(t *testing.T) {
	path := writeSnippetFixture(t)
	rendered := Test{Name: "render", Failed: true, WorkDir: filepath.Dir(path), File: filepath.Base(path), Line: 5, EndLine: 5}.Pretty().String()
	if !strings.Contains(rendered, "- id: render") {
		t.Errorf("relative source should resolve against WorkDir, got:\n%s", rendered)
	}
}

func TestPretty_PassedTestShowsNoSource(t *testing.T) {
	path := writeSnippetFixture(t)
	rendered := Test{Name: "render", Passed: true, File: path, Line: 5, EndLine: 7}.Pretty().String()
	if strings.Contains(rendered, "exec: ./render") {
		t.Errorf("a passing test should not render its source, got:\n%s", rendered)
	}
}

func TestPretty_UnreadableSourceSaysWhy(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "gone.fixture.yaml")
	rendered := Test{Name: "render", Failed: true, File: missing, Line: 5, EndLine: 7}.Pretty().String()
	if !strings.Contains(rendered, "source unavailable") || !strings.Contains(rendered, "gone.fixture.yaml") {
		t.Errorf("an unreadable source should be reported, got:\n%s", rendered)
	}
}

func TestPretty_SourcePastEndOfFileSaysWhy(t *testing.T) {
	path := writeSnippetFixture(t)
	rendered := Test{Name: "render", Failed: true, File: path, Line: 5, EndLine: 40}.Pretty().String()
	if !strings.Contains(rendered, "source unavailable") || !strings.Contains(rendered, "lines 5-40") {
		t.Errorf("a range past the end of the file should be reported, got:\n%s", rendered)
	}
}
