package outline

import (
	"strings"
	"unicode"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// noiseCalls are assertion/harness calls that say nothing about what a test
// exercises.
var noiseCalls = map[string]bool{
	"Expect": true, "Eventually": true, "Consistently": true, "Fail": true,
	"RegisterFailHandler": true, "RunSpecs": true, "By": true,
	"BeforeEach": true, "AfterEach": true, "DeferCleanup": true,
	"make": true, "append": true, "len": true, "cap": true, "new": true,
	"panic": true, "recover": true, "copy": true, "delete": true,
}

var noisePackages = map[string]bool{
	"assert": true, "require": true, "mock": true, "testify": true,
	"fmt": true, "strings": true, "os": true, "filepath": true, "time": true,
	"errors": true, "context": true, "json": true, "sort": true,
	"strconv": true, "bytes": true, "io": true, "ginkgo": true, "gomega": true,
}

// applyDescriptions sets a zero-cost static description on every leaf.
func applyDescriptions(report *Report) {
	for _, leaf := range report.Leaves() {
		leaf.Description = staticDescription(leaf)
	}
}

// staticDescription derives WHAT a test verifies from its name and hierarchy.
// Ginkgo and vitest descriptions already read as sentences; go test names are
// humanized and annotated with the functions the body exercises.
func staticDescription(e *Entry) string {
	if e.Framework != parsers.GoTest {
		return strings.Join(append(append([]string(nil), e.Suite...), e.Name), " ")
	}

	desc := humanizeGoTestName(e.Name)
	if notable := notableCalls(e.calls, 3); len(notable) > 0 {
		desc += " — exercises " + strings.Join(notable, ", ")
	}
	return desc
}

func notableCalls(calls []string, limit int) []string {
	var notable []string
	for _, call := range calls {
		pkg, _, qualified := strings.Cut(call, ".")
		if noiseCalls[call] || (qualified && noisePackages[pkg]) {
			continue
		}
		notable = append(notable, call)
		if len(notable) == limit {
			break
		}
	}
	return notable
}

// humanizeGoTestName turns "TestBuildLocationMap_SkipsVendor/edge case" into
// "build location map skips vendor / edge case".
func humanizeGoTestName(name string) string {
	parts := strings.Split(name, "/")
	for i, part := range parts {
		part = strings.TrimPrefix(part, "Test")
		part = strings.ReplaceAll(part, "_", " ")
		parts[i] = strings.TrimSpace(strings.ToLower(splitCamel(part)))
	}
	return strings.Join(parts, " / ")
}

func splitCamel(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(runes[i-1]) ||
			(i+1 < len(runes) && unicode.IsLower(runes[i+1]) && unicode.IsUpper(runes[i-1]))) {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
