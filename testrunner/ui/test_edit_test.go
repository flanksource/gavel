package testui_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestTestEditEndpointSkipsGoTestAndUpdatesSnapshot(t *testing.T) {
	repo := testEditRepo(t)
	writeTestFile(t, filepath.Join(repo, "foo_test.go"), `package foo

import "testing"

func TestAlpha(t *testing.T) {
	t.Fatal("boom")
}
`)

	srv, handler := newTestServer(t)
	srv.SetGitRoot(repo)
	srv.SetResults([]parsers.Test{{
		Name:      "TestAlpha",
		Framework: parsers.GoTest,
		File:      "foo_test.go",
		Line:      5,
		Passed:    true,
	}})

	resp := postTestEdit(t, handler, testui.TestEditRequest{
		Action:    "skip",
		Scope:     "test",
		Framework: parsers.GoTest.String(),
		File:      "foo_test.go",
		Line:      5,
		TestName:  "TestAlpha",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", resp.Code, resp.Body.String())
	}

	data := readTestFile(t, filepath.Join(repo, "foo_test.go"))
	if !strings.Contains(data, `t.Skip("skipped by gavel")`) {
		t.Fatalf("skip not inserted:\n%s", data)
	}

	var snap testui.Snapshot
	got := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	if err := json.NewDecoder(got.Body).Decode(&snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if !snap.Status.TestEditSupported {
		t.Fatalf("test_edit_supported = false, want true")
	}
	if len(snap.Tests) != 1 || !snap.Tests[0].Skipped || snap.Tests[0].Passed {
		t.Fatalf("snapshot test not marked skipped: %+v", snap.Tests)
	}
}

func TestSnapshotSerializesTestEditSupportedFalse(t *testing.T) {
	_, handler := newTestServer(t)

	resp := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), `"test_edit_supported":false`) {
		t.Fatalf("snapshot should serialize unsupported test edit flag, got: %s", resp.Body.String())
	}
}

func TestTestEditEndpointRejectsPathTraversal(t *testing.T) {
	repo := testEditRepo(t)
	srv, handler := newTestServer(t)
	srv.SetGitRoot(repo)

	resp := postTestEdit(t, handler, testui.TestEditRequest{
		Action:    "skip",
		Scope:     "test",
		Framework: parsers.GoTest.String(),
		File:      "../outside_test.go",
		TestName:  "TestAlpha",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", resp.Code, resp.Body.String())
	}
}

func TestTestEditEndpointRejectsUnknownWorkDir(t *testing.T) {
	repo := testEditRepo(t)
	other := testEditRepo(t)
	writeTestFile(t, filepath.Join(other, "foo_test.go"), `package foo

import "testing"

func TestAlpha(t *testing.T) {}
`)

	srv, handler := newTestServer(t)
	srv.SetResults([]parsers.Test{{
		Name:      "TestAlpha",
		Framework: parsers.GoTest,
		WorkDir:   repo,
		File:      "foo_test.go",
		Line:      5,
		Passed:    true,
	}})

	resp := postTestEdit(t, handler, testui.TestEditRequest{
		Action:    "skip",
		Scope:     "test",
		Framework: parsers.GoTest.String(),
		WorkDir:   other,
		File:      "foo_test.go",
		TestName:  "TestAlpha",
	})
	if resp.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501: %s", resp.Code, resp.Body.String())
	}
}

func TestTestEditSkipsGinkgoFile(t *testing.T) {
	repo := testEditRepo(t)
	path := filepath.Join(repo, "spec_test.go")
	writeTestFile(t, path, `package foo

import ginkgo "github.com/onsi/ginkgo/v2"

var _ = ginkgo.Describe("suite", func() {
	ginkgo.It("works", func() {})
	ginkgo.Entry("case", 1)
})
`)

	srv, handler := newTestServer(t)
	srv.SetGitRoot(repo)
	resp := postTestEdit(t, handler, testui.TestEditRequest{
		Action:    "skip",
		Scope:     "file",
		Framework: parsers.Ginkgo.String(),
		File:      "spec_test.go",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", resp.Code, resp.Body.String())
	}
	data := readTestFile(t, path)
	if !strings.Contains(data, "ginkgo.PIt(") || !strings.Contains(data, "ginkgo.PEntry(") {
		t.Fatalf("ginkgo calls not made pending:\n%s", data)
	}
}

func TestTestEditVitestSkipOnlyAndDelete(t *testing.T) {
	repo := testEditRepo(t)
	path := filepath.Join(repo, "sum.test.ts")
	writeTestFile(t, path, `import { it, expect } from 'vitest';

it.only("works", () => {
	expect(1).toBe(1);
});

it("remove me", () => {
	expect(2).toBe(2);
});
`)

	srv, handler := newTestServer(t)
	srv.SetGitRoot(repo)
	resp := postTestEdit(t, handler, testui.TestEditRequest{
		Action:    "skip",
		Scope:     "test",
		Framework: parsers.Vitest.String(),
		File:      "sum.test.ts",
		Line:      3,
		TestName:  "works",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("skip status = %d, want 200: %s", resp.Code, resp.Body.String())
	}
	data := readTestFile(t, path)
	if !strings.Contains(data, `it.skip("works"`) || strings.Contains(data, `it.only`) {
		t.Fatalf("vitest only not converted to skip:\n%s", data)
	}

	resp = postTestEdit(t, handler, testui.TestEditRequest{
		Action:    "delete",
		Scope:     "test",
		Framework: parsers.Vitest.String(),
		File:      "sum.test.ts",
		Line:      7,
		TestName:  "remove me",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200: %s", resp.Code, resp.Body.String())
	}
	data = readTestFile(t, path)
	if strings.Contains(data, "remove me") {
		t.Fatalf("vitest test was not deleted:\n%s", data)
	}
}

func TestTestEditVitestRejectsTableForm(t *testing.T) {
	repo := testEditRepo(t)
	writeTestFile(t, filepath.Join(repo, "table.test.ts"), `import { test } from 'vitest';

test.each([[1]])("case %s", () => {});
`)

	srv, handler := newTestServer(t)
	srv.SetGitRoot(repo)
	resp := postTestEdit(t, handler, testui.TestEditRequest{
		Action:    "skip",
		Scope:     "test",
		Framework: parsers.Vitest.String(),
		File:      "table.test.ts",
		Line:      3,
		TestName:  "case %s",
	})
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", resp.Code, resp.Body.String())
	}
}

func testEditRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("create .git: %v", err)
	}
	return repo
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func postTestEdit(t *testing.T, handler http.Handler, req testui.TestEditRequest) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return doRequest(t, handler, http.MethodPost, "/api/tests/edit", bytes.NewReader(body))
}
