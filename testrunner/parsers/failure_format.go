package parsers

import (
	"regexp"
	"strconv"
	"strings"
)

// FailureKind classifies the structure recognised in a Test.Message.
type FailureKind string

const (
	FailureKindGomega FailureKind = "gomega"
	FailureKindPanic  FailureKind = "panic"
	FailureKindGoTest FailureKind = "go_test"
	FailureKindRaw    FailureKind = "raw"
)

// FailureDetail is the structured form of a Test.Message. Renderers use it to
// show expected/actual side-by-side and to keep the per-test summary line short
// without losing the full diagnostic.
type FailureDetail struct {
	Kind     FailureKind `json:"kind,omitempty"`
	Summary  string      `json:"summary,omitempty"`
	Matcher  string      `json:"matcher,omitempty"`
	Expected string      `json:"expected,omitempty"`
	Actual   string      `json:"actual,omitempty"`
	Location string      `json:"location,omitempty"`
	Stack    string      `json:"stack,omitempty"`
}

// summaryMaxLen caps the single-line summary used in tree views and PR
// comments. Long enough to read, short enough to fit on one terminal row.
const summaryMaxLen = 200

var (
	gomegaTypeMarkerRe   = regexp.MustCompile(`^\s*<[^>]+>:\s?`)
	gomegaMatcherLineRe  = regexp.MustCompile(`^(to\s|not\s+to\s)\S`)
	goTestFileLineRe     = regexp.MustCompile(`^\s+([^\s:]+\.go):(\d+):\s*(.*)$`)
	goTestExitStatusRe   = regexp.MustCompile(`^\s*exit status \d+\s*$`)
	goTestPackageTrailRe = regexp.MustCompile(`^(FAIL|ok|PASS)\s+\S+\s+[\d.]+s\s*$`)
	goTestDashFailRe     = regexp.MustCompile(`^---\s+(FAIL|PASS|SKIP):`)
)

// ParseFailureDetail recognises common failure shapes (gomega, panic,
// go test trailers) and returns a structured view. Returns nil for messages
// it can't classify usefully — callers fall back to rendering Message raw.
func ParseFailureDetail(msg string) *FailureDetail {
	msg = strings.TrimRight(msg, " \t\n")
	if msg == "" {
		return nil
	}
	if d := parseGomegaTimeout(msg); d != nil {
		return d
	}
	if d := parseGomegaUnexpectedError(msg); d != nil {
		return d
	}
	if d := parseGomega(msg); d != nil {
		return d
	}
	if d := parsePanic(msg); d != nil {
		return d
	}
	if d := parseGoTestTrailer(msg); d != nil {
		return d
	}
	return &FailureDetail{
		Kind:    FailureKindRaw,
		Summary: oneLine(msg, summaryMaxLen),
	}
}

var gomegaTimeoutHeaderRe = regexp.MustCompile(`^Timed out after ([\d.]+s)\.\s*$`)

// parseGomegaUnexpectedError recognises gomega's `Expect(err).ToNot(HaveOccurred())`
// / `.To(Succeed())` failure shape:
//
//	Unexpected error:
//	    <type>: <value>
//	occurred
//
// Synthesises Matcher = "to succeed" so the rendering path treats it as a
// gomega failure with an Actual (the error) and no Expected.
func parseGomegaUnexpectedError(msg string) *FailureDetail {
	lines := strings.Split(msg, "\n")
	headerIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "Unexpected error:" {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return nil
	}
	end := len(lines)
	for i := headerIdx + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "occurred" {
			end = i
			break
		}
	}
	typeName := extractGomegaType(firstNonBlank(lines, headerIdx+1, end))
	actual := joinIndentedBlock(lines, headerIdx+1, end)
	if actual == "" {
		return nil
	}
	rendered := renderGomegaValue(typeName, actual)
	return &FailureDetail{
		Kind:    FailureKindGomega,
		Matcher: "to succeed",
		Actual:  rendered,
		Summary: buildGomegaSummary(rendered, "to succeed", ""),
	}
}

// parseGomegaTimeout recognises the gomega `Eventually(...).Should(Succeed())`
// timeout shape:
//
//	Timed out after 30.000s.
//	Expected success, but got an error:
//	    <type>: <value>
//
// Captures the timeout duration into the Summary and treats the error as
// Actual.
func parseGomegaTimeout(msg string) *FailureDetail {
	lines := strings.Split(msg, "\n")
	timeoutIdx := -1
	timeout := ""
	for i, line := range lines {
		if m := gomegaTimeoutHeaderRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			timeoutIdx = i
			timeout = m[1]
			break
		}
	}
	if timeoutIdx < 0 {
		return nil
	}
	headerIdx := -1
	for i := timeoutIdx + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "Expected success, but got an error:" {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return nil
	}
	typeName := extractGomegaType(firstNonBlank(lines, headerIdx+1, len(lines)))
	actual := joinIndentedBlock(lines, headerIdx+1, len(lines))
	if actual == "" {
		return nil
	}
	rendered := renderGomegaValue(typeName, actual)
	summary := "timed out after " + timeout + " — " + summariseValue(rendered)
	return &FailureDetail{
		Kind:    FailureKindGomega,
		Matcher: "to succeed",
		Actual:  rendered,
		Summary: oneLine(summary, summaryMaxLen),
	}
}

