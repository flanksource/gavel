package testrunner

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// maxTaskLabel is the width budget for per-task lines rendered to stderr
// in CI. Clicky adds a status icon prefix (~2ch) and a duration column
// (~10ch) on top, so the visible total lands around 80 characters.
const maxTaskLabel = 68

// ansiRe matches VT100/xterm SGR escapes. Used to strip colors from
// test framework output before measuring widths or truncating.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// stripANSI removes every ANSI escape from s. Safe to call on plain text.
func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// renderText returns a ANSI-coloured string when the caller has not passed
// --no-color, and a plain string otherwise. Use this instead of a raw
// `.ANSI()` call on any user-facing task label so `--no-color` actually
// takes effect.
func renderText(t api.Text) string {
	if clicky.Flags.NoColor {
		return t.String()
	}
	return t.ANSI()
}

// formatPackageLabel builds the compact task name for a finished test
// package. The name fits in `maxTaskLabel` characters; clicky will add
// its own status icon + duration when it renders, landing the final line
// at ≤ 80 chars.
//
// On pass: `gotest ./pkg/verify  5 passed 5 total 1.23s`
// On fail: `gotest ./pkg/serve  first-failure-snippet…`
func formatPackageLabel(framework parsers.Framework, pkgPath string, sum parsers.TestSummary, firstFail *parsers.Test) string {
	fw := shortFrameworkName(framework)
	prefix := fmt.Sprintf("%s %s", fw, pkgPath)
	var details string
	if firstFail != nil && sum.Failed > 0 {
		details = failureSnippet(*firstFail)
	} else {
		details = compactCounts(sum)
	}
	return joinLabel(prefix, details, maxTaskLabel)
}

// joinLabel joins the prefix and details with padding, right-truncating
// details (with an ellipsis) so the total length is ≤ budget. If the
// prefix alone exceeds the budget, it is right-truncated too.
func joinLabel(prefix, details string, budget int) string {
	prefix = stripANSI(prefix)
	details = stripANSI(strings.TrimSpace(details))
	if runeLen(prefix) > budget {
		return truncateRunes(prefix, budget)
	}
	if details == "" {
		return prefix
	}
	// Two spaces between prefix and details for readability.
	const sep = "  "
	remaining := budget - runeLen(prefix) - len(sep)
	if remaining <= 1 {
		return prefix
	}
	return prefix + sep + truncateRunes(details, remaining)
}

// compactCounts returns a short human-readable count line for a passing
// or warning package.
func compactCounts(sum parsers.TestSummary) string {
	parts := make([]string, 0, 4)
	if sum.Passed > 0 {
		parts = append(parts, fmt.Sprintf("%d passed", sum.Passed))
	}
	if sum.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", sum.Failed))
	}
	if sum.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", sum.Skipped))
	}
	if sum.Total > 0 {
		parts = append(parts, fmt.Sprintf("%d total", sum.Total))
	}
	if sum.Duration > 0 {
		parts = append(parts, sum.Duration.Round(10_000_000).String()) // 10ms resolution
	}
	return strings.Join(parts, " ")
}

// failureSnippet pulls the best available failure context from a test:
// stderr → stdout → message, then the first non-blank line, stripped
// of ANSI and trimmed of surrounding whitespace.
func failureSnippet(t parsers.Test) string {
	for _, body := range []string{t.Stderr, t.Stdout, t.Message} {
		line := firstNonBlankLine(body)
		if line != "" {
			return line
		}
	}
	if len(t.Suite) > 0 || t.Name != "" {
		return t.FullName()
	}
	return "no failure details"
}

// firstNonBlankLine returns the first line of s with non-whitespace
// content, stripped of ANSI.
func firstNonBlankLine(s string) string {
	s = stripANSI(s)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// shortFrameworkName normalises framework names so the label prefix
// stays short. `go test` (with a space) becomes `gotest`.
func shortFrameworkName(fw parsers.Framework) string {
	s := string(fw)
	s = strings.ReplaceAll(s, " ", "")
	return s
}

// truncateRunes right-truncates s to at most n runes, appending an
// ellipsis when it had to cut anything.
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if runeLen(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	runes := []rune(s)
	return string(runes[:n-1]) + "…"
}

// runeLen counts runes in s. Used instead of len() so multi-byte
// characters (the `…` ellipsis, unicode test names) are counted as
// one column.
func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// firstFailingLeaf walks a package's test results and returns a pointer
// to the first leaf test where Failed==true. Returns nil if none.
func firstFailingLeaf(results parsers.TestSuiteResults) *parsers.Test {
	for _, tr := range results {
		for i := range tr.Tests {
			if found := findFailingLeaf(&tr.Tests[i]); found != nil {
				return found
			}
		}
	}
	return nil
}

func findFailingLeaf(t *parsers.Test) *parsers.Test {
	for i := range t.Children {
		if found := findFailingLeaf(&t.Children[i]); found != nil {
			return found
		}
	}
	if len(t.Children) > 0 {
		return nil
	}
	if t.Failed {
		return t
	}
	return nil
}
