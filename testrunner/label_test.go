package testrunner

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestStripANSI(t *testing.T) {
	cases := map[string]string{
		"plain text":         "plain text",
		"\x1b[31mred\x1b[0m": "red",
		"\x1b[1;38;2;107;113;128mbold gray\x1b[0m": "bold gray",
		"prefix \x1b[32mmiddle\x1b[0m suffix":      "prefix middle suffix",
		"\x1b[Kedge-case escape":                   "edge-case escape",
		"already-\x1b[0mwithout-colour":            "already-without-colour",
	}
	for in, want := range cases {
		if got := stripANSI(in); got != want {
			t.Errorf("stripANSI(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCompactCounts(t *testing.T) {
	sum := parsers.TestSummary{Passed: 5, Total: 5, Duration: 1230 * time.Millisecond}
	got := compactCounts(sum)
	// Must be free of ANSI and short.
	if strings.Contains(got, "\x1b") {
		t.Errorf("compactCounts contained ANSI: %q", got)
	}
	if !strings.Contains(got, "5 passed") || !strings.Contains(got, "5 total") {
		t.Errorf("compactCounts(%+v) = %q, missing counts", sum, got)
	}
}

func TestFormatRunningCommand(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
		args []string
		want string
	}{
		{name: "command and args", cmd: "go", args: []string{"test", "-json", "./pkg/foo"}, want: "go test -json ./pkg/foo"},
		{name: "command only", cmd: "golangci-lint", want: "golangci-lint"},
		{name: "args only", args: []string{"eslint", "--format=json", "."}, want: "eslint --format=json ."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatRunningCommand(tc.cmd, tc.args); got != tc.want {
				t.Errorf("formatRunningCommand(%q, %v) = %q, want %q", tc.cmd, tc.args, got, tc.want)
			}
		})
	}
}

func TestFirstNonBlankLine(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"only blank", "\n\n  \t\n", ""},
		{"strips ansi", "\x1b[31merror: boom\x1b[0m", "error: boom"},
		{"first real line", "\n  \nreal line\nsecond", "real line"},
		{"trims surrounding ws", "   padded line   \nmore", "padded line"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstNonBlankLine(tc.in); got != tc.want {
				t.Errorf("firstNonBlankLine(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFailureSnippet(t *testing.T) {
	cases := []struct {
		name string
		test parsers.Test
		want string
	}{
		{
			name: "stderr wins",
			test: parsers.Test{
				Stderr:  "Load key '/tmp/ginkgo/id_ed25519': invalid format\nPermission denied",
				Stdout:  "extra",
				Message: "ignored",
			},
			want: "Load key '/tmp/ginkgo/id_ed25519': invalid format",
		},
		{
			name: "stdout when no stderr",
			test: parsers.Test{
				Stdout:  "the output\nmore",
				Message: "ignored",
			},
			want: "the output",
		},
		{
			name: "message when no stderr/stdout",
			test: parsers.Test{Message: "simple message"},
			want: "simple message",
		},
		{
			name: "falls back to name",
			test: parsers.Test{Name: "TestExample"},
			want: "TestExample",
		},
		{
			name: "defaults to 'no failure details'",
			test: parsers.Test{},
			want: "no failure details",
		},
		{
			name: "strips ANSI from stderr",
			test: parsers.Test{
				Stderr: "\x1b[31m[FAILED] git push failed\x1b[0m",
			},
			want: "[FAILED] git push failed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := failureSnippet(tc.test); got != tc.want {
				t.Errorf("failureSnippet(%+v) = %q, want %q", tc.test, got, tc.want)
			}
		})
	}
}

func TestJoinLabel(t *testing.T) {
	cases := []struct {
		name    string
		prefix  string
		details string
		budget  int
		want    string
	}{
		{
			name:    "fits comfortably",
			prefix:  "gotest ./verify",
			details: "5 passed 5 total 1.23s",
			budget:  68,
			want:    "gotest ./verify  5 passed 5 total 1.23s",
		},
		{
			name:    "truncates details with ellipsis",
			prefix:  "gotest ./serve",
			details: strings.Repeat("x", 200),
			budget:  40,
			want:    "gotest ./serve  " + strings.Repeat("x", 40-len("gotest ./serve  ")-1) + "…",
		},
		{
			name:    "empty details returns prefix",
			prefix:  "golangci-lint",
			details: "",
			budget:  40,
			want:    "golangci-lint",
		},
		{
			name:    "strips ANSI before width calc",
			prefix:  "\x1b[1mgotest\x1b[0m ./pkg",
			details: "\x1b[31mfail\x1b[0m",
			budget:  40,
			want:    "gotest ./pkg  fail",
		},
		{
			name:    "prefix too long truncates prefix",
			prefix:  strings.Repeat("x", 80),
			details: "hi",
			budget:  10,
			want:    strings.Repeat("x", 9) + "…",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := joinLabel(tc.prefix, tc.details, tc.budget)
			if got != tc.want {
				t.Errorf("joinLabel(%q, %q, %d) = %q, want %q",
					tc.prefix, tc.details, tc.budget, got, tc.want)
			}
			if runeLen(got) > tc.budget {
				t.Errorf("joinLabel output exceeded budget %d: %q (%d)",
					tc.budget, got, runeLen(got))
			}
		})
	}
}