// firstNonBlank returns the first non-blank line in lines[start:end], or "".
func firstNonBlank(lines []string, start, end int) string {
	for i := start; i < end; i++ {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}

// gomegaAnyMarkerRe matches a gomega `<type | 0xADDR>` marker in either the
// header form (`<type | 0xADDR>:`) or the inline-value form
// (`<type | 0xADDR>{...}`). Captures the type name in group 1.
var gomegaAnyMarkerRe = regexp.MustCompile(`<\s*([^>|]+?)\s*(?:\|\s*0x[0-9a-fA-F]+\s*)?>`)

// extractGomegaType pulls the Go type name out of a gomega marker — either
//
//	<*pgconn.PgError | 0x123>: ...     (header form)
//	<*errors.errorString | 0x456>{...} (inline form)
//
// returning the bare type ("*pgconn.PgError"). Returns "" when the marker is
// missing or the type name is a primitive placeholder ("string", "int", etc.)
// since those have no struct body worth rendering specially.
func extractGomegaType(line string) string {
	m := gomegaAnyMarkerRe.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	inside := strings.TrimSpace(m[1])
	// Filter out gomega's primitive placeholders.
	switch inside {
	case "string", "int", "int32", "int64", "uint", "uint32", "uint64",
		"bool", "float32", "float64", "nil":
		return ""
	}
	return inside
}

// parseGomega recognises the canonical gomega layout:
//
//	Expected
//	    <type>: value
//	to <matcher>
//	    <type>: value
//
// The "Expected" header is sometimes "Expected error" or absent on the LHS
// when the matcher is asymmetric (e.g. "Expected error: <nil>"). We stay
// strict: require an indented value after either the header or the matcher
// line, otherwise let the raw fallback handle it.
func parseGomega(msg string) *FailureDetail {
	lines := strings.Split(msg, "\n")
	headerIdx := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Expected" || strings.HasPrefix(trimmed, "Expected ") {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return nil
	}

	matcherIdx, matcher := findGomegaMatcher(lines, headerIdx+1)
	if matcherIdx < 0 {
		return nil
	}

	actualType := extractGomegaType(firstNonBlank(lines, headerIdx+1, matcherIdx))
	expectedType := extractGomegaType(firstNonBlank(lines, matcherIdx+1, len(lines)))
	actual := renderGomegaValue(actualType, joinIndentedBlock(lines, headerIdx+1, matcherIdx))
	expected := renderGomegaValue(expectedType, joinIndentedBlock(lines, matcherIdx+1, len(lines)))
	if actual == "" && expected == "" {
		return nil
	}

	d := &FailureDetail{
		Kind:     FailureKindGomega,
		Matcher:  matcher,
		Actual:   actual,
		Expected: expected,
		Summary:  buildGomegaSummary(actual, matcher, expected),
	}
	return d
}

// findGomegaMatcher walks lines starting at start looking for a line whose
// trimmed form starts with "to " or "not to " (gomega matcher phrases).
// Returns the index of the matcher line and the trimmed matcher phrase.
func findGomegaMatcher(lines []string, start int) (int, string) {
	for i := start; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}
		if gomegaMatcherLineRe.MatchString(trimmed) {
			return i, trimmed
		}
	}
	return -1, ""
}

// joinIndentedBlock collects consecutive lines between start (inclusive) and
// end (exclusive) that look like part of a gomega value block: indented (or
// blank) lines, optionally prefixed with the "<type>:" marker on the first
// non-blank line. Returns the joined value with (a) the type marker stripped,
// and (b) the common minimum leading indent removed across all non-blank
// lines so internal struct indentation is preserved. textwrap.dedent
// semantics — necessary because gomega prints multi-line Go struct literals
// as the value body and we want the inner braces to stay scannable.
func joinIndentedBlock(lines []string, start, end int) string {
	var collected []string
	for i := start; i < end; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(collected) == 0 {
				continue
			}
			collected = append(collected, "")
			continue
		}
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		collected = append(collected, strings.TrimRight(line, " \t"))
	}
	for len(collected) > 0 && collected[len(collected)-1] == "" {
		collected = collected[:len(collected)-1]
	}
	if len(collected) == 0 {
		return ""
	}
	// Compute the dedent BEFORE stripping the type marker so the marker line
	// contributes its full leading-whitespace count (gomega aligns the marker
	// at the same column as the value continuation lines).
	dedented := dedent(collected)
	out := strings.Split(dedented, "\n")
	// Now strip "<type>: " from the first non-blank line. If the marker
	// consumed the entire line (marker followed only by trailing space),
	// drop that line entirely.
	for i, line := range out {
		if strings.TrimSpace(line) == "" {
			continue
		}
		stripped := gomegaTypeMarkerRe.ReplaceAllString(line, "")
		if strings.TrimSpace(stripped) == "" {
			out = append(out[:i], out[i+1:]...)
		} else {
			out[i] = stripped
		}
		break
	}
	return strings.Join(out, "\n")
}

