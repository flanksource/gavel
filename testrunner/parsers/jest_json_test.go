package parsers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