func TestFormatPackageLabelPass(t *testing.T) {
	sum := parsers.TestSummary{Passed: 5, Total: 5, Duration: 1230 * time.Millisecond}
	got := formatPackageLabel(parsers.GoTest, "./pkg/verify", sum, nil)
	if !strings.Contains(got, "./pkg/verify") {
		t.Errorf("missing package path: %q", got)
	}
	if !strings.Contains(got, "5 passed") {
		t.Errorf("missing pass count: %q", got)
	}
	if runeLen(got) > maxTaskLabel {
		t.Errorf("label exceeded budget: %q (%d > %d)", got, runeLen(got), maxTaskLabel)
	}
	if strings.Contains(got, "\x1b") {
		t.Errorf("label contained ANSI: %q", got)
	}
}

func TestFormatPackageLabelFailWithStderr(t *testing.T) {
	sum := parsers.TestSummary{Passed: 5, Failed: 1, Total: 6, Duration: 500 * time.Millisecond}
	fail := &parsers.Test{
		Name:   "TestBoom",
		Failed: true,
		Stderr: "Load key '/tmp/ginkgo/id_ed25519': invalid format\nmore",
	}
	got := formatPackageLabel(parsers.GoTest, "./serve", sum, fail)
	if !strings.Contains(got, "./serve") {
		t.Errorf("missing package path: %q", got)
	}
	// First stderr line should be present (possibly truncated).
	if !strings.Contains(got, "Load key") {
		t.Errorf("missing stderr snippet: %q", got)
	}
	if runeLen(got) > maxTaskLabel {
		t.Errorf("label exceeded budget: %q (%d > %d)", got, runeLen(got), maxTaskLabel)
	}
	// Pass-counts should NOT appear on a failure label.
	if strings.Contains(got, "passed") {
		t.Errorf("failure label leaked pass counts: %q", got)
	}
}

func TestFormatPackageLabelFailNoDetails(t *testing.T) {
	sum := parsers.TestSummary{Failed: 1, Total: 1}
	fail := &parsers.Test{Failed: true}
	got := formatPackageLabel(parsers.Ginkgo, "./testrunner/runners", sum, fail)
	if !strings.Contains(got, "no failure details") {
		t.Errorf("expected 'no failure details' fallback, got: %q", got)
	}
}

// Linter label assembly lives in the linters package (buildLinterLabel in
// linters/runner.go); its coverage is in linters/runner_compact_test.go.

func TestFirstFailingLeaf(t *testing.T) {
	// Build a tree: package → suite → [passing leaf, failing leaf]
	results := parsers.TestSuiteResults{{
		Tests: parsers.Tests{
			{
				Name: "suite",
				Children: parsers.Tests{
					{Name: "PassingChild", Passed: true},
					{Name: "FailingChild", Failed: true, Message: "boom"},
				},
			},
		},
	}}
	got := firstFailingLeaf(results)
	if got == nil {
		t.Fatal("expected failing leaf, got nil")
	}
	if got.Name != "FailingChild" {
		t.Errorf("got %q, want FailingChild", got.Name)
	}

	// No failures at all → nil.
	cleanResults := parsers.TestSuiteResults{{
		Tests: parsers.Tests{
			{Name: "OK", Passed: true},
		},
	}}
	if got := firstFailingLeaf(cleanResults); got != nil {
		t.Errorf("expected nil for clean results, got %+v", got)
	}
}

func TestShortFrameworkName(t *testing.T) {
	cases := map[parsers.Framework]string{
		parsers.GoTest: "gotest",
		parsers.Ginkgo: "ginkgo",
	}
	for fw, want := range cases {
		if got := shortFrameworkName(fw); got != want {
			t.Errorf("shortFrameworkName(%v) = %q, want %q", fw, got, want)
		}
	}
}
