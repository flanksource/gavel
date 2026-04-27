package parsers

import (
	"strings"
	"testing"

	"github.com/flanksource/clicky"
)

func TestParseFailureDetail_GomegaToHavePrefix(t *testing.T) {
	msg := "Expected\n" +
		"    <string>: Add timeout status tracking and display in test results Add support for tracking and displaying test timeout status throughout the UI:\n" +
		"to have prefix\n" +
		"    <string>: add"

	d := ParseFailureDetail(msg)
	if d == nil {
		t.Fatal("expected non-nil FailureDetail")
	}
	if d.Kind != FailureKindGomega {
		t.Errorf("kind = %q, want %q", d.Kind, FailureKindGomega)
	}
	if d.Matcher != "to have prefix" {
		t.Errorf("matcher = %q, want %q", d.Matcher, "to have prefix")
	}
	if !strings.HasPrefix(d.Actual, "Add timeout status") {
		t.Errorf("actual stripped marker badly: %q", d.Actual)
	}
	if d.Expected != "add" {
		t.Errorf("expected = %q, want %q", d.Expected, "add")
	}
	if !strings.Contains(d.Summary, "to have prefix") || !strings.Contains(d.Summary, "\"add\"") {
		t.Errorf("summary missing matcher/expected: %q", d.Summary)
	}
}

func TestParseFailureDetail_GomegaPgErrorCompactRendering(t *testing.T) {
	// Live capture from .gavel/last.json — *pgconn.PgError dumped as a
	// multi-line Go struct. The registered renderer collapses it to the
	// fields a developer reads to diagnose, dropping noise like
	// SeverityUnlocalized, InternalQuery, File/Line/Routine.
	msg := "Expected\n" +
		"    <*pgconn.PgError | 0x2ff686272300>: \n" +
		"    ERROR: column reference \"deleted_at\" is ambiguous (SQLSTATE 42702)\n" +
		"    {\n" +
		"        Severity: \"ERROR\",\n" +
		"        SeverityUnlocalized: \"ERROR\",\n" +
		"        Code: \"42702\",\n" +
		"        Message: \"column reference \\\"deleted_at\\\" is ambiguous\",\n" +
		"        Detail: \"It could refer to either a PL/pgSQL variable or a table column.\",\n" +
		"        Hint: \"\",\n" +
		"        InternalQuery: \"SELECT ...\",\n" +
		"        File: \"pl_comp.c\",\n" +
		"        Line: 1138,\n" +
		"        Routine: \"plpgsql_post_column_ref\",\n" +
		"    }\n" +
		"to be nil"

	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindGomega {
		t.Fatalf("want gomega, got %#v", d)
	}
	if d.Matcher != "to be nil" {
		t.Errorf("matcher = %q", d.Matcher)
	}
	wantHead := "ERROR 42702: column reference \"deleted_at\" is ambiguous"
	if !strings.HasPrefix(d.Actual, wantHead) {
		t.Errorf("actual head = %q, want prefix %q", d.Actual, wantHead)
	}
	if !strings.Contains(d.Actual, "Detail: It could refer to either a PL/pgSQL variable") {
		t.Errorf("detail line missing; actual:\n%s", d.Actual)
	}
	if strings.Contains(d.Actual, "InternalQuery") || strings.Contains(d.Actual, "Routine") {
		t.Errorf("noise fields should be dropped; got:\n%s", d.Actual)
	}
}

func TestParseFailureDetail_GomegaToContainSubstringMultiline(t *testing.T) {
	msg := "Expected\n" +
		"    <string>: Test Results\n" +
		"    Expand\n" +
		"    Collapse\n" +
		"    JSON\n" +
		"to contain substring\n" +
		"    <string>: Parser"

	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindGomega {
		t.Fatalf("want gomega kind, got %#v", d)
	}
	if d.Matcher != "to contain substring" {
		t.Errorf("matcher = %q", d.Matcher)
	}
	if !strings.HasPrefix(d.Actual, "Test Results\nExpand\nCollapse\nJSON") {
		t.Errorf("actual didn't preserve multiline: %q", d.Actual)
	}
	if d.Expected != "Parser" {
		t.Errorf("expected = %q", d.Expected)
	}
	if !strings.Contains(d.Summary, "more line") {
		t.Errorf("summary should mention extra lines, got %q", d.Summary)
	}
}

