package parsers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractStackLocation(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantLn  int
	}{
		{
			name:   "vitest caret frame",
			in:     "AssertionError: boom\n ❯ src/retry.test.ts:42:5",
			want:   "src/retry.test.ts",
			wantLn: 42,
		},
		{
			name:   "skip node_modules",
			in:     "err\n    at /repo/node_modules/vitest/chunk.js:10:5\n    at /repo/src/retry.test.ts:42:5",
			want:   "/repo/src/retry.test.ts",
			wantLn: 42,
		},
		{
			name:   "skip node: internals",
			in:     "err\n    at node:internal/timers:42:17\n    at src/app.test.ts:7:9",
			want:   "src/app.test.ts",
			wantLn: 7,
		},
		{
			name:   "no frame",
			in:     "boring error with no location",
			want:   "",
			wantLn: 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotFile, gotLn := extractStackLocation(tc.in)
			if gotFile != tc.want || gotLn != tc.wantLn {
				t.Errorf("got (%q, %d), want (%q, %d)", gotFile, gotLn, tc.want, tc.wantLn)
			}
		})
	}
}

func TestJestJSON_VitestRichContext(t *testing.T) {
	data, err := os.ReadFile("testdata/vitest-rich.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	parser := NewJestJSON("/repo", Vitest)
	tests, err := parser.Parse(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	byName := map[string]Test{}
	for _, tt := range tests {
		byName[tt.Name] = tt
	}

	// File-level setup error emits a synthetic failing record carrying
	// the Message, raw ANSI Stderr, and captured console entries as Stdout.
	setup, ok := byName["broken.test.ts"]
	if !ok {
		t.Fatalf("expected synthetic test for broken.test.ts, got %+v", tests)
	}
	if !setup.Failed {
		t.Error("setup-error record should be Failed=true")
	}
	if !strings.Contains(setup.Message, "SyntaxError: Unexpected token") {
		t.Errorf("message should carry syntax error, got %q", setup.Message)
	}
	if !strings.Contains(setup.Stdout, "setup tick") {
		t.Errorf("stdout should contain console entries, got %q", setup.Stdout)
	}

	// Retry assertion: Line extracted from user-code stack frame (42)
	// rather than assertion's own Location (40). node_modules frame is
	// skipped. Retries count is prefixed onto the Message. Console is
	// attached as Stdout. ANSI preserved in Stderr, stripped in Message.
	retry, ok := byName["flaky case"]
	if !ok {
		t.Fatalf("expected flaky-case test, got %+v", tests)
	}
	if retry.Line != 42 {
		t.Errorf("Line should come from user-code frame (42), got %d", retry.Line)
	}
	if !strings.Contains(retry.Message, "after 2 retries") {
		t.Errorf("Message should carry retry count, got %q", retry.Message)
	}
	if strings.Contains(retry.Message, "\x1b[") {
		t.Errorf("Message must be ANSI-stripped, got %q", retry.Message)
	}
	if !strings.Contains(retry.Stderr, "\x1b[") {
		t.Errorf("Stderr should retain ANSI codes for AnsiHtml render, got %q", retry.Stderr)
	}
	if !strings.Contains(retry.Stdout, "retry attempt 1") || !strings.Contains(retry.Stdout, "[ERROR]") {
		t.Errorf("Stdout should render console entries with type prefix, got %q", retry.Stdout)
	}
}

// TestJestJSON_VitestRealReport parses a real Vitest JSON report captured
// from `gavel test` against testrunner/ui. The report exercises the common
// "file-level setup failure with passing sibling suite" shape: one test
// file fails to import (`window is not defined`) so it has no assertion
// results, while another file emits 24 passing assertions across three
// describe blocks. The fixture is copied verbatim from
// .vitest/vitest-report-testrunner-ui-*.json.
func TestJestJSON_VitestRealReport(t *testing.T) {
	data, err := os.ReadFile("testdata/vitest-testrunner-ui-real.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	// The absolute paths in the fixture start at the gavel repo root.
	parser := NewJestJSON("/Users/moshe/go/src/github.com/flanksource/gavel", Vitest)
	tests, err := parser.Parse(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// 1 synthetic failing record (routes.test.ts setup failure) + 25
	// passing assertions from utils.test.ts (16 formatCount + 7
	// collapseSingleChildChains + 2 groupLintByLinterRuleFile) = 26 tests.
	if len(tests) != 26 {
		t.Fatalf("expected 26 tests (1 setup-failure + 25 passing), got %d", len(tests))
	}

	var setupFail *Test
	passedByName := map[string]*Test{}
	suites := map[string]int{}
	for i := range tests {
		tt := &tests[i]
		switch {
		case tt.Failed:
			setupFail = tt
		case tt.Passed:
			passedByName[tt.Name] = tt
			if len(tt.Suite) > 0 {
				suites[tt.Suite[0]]++
			}
		}
		if tt.Framework != Vitest {
			t.Errorf("test %q framework = %q, want vitest", tt.Name, tt.Framework)
		}
	}

	// File-level setup failure: emitted as a synthetic Test named after
	// the file's basename, carrying the setup message and the relative
	// path. No assertion means no Line.
	if setupFail == nil {
		t.Fatal("expected a synthetic failing test for the file-level setup error")
	}
	if setupFail.Name != "routes.test.ts" {
		t.Errorf("setup-fail name = %q, want routes.test.ts", setupFail.Name)
	}
	if setupFail.File != "testrunner/ui/src/routes.test.ts" {
		t.Errorf("setup-fail file = %q, want testrunner/ui/src/routes.test.ts", setupFail.File)
	}
	if !strings.Contains(setupFail.Message, "window is not defined") {
		t.Errorf("setup-fail message should contain 'window is not defined', got %q", setupFail.Message)
	}
	if setupFail.Line != 0 {
		t.Errorf("setup-fail Line should be 0 (no stack in message), got %d", setupFail.Line)
	}

	// Passing assertions: every real spec should be Passed with a duration.
	// The fixture has three top-level describe blocks; assertion counts
	// match the real run.
	wantSuites := map[string]int{
		"formatCount":              16,
		"collapseSingleChildChains": 7,
		"groupLintByLinterRuleFile": 2,
	}
	for suite, want := range wantSuites {
		if got := suites[suite]; got != want {
			t.Errorf("suite %q: got %d passing, want %d", suite, got, want)
		}
	}

	// Spot-check a specific assertion so we know Suite, Name, File, and
	// sub-millisecond Duration all survive the decode.
	spot := passedByName["formats 1234567 as 1.2M"]
	if spot == nil {
		t.Fatalf("missing passing test 'formats 1234567 as 1.2M'; have: %v", sortedKeys(passedByName))
	}
	if len(spot.Suite) != 1 || spot.Suite[0] != "formatCount" {
		t.Errorf("suite = %v, want [formatCount]", spot.Suite)
	}
	if spot.File != "testrunner/ui/src/utils.test.ts" {
		t.Errorf("file = %q, want testrunner/ui/src/utils.test.ts", spot.File)
	}
	if spot.Duration <= 0 {
		t.Errorf("duration must be positive (sub-ms), got %v", spot.Duration)
	}
	if spot.Stdout != "" || spot.Stderr != "" || spot.Message != "" {
		t.Errorf("passing test should have no stdout/stderr/message, got %+v", spot)
	}
}

func sortedKeys(m map[string]*Test) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestJestJSON_Parse(t *testing.T) {
	cases := []struct {
		name       string
		fixture    string
		framework  Framework
		file       string
		passTitle  string
		failTitle  string
		failSubstr string
		skipTitle  string
	}{
		{
			name:       "jest",
			fixture:    "testdata/jest-sample.json",
			framework:  Jest,
			file:       "src/math.test.js",
			passTitle:  "adds 1 + 2",
			failTitle:  "fails for string input",
			failSubstr: "expected 3 but got NaN",
			skipTitle:  "division not implemented",
		},
		{
			name:       "vitest",
			fixture:    "testdata/vitest-sample.json",
			framework:  Vitest,
			file:       "src/sum.test.ts",
			passTitle:  "adds numbers",
			failTitle:  "fails on NaN",
			failSubstr: "expected NaN to be 3",
			skipTitle:  "handles negatives",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			parser := NewJestJSON("/repo", tc.framework)
			tests, err := parser.Parse(strings.NewReader(string(data)))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if len(tests) != 3 {
				t.Fatalf("expected 3 tests, got %d", len(tests))
			}

			byName := map[string]Test{}
			for _, tst := range tests {
				byName[tst.Name] = tst
			}

			pass, ok := byName[tc.passTitle]
			if !ok {
				t.Fatalf("missing pass test %q", tc.passTitle)
			}
			if !pass.Passed || pass.Failed || pass.Skipped {
				t.Errorf("expected pass, got %+v", pass)
			}
			if pass.Framework != tc.framework {
				t.Errorf("framework = %q, want %q", pass.Framework, tc.framework)
			}
			if pass.File != filepath.FromSlash(tc.file) {
				t.Errorf("file = %q, want %q", pass.File, tc.file)
			}

			fail, ok := byName[tc.failTitle]
			if !ok {
				t.Fatalf("missing fail test %q", tc.failTitle)
			}
			if !fail.Failed {
				t.Errorf("expected failed=true, got %+v", fail)
			}
			if !strings.Contains(fail.Message, tc.failSubstr) {
				t.Errorf("message missing %q: %q", tc.failSubstr, fail.Message)
			}
			if strings.Contains(fail.Message, "\x1b[") {
				t.Errorf("ANSI codes leaked into message: %q", fail.Message)
			}
			if fail.Line == 0 {
				t.Errorf("expected non-zero Line")
			}

			skip, ok := byName[tc.skipTitle]
			if !ok {
				t.Fatalf("missing skipped test %q", tc.skipTitle)
			}
			if !skip.Skipped {
				t.Errorf("expected skipped=true, got %+v", skip)
			}
		})
	}
}