// dedent removes the longest common leading whitespace prefix shared by every
// non-blank line in lines. Mirrors Python's textwrap.dedent. Blank lines are
// preserved as-is.
func dedent(lines []string) string {
	prefix := ""
	prefixSet := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lead := leadingWhitespace(line)
		if !prefixSet {
			prefix = lead
			prefixSet = true
			continue
		}
		prefix = commonPrefix(prefix, lead)
		if prefix == "" {
			break
		}
	}
	if prefix == "" {
		return strings.Join(lines, "\n")
	}
	out := make([]string, len(lines))
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			out[i] = ""
			continue
		}
		out[i] = strings.TrimPrefix(line, prefix)
	}
	return strings.Join(out, "\n")
}

func leadingWhitespace(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return s[:i]
		}
	}
	return s
}

func commonPrefix(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:n]
}

// buildGomegaSummary produces the one-line headline shown next to the test
// name. Quotes the values when short, elides them otherwise.
func buildGomegaSummary(actual, matcher, expected string) string {
	if matcher == "" {
		matcher = "matched"
	}
	parts := []string{"expected"}
	if v := summariseValue(actual); v != "" {
		parts = append(parts, v)
	}
	parts = append(parts, matcher)
	if v := summariseValue(expected); v != "" {
		parts = append(parts, v)
	}
	return oneLine(strings.Join(parts, " "), summaryMaxLen)
}

// summariseValue shortens a gomega value for inline use. Multi-line values
// collapse to "<n lines>"; long single-line values get truncated.
func summariseValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.Contains(v, "\n") {
		n := strings.Count(v, "\n") + 1
		first := strings.SplitN(v, "\n", 2)[0]
		first = truncate(first, 60)
		return quote(first) + " (+" + plural(n-1, "more line") + ")"
	}
	return quote(truncate(v, 80))
}

// parsePanic recognises a Go runtime panic followed by a goroutine dump.
// The first "panic: <message>" line becomes Summary; everything from the
// first "goroutine N [" line onwards goes into Stack.
func parsePanic(msg string) *FailureDetail {
	lines := strings.Split(msg, "\n")
	panicIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "panic:") {
			panicIdx = i
			break
		}
	}
	if panicIdx < 0 {
		return nil
	}
	panicLine := strings.TrimSpace(lines[panicIdx])
	panicLine = strings.TrimPrefix(panicLine, "panic:")
	panicLine = strings.TrimSpace(panicLine)

	stackStart := -1
	for i := panicIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "goroutine ") {
			stackStart = i
			break
		}
	}

	d := &FailureDetail{
		Kind:    FailureKindPanic,
		Summary: oneLine("panic: "+panicLine, summaryMaxLen),
		Actual:  panicLine,
	}
	if stackStart > 0 {
		d.Stack = strings.TrimSpace(strings.Join(lines[stackStart:], "\n"))
	}
	return d
}

// parseGoTestTrailer recognises the noisy boilerplate go test prints around
// a real failure ("--- FAIL: ...", "FAIL\tpkg 0.5s", "exit status 1"). It
// strips that, lifts an inline "\tfile.go:NN:" prefix into Location, and
// surfaces the actual diagnostic as Summary.
func parseGoTestTrailer(msg string) *FailureDetail {
	lines := strings.Split(msg, "\n")
	var kept []string
	location := ""
	for _, line := range lines {
		if goTestDashFailRe.MatchString(line) {
			continue
		}
		if goTestExitStatusRe.MatchString(line) {
			continue
		}
		if goTestPackageTrailRe.MatchString(line) {
			continue
		}
		if location == "" {
			if m := goTestFileLineRe.FindStringSubmatch(line); m != nil {
				location = m[1] + ":" + m[2]
				if rest := strings.TrimSpace(m[3]); rest != "" {
					kept = append(kept, rest)
					continue
				}
				continue
			}
		}
		kept = append(kept, line)
	}
	body := strings.TrimSpace(strings.Join(kept, "\n"))
	if location == "" && body == msg {
		return nil
	}
	if body == "" && location == "" {
		return nil
	}
	return &FailureDetail{
		Kind:     FailureKindGoTest,
		Summary:  oneLine(body, summaryMaxLen),
		Actual:   body,
		Location: location,
	}
}

// oneLine collapses whitespace to a single line and truncates with an ellipsis
// when longer than maxLen.
func oneLine(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if idx := strings.Index(s, "\n"); idx >= 0 {
		s = strings.TrimSpace(s[:idx]) + " …"
	}
	return truncate(s, maxLen)
}

// truncate keeps the result within maxLen bytes, replacing the tail with "…"
// (3 UTF-8 bytes) when truncation occurs.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "…"
}

func quote(s string) string {
	if s == "" {
		return ""
	}
	return "\"" + s + "\""
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return strconv.Itoa(n) + " " + word + "s"
}
