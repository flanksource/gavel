package parsers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlaywrightJSON_Parse(t *testing.T) {
	data, err := os.ReadFile("testdata/playwright-sample.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	parser := NewPlaywrightJSON("/repo")
	tests, err := parser.Parse(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(tests) != 3 {
		t.Fatalf("expected 3 tests, got %d: %+v", len(tests), tests)
	}

	byName := map[string]Test{}
	for _, tst := range tests {
		byName[tst.Name] = tst
	}

	pass, ok := byName["renders title"]
	if !ok {
		t.Fatal("missing 'renders title'")
	}
	if !pass.Passed {
		t.Errorf("expected pass, got %+v", pass)
	}
	if pass.Framework != Playwright {
		t.Errorf("framework = %q", pass.Framework)
	}
	wantFile := filepath.FromSlash("e2e/home.spec.ts")
	if pass.File != wantFile {
		t.Errorf("file = %q, want %q", pass.File, wantFile)
	}
	wantSuite := []string{"home.spec.ts", "homepage"}
	if strings.Join(pass.Suite, ">") != strings.Join(wantSuite, ">") {
		t.Errorf("suite = %v, want %v", pass.Suite, wantSuite)
	}

	fail, ok := byName["login works"]
	if !ok {
		t.Fatal("missing 'login works'")
	}
	if !fail.Failed {
		t.Errorf("expected failed=true, got %+v", fail)
	}
	if !strings.Contains(fail.Message, "Timed out waiting") {
		t.Errorf("message missing expected text: %q", fail.Message)
	}
	if strings.Contains(fail.Message, "\x1b[") {
		t.Errorf("ANSI leaked: %q", fail.Message)
	}
	if fail.Duration == 0 {
		t.Errorf("expected non-zero duration")
	}

	skip, ok := byName["admin only"]
	if !ok {
		t.Fatal("missing 'admin only'")
	}
	if !skip.Skipped {
		t.Errorf("expected skipped=true, got %+v", skip)
	}
}