func TestParseFailureDetail_GomegaToEqualSimple(t *testing.T) {
	msg := "Expected\n    <int>: 1\nto equal\n    <int>: 2"
	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindGomega {
		t.Fatalf("want gomega, got %#v", d)
	}
	if d.Actual != "1" {
		t.Errorf("actual = %q, want %q", d.Actual, "1")
	}
	if d.Expected != "2" {
		t.Errorf("expected = %q, want %q", d.Expected, "2")
	}
	if d.Matcher != "to equal" {
		t.Errorf("matcher = %q", d.Matcher)
	}
}

func TestParseFailureDetail_GomegaNotToBe(t *testing.T) {
	msg := "Expected\n    <string>: hello\nnot to be empty"
	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindGomega {
		t.Fatalf("want gomega, got %#v", d)
	}
	if d.Matcher != "not to be empty" {
		t.Errorf("matcher = %q", d.Matcher)
	}
	if d.Actual != "hello" {
		t.Errorf("actual = %q", d.Actual)
	}
	if d.Expected != "" {
		t.Errorf("expected should be empty for unary matcher, got %q", d.Expected)
	}
}

func TestParseFailureDetail_Panic(t *testing.T) {
	msg := "panic: runtime error: index out of range [5] with length 3\n" +
		"\n" +
		"goroutine 17 [running]:\n" +
		"github.com/flanksource/gavel/foo.Bar(...)\n" +
		"\t/repo/foo/bar.go:42 +0x123\n" +
		"main.main()\n" +
		"\t/repo/main.go:10 +0x45"

	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindPanic {
		t.Fatalf("want panic, got %#v", d)
	}
	if !strings.Contains(d.Summary, "runtime error: index out of range") {
		t.Errorf("summary = %q", d.Summary)
	}
	if !strings.HasPrefix(d.Stack, "goroutine 17 [running]:") {
		t.Errorf("stack should start with goroutine line, got %q", d.Stack)
	}
}

func TestParseFailureDetail_GoTestTrailer(t *testing.T) {
	msg := "    foo_test.go:42: assertion failed: got 1, want 2\n" +
		"--- FAIL: TestFoo (0.00s)\n" +
		"FAIL\n" +
		"FAIL\tgithub.com/x/y\t0.012s\n" +
		"exit status 1"

	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindGoTest {
		t.Fatalf("want go_test, got %#v", d)
	}
	if d.Location != "foo_test.go:42" {
		t.Errorf("location = %q", d.Location)
	}
	if !strings.Contains(d.Actual, "assertion failed: got 1, want 2") {
		t.Errorf("actual = %q", d.Actual)
	}
	if strings.Contains(d.Actual, "exit status") || strings.Contains(d.Actual, "--- FAIL") {
		t.Errorf("actual should be cleaned of trailers: %q", d.Actual)
	}
}

func TestParseFailureDetail_RawFallback(t *testing.T) {
	msg := "something went wrong somewhere"
	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindRaw {
		t.Fatalf("want raw, got %#v", d)
	}
	if d.Summary != msg {
		t.Errorf("summary = %q", d.Summary)
	}
}

func TestParseFailureDetail_RawTruncates(t *testing.T) {
	msg := strings.Repeat("a", summaryMaxLen+50)
	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindRaw {
		t.Fatalf("want raw, got %#v", d)
	}
	if len(d.Summary) > summaryMaxLen {
		t.Errorf("summary not truncated: len=%d", len(d.Summary))
	}
}

func TestParseFailureDetail_Empty(t *testing.T) {
	if d := ParseFailureDetail(""); d != nil {
		t.Errorf("empty message should return nil, got %#v", d)
	}
	if d := ParseFailureDetail("   \n   "); d != nil {
		t.Errorf("whitespace-only should return nil, got %#v", d)
	}
}

func TestPrettyUsesFailureDetailSummary(t *testing.T) {
	test := Test{
		Name:    "should equal",
		Failed:  true,
		Message: "Expected\n    <int>: 1\nto equal\n    <int>: 2",
	}
	test.FailureDetail = ParseFailureDetail(test.Message)

	rendered := clicky.MustFormat(test.Pretty())
	if !strings.Contains(rendered, "should equal") {
		t.Errorf("expected test name in output, got %q", rendered)
	}
	if !strings.Contains(rendered, "to equal") {
		t.Errorf("expected matcher in summary, got %q", rendered)
	}
	if !strings.Contains(rendered, "expected \"1\" to equal \"2\"") {
		t.Errorf("expected structured summary line, got %q", rendered)
	}
}

