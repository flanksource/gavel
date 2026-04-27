package parsers

import (
	"strings"
	"testing"
)

func TestRenderPgError_LiveSample(t *testing.T) {
	body := "ERROR: column reference \"deleted_at\" is ambiguous (SQLSTATE 42702)\n" +
		"{\n" +
		"    Severity: \"ERROR\",\n" +
		"    SeverityUnlocalized: \"ERROR\",\n" +
		"    Code: \"42702\",\n" +
		"    Message: \"column reference \\\"deleted_at\\\" is ambiguous\",\n" +
		"    Detail: \"It could refer to either a PL/pgSQL variable or a table column.\",\n" +
		"    Hint: \"\",\n" +
		"    Where: \"PL/pgSQL function related_changes_recursive line 17 at RETURN QUERY\",\n" +
		"    File: \"pl_comp.c\",\n" +
		"    Line: 1138,\n" +
		"    Routine: \"plpgsql_post_column_ref\",\n" +
		"}"
	out := renderGomegaValue("*pgconn.PgError", body)
	if !strings.HasPrefix(out, "ERROR 42702: column reference \"deleted_at\" is ambiguous") {
		t.Errorf("head wrong; got:\n%s", out)
	}
	if !strings.Contains(out, "Detail: It could refer to either a PL/pgSQL variable") {
		t.Errorf("Detail line missing; got:\n%s", out)
	}
	if !strings.Contains(out, "Where: PL/pgSQL function") {
		t.Errorf("Where line missing; got:\n%s", out)
	}
	for _, banned := range []string{"InternalQuery", "Routine", "SeverityUnlocalized", "pl_comp.c"} {
		if strings.Contains(out, banned) {
			t.Errorf("noise field %q should be dropped; got:\n%s", banned, out)
		}
	}
}

func TestRenderURLError_LiveSample(t *testing.T) {
	body := "Get \"http://localhost:3100/ready\": dial tcp [::1]:3100: connect: connection refused\n" +
		"{\n" +
		"    Op: \"Get\",\n" +
		"    URL: \"http://localhost:3100/ready\",\n" +
		"    Err: <*net.OpError | 0x7fffd13d4000>{Op: \"dial\"},\n" +
		"}"
	out := renderGomegaValue("*url.Error", body)
	// Err is a struct so renderURLError falls back to the leading line.
	want := "Get http://localhost:3100/ready: dial tcp [::1]:3100: connect: connection refused"
	if !strings.Contains(out, want) {
		t.Errorf("want %q in output, got:\n%s", want, out)
	}
}

func TestRenderWrapError_LiveSample(t *testing.T) {
	body := "starting minio: get provider: rootless Docker not found\n" +
		"{\n" +
		"    msg: \"starting minio: get provider: rootless Docker not found\",\n" +
		"    err: <*fmt.wrapError | 0xbe30e5d1600>{msg: \"...\", err: <...>{}},\n" +
		"}"
	out := renderGomegaValue("*fmt.wrapError", body)
	if out != "starting minio: get provider: rootless Docker not found" {
		t.Errorf("got %q", out)
	}
}

func TestRenderErrorString(t *testing.T) {
	body := "rootless Docker not found\n{\n    s: \"rootless Docker not found\",\n}"
	out := renderGomegaValue("*errors.errorString", body)
	if out != "rootless Docker not found" {
		t.Errorf("got %q", out)
	}
}

func TestRenderJoinError(t *testing.T) {
	body := "rootless Docker not found\n" +
		"{\n" +
		"    errs: [\n" +
		"        <*errors.errorString | 0x10db646f0>{s: \"rootless Docker not found\"},\n" +
		"        <*errors.errorString | 0x10db64700>{s: \"socket not available\"},\n" +
		"    ],\n" +
		"}"
	out := renderGomegaValue("*errors.joinError", body)
	if !strings.Contains(out, "• rootless Docker not found") {
		t.Errorf("first bullet missing; got:\n%s", out)
	}
	if !strings.Contains(out, "• socket not available") {
		t.Errorf("second bullet missing; got:\n%s", out)
	}
}

func TestRegisterValueRenderer_OverrideAndFallback(t *testing.T) {
	called := false
	RegisterValueRenderer("*foo.Bar", ValueRendererFunc(func(b string) string {
		called = true
		return "custom: " + b
	}))
	out := renderGomegaValue("*foo.Bar", "raw body")
	if !called {
		t.Error("custom renderer was not invoked")
	}
	if !strings.Contains(out, "custom: raw body") {
		t.Errorf("custom output not surfaced; got %q", out)
	}

	// Unknown type goes through prettyDefault only — pointer suffix stripped,
	// trailing comma before close brace removed.
	body := "<*pkg.Other | 0xdeadbeef>{\n    Field: \"x\",\n}"
	got := renderGomegaValue("*not.Registered", body)
	if strings.Contains(got, "0xdeadbeef") {
		t.Errorf("pointer hex should be stripped, got %q", got)
	}
}

func TestPrettyDefault_StripsPointerNoise(t *testing.T) {
	in := "<*foo.Bar | 0xdeadbeef>{\n" +
		"    msg: \"x\",\n" +
		"    inner: <*foo.Inner | 0xcafebabe>{},\n" +
		"}"
	out := prettyDefault(in)
	if strings.Contains(out, "0xdeadbeef") || strings.Contains(out, "0xcafebabe") {
		t.Errorf("hex addresses survived; got %q", out)
	}
	if !strings.Contains(out, "<*foo.Bar>") || !strings.Contains(out, "<*foo.Inner>") {
		t.Errorf("type names should remain after stripping pointer; got %q", out)
	}
}

func TestStructFields_SimpleAndNested(t *testing.T) {
	body := "{\n" +
		"    A: \"x\",\n" +
		"    B: 42,\n" +
		"    C: <inner>{\n" +
		"        D: \"y\",\n" +
		"    },\n" +
		"    E: \"end\",\n" +
		"}"
	fields, keys := structFields(body)
	if len(fields) != 4 {
		t.Errorf("want 4 top-level fields, got %d: %v", len(fields), keys)
	}
	if unquote(fields["A"]) != "x" {
		t.Errorf("A = %q", fields["A"])
	}
	if fields["B"] != "42" {
		t.Errorf("B = %q", fields["B"])
	}
	if !strings.Contains(fields["C"], "D:") {
		t.Errorf("nested C should contain D; got %q", fields["C"])
	}
	if unquote(fields["E"]) != "end" {
		t.Errorf("E = %q", fields["E"])
	}
}