func TestPrettyFallsBackToRawMessageWhenNoDetail(t *testing.T) {
	test := Test{
		Name:    "weird",
		Failed:  true,
		Message: "totally arbitrary failure",
	}
	rendered := clicky.MustFormat(test.Pretty())
	if !strings.Contains(rendered, "totally arbitrary failure") {
		t.Errorf("raw message must surface when no FailureDetail, got %q", rendered)
	}
}

func TestParseFailureDetail_GomegaUnexpectedError(t *testing.T) {
	// Live capture: gomega `Expect(err).ToNot(HaveOccurred())` failure.
	// The renderer collapses *fmt.wrapError to its message, and the parser
	// synthesises the "to succeed" matcher.
	msg := "Unexpected error:\n" +
		"    <*fmt.wrapError | 0xbe30e5d1640>: \n" +
		"    starting minio: get provider: rootless Docker not found, failed to create Docker provider\n" +
		"    {\n" +
		"        msg: \"starting minio: get provider: rootless Docker not found, failed to create Docker provider\",\n" +
		"        err: <*fmt.wrapError | 0xbe30e5d1600>{msg: \"...\", err: <...>{}},\n" +
		"    }\n" +
		"occurred"

	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindGomega {
		t.Fatalf("want gomega, got %#v", d)
	}
	if d.Matcher != "to succeed" {
		t.Errorf("matcher = %q", d.Matcher)
	}
	if !strings.HasPrefix(d.Actual, "starting minio: get provider:") {
		t.Errorf("actual = %q", d.Actual)
	}
	if !strings.Contains(d.Summary, "to succeed") {
		t.Errorf("summary should mention matcher, got %q", d.Summary)
	}
}

func TestParseFailureDetail_GomegaTimeoutThenError(t *testing.T) {
	// Live capture: gomega `Eventually(...).Should(Succeed())` timing out.
	msg := "Timed out after 30.000s.\n" +
		"Expected success, but got an error:\n" +
		"    <*url.Error | 0x7fffd1a02150>: \n" +
		"    Get \"http://localhost:3100/ready\": dial tcp [::1]:3100: connect: connection refused\n" +
		"    {\n" +
		"        Op: \"Get\",\n" +
		"        URL: \"http://localhost:3100/ready\",\n" +
		"        Err: <*net.OpError | 0x7fffd13d4000>{Op: \"dial\"},\n" +
		"    }"

	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindGomega {
		t.Fatalf("want gomega, got %#v", d)
	}
	if d.Matcher != "to succeed" {
		t.Errorf("matcher = %q", d.Matcher)
	}
	if !strings.Contains(d.Summary, "timed out after 30.000s") {
		t.Errorf("summary should mention timeout, got %q", d.Summary)
	}
	// Renderer for *url.Error should produce "Get URL: Err". Err is a struct
	// here so it falls back to the leading text line.
	if !strings.Contains(d.Actual, "Get http://localhost:3100/ready") {
		t.Errorf("actual = %q", d.Actual)
	}
}

func TestParseFailureDetail_UnexpectedErrorWithoutValueBlockReturnsNil(t *testing.T) {
	// "Unexpected error:" with no indented value should not be claimed by
	// the gomega parser — it's some other error message that happens to
	// contain that string. Falls through to raw.
	msg := "Unexpected error: thing went wrong"
	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindRaw {
		t.Errorf("want raw fallback, got %#v", d)
	}
}

func TestParseFailureDetail_GomegaWithErrorMarker(t *testing.T) {
	// Gomega's HaveOccurred / MatchError uses a slightly different shape:
	// the type marker can be "<*errors.errorString | 0xc0001>: {s: \"boom\"}".
	msg := "Expected\n" +
		"    <*errors.errorString | 0xc000123>: {s: \"boom\"}\n" +
		"to be nil"
	d := ParseFailureDetail(msg)
	if d == nil || d.Kind != FailureKindGomega {
		t.Fatalf("want gomega, got %#v", d)
	}
	if !strings.Contains(d.Actual, "boom") {
		t.Errorf("actual should retain error body, got %q", d.Actual)
	}
	if d.Matcher != "to be nil" {
		t.Errorf("matcher = %q", d.Matcher)
	}
}
