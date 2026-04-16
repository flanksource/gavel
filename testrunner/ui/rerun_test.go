package testui_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func newTestServer(t *testing.T) (*testui.Server, http.Handler) {
	t.Helper()
	srv := testui.NewServer()
	return srv, srv.Handler()
}

func doRequest(t *testing.T, handler http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func doHTMLRequest(t *testing.T, handler http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Accept", "text/html")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestSnapshotIncludesLint(t *testing.T) {
	srv, handler := newTestServer(t)
	msg := "unused var"
	srv.SetLintResults([]*linters.LinterResult{{
		Linter:  "golangci-lint",
		Success: false,
		Violations: []models.Violation{{
			File: "foo.go", Line: 12, Severity: models.SeverityError, Message: &msg,
		}},
	}})

	var snap struct {
		Lint    []*linters.LinterResult `json:"lint"`
		LintRun bool                    `json:"lint_run"`
	}
	resp := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !snap.LintRun {
		t.Errorf("lint_run = false, want true")
	}
	if len(snap.Lint) != 1 || len(snap.Lint[0].Violations) != 1 {
		t.Errorf("unexpected lint payload: %+v", snap.Lint)
	}
}

func TestSnapshotIncludesRunMetadata(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.BeginRun("rerun")
	srv.MarkDone()

	var snap struct {
		Run struct {
			Sequence   int    `json:"sequence"`
			Kind       string `json:"kind"`
			StartedAt  string `json:"started_at"`
			FinishedAt string `json:"finished_at"`
		} `json:"run"`
		Done bool `json:"done"`
	}
	resp := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if snap.Run.Sequence != 1 {
		t.Fatalf("sequence = %d, want 1", snap.Run.Sequence)
	}
	if snap.Run.Kind != "rerun" {
		t.Fatalf("kind = %q, want rerun", snap.Run.Kind)
	}
	if snap.Run.StartedAt == "" || snap.Run.FinishedAt == "" {
		t.Fatalf("run timestamps missing: %+v", snap.Run)
	}
	if !snap.Done {
		t.Fatalf("snapshot should be done")
	}
}

func TestRerunRequiresPOST(t *testing.T) {
	_, handler := newTestServer(t)
	resp := doRequest(t, handler, http.MethodGet, "/api/rerun", nil)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.Code)
	}
}

func TestRerunWithoutHandlerReturns501(t *testing.T) {
	_, handler := newTestServer(t)
	resp := doRequest(t, handler, http.MethodPost, "/api/rerun", strings.NewReader(`{}`))
	if resp.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want 501", resp.Code)
	}
}

func TestRerunBadJSON(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetRerunFunc(func(testui.RerunRequest) error { return nil })

	resp := doRequest(t, handler, http.MethodPost, "/api/rerun", strings.NewReader(`not json`))
	if resp.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.Code)
	}
}

func TestRerunSuccessAndPayload(t *testing.T) {
	srv, handler := newTestServer(t)
	var got testui.RerunRequest
	srv.SetRerunFunc(func(req testui.RerunRequest) error {
		got = req
		return nil
	})

	body, _ := json.Marshal(testui.RerunRequest{
		PackagePaths: []string{"./pkg/foo"},
		TestName:     "TestX",
		Suite:        []string{"Outer", "Inner"},
		Framework:    "ginkgo",
	})
	resp := doRequest(t, handler, http.MethodPost, "/api/rerun", bytes.NewReader(body))
	if resp.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202", resp.Code)
	}
	if got.TestName != "TestX" || got.Framework != "ginkgo" || len(got.Suite) != 2 {
		t.Errorf("rerun callback received %+v", got)
	}
}

func TestRerunConcurrentReturns409(t *testing.T) {
	srv, handler := newTestServer(t)
	release := make(chan struct{})
	started := make(chan struct{})
	var firstStarted atomic.Bool
	srv.SetRerunFunc(func(testui.RerunRequest) error {
		if firstStarted.CompareAndSwap(false, true) {
			close(started)
			<-release
		}
		return nil
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = doRequest(t, handler, http.MethodPost, "/api/rerun", strings.NewReader(`{}`))
	}()

	<-started
	resp := doRequest(t, handler, http.MethodPost, "/api/rerun", strings.NewReader(`{}`))
	if resp.Code != http.StatusConflict {
		t.Errorf("second status = %d, want 409", resp.Code)
	}

	close(release)
	wg.Wait()
}

func TestSnapshotIncludesVirtualTaskTests(t *testing.T) {
	clicky.ClearGlobalTasks()
	t.Cleanup(clicky.ClearGlobalTasks)

	_, handler := newTestServer(t)
	release := make(chan struct{})
	group := clicky.StartGroup[string](testui.TestTaskGroupName, clickytask.WithConcurrency(1))
	task := group.Add("dummy", func(ctx commonsContext.Context, t *clickytask.Task) (string, error) {
		t.SetName("go test -json ./pkg/foo")
		t.Infof("go test -json ./pkg/foo")
		<-release
		t.Success()
		return "ok", nil
	})
	defer close(release)

	time.Sleep(50 * time.Millisecond)

	var snap struct {
		Tests []parsers.Test `json:"tests"`
	}
	resp := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(snap.Tests) == 0 {
		t.Fatalf("expected task tests in snapshot")
	}

	virtualGroup := snap.Tests[0]
	if virtualGroup.Framework != parsers.Framework("task") {
		t.Fatalf("group framework = %q, want task", virtualGroup.Framework)
	}
	if virtualGroup.Name != testui.TestTaskGroupName {
		t.Fatalf("group name = %q, want %q", virtualGroup.Name, testui.TestTaskGroupName)
	}
	if len(virtualGroup.Children) == 0 {
		t.Fatalf("expected child task under virtual group")
	}

	foundTask := false
	for _, child := range virtualGroup.Children {
		if child.Framework == parsers.Framework("task") && child.Command == "go test -json ./pkg/foo" {
			foundTask = true
			if !child.Pending {
				t.Fatalf("child task should be pending/running: %+v", child)
			}
			if !strings.Contains(child.Stderr, "go test -json ./pkg/foo") {
				t.Fatalf("child stderr should contain task logs, got %q", child.Stderr)
			}
		}
	}
	if !foundTask {
		t.Fatalf("missing task child in %+v", virtualGroup.Children)
	}

	release <- struct{}{}
	if _, err := task.GetResult(); err != nil {
		t.Fatalf("task get result: %v", err)
	}
}

func TestNestedRoutesServeHTMLShell(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetResults([]parsers.Test{{
		Name:      "testrunner/",
		Framework: parsers.GoTest,
		Children: parsers.Tests{{
			Name:      "TestBuildFailed",
			Framework: parsers.GoTest,
			Failed:    true,
			Message:   "boom",
		}},
	}})

	resp := doHTMLRequest(t, handler, http.MethodGet, "/tests/testrunner/build-failed")
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.Code)
	}
	if got := resp.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("content-type = %q, want html", got)
	}
	if !strings.Contains(resp.Body.String(), `<div id="root"></div>`) {
		t.Fatalf("expected html shell, got %s", resp.Body.String())
	}
}

func TestSelectedTestJSONExport(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetResults([]parsers.Test{{
		Name:      "testrunner/",
		Framework: parsers.GoTest,
		Children: parsers.Tests{
			{
				Name:      "TestBuildFailed",
				Framework: parsers.GoTest,
				Failed:    true,
				Message:   "boom",
			},
			{
				Name:      "TestParser",
				Framework: parsers.GoTest,
				Passed:    true,
			},
		},
	}})

	var report struct {
		Tab      string `json:"tab"`
		Path     string `json:"path"`
		Selected *struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"selected"`
		Tests []struct {
			Name string `json:"name"`
		} `json:"tests"`
	}

	resp := doRequest(t, handler, http.MethodGet, "/tests/testrunner/build-failed.json?status=failed&framework=go%20test", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("content-type = %q, want json", got)
	}
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if report.Tab != "tests" {
		t.Fatalf("tab = %q, want tests", report.Tab)
	}
	if report.Path != "testrunner/build-failed" {
		t.Fatalf("path = %q", report.Path)
	}
	if report.Selected == nil || report.Selected.Name != "Build Failed" || report.Selected.Status != "failed" {
		t.Fatalf("unexpected selected payload: %+v", report.Selected)
	}
	if len(report.Tests) != 1 || report.Tests[0].Name != "Build Failed" {
		t.Fatalf("unexpected tests payload: %+v", report.Tests)
	}
}

func TestFilteredTestsJSONExportUsesTestsField(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetResults([]parsers.Test{{
		Name:      "testrunner/",
		Framework: parsers.GoTest,
		Children: parsers.Tests{
			{Name: "TestBuildFailed", Framework: parsers.GoTest, Failed: true, Message: "boom"},
			{Name: "TestParser", Framework: parsers.GoTest, Passed: true},
		},
	}})

	var report struct {
		Tab   string `json:"tab"`
		Tests []struct {
			Name     string `json:"name"`
			Path     string `json:"path"`
			Children []struct {
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"children"`
		} `json:"tests"`
	}

	resp := doRequest(t, handler, http.MethodGet, "/tests.json?status=failed", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", resp.Code, resp.Body.String())
	}
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if report.Tab != "tests" {
		t.Fatalf("tab = %q, want tests", report.Tab)
	}
	if len(report.Tests) != 1 {
		t.Fatalf("expected 1 test root, got %+v", report.Tests)
	}
	if len(report.Tests[0].Children) != 1 || report.Tests[0].Children[0].Name != "Build Failed" || report.Tests[0].Children[0].Status != "failed" {
		t.Fatalf("unexpected filtered children: %+v", report.Tests[0].Children)
	}
}

func TestLintMarkdownExport(t *testing.T) {
	srv, handler := newTestServer(t)
	msg := "unused var"
	srv.SetLintResults([]*linters.LinterResult{{
		Linter:  "golangci-lint",
		Success: false,
		Violations: []models.Violation{{
			File: "foo.go", Line: 12, Severity: models.SeverityError, Message: &msg,
		}},
	}})

	resp := doRequest(t, handler, http.MethodGet, "/lint.md?linter=golangci-lint", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Content-Type"); !strings.Contains(got, "text/markdown") {
		t.Fatalf("content-type = %q, want markdown", got)
	}
	body := resp.Body.String()
	if !strings.Contains(body, "golangci-lint") {
		t.Fatalf("expected markdown export to include linter name, got %s", body)
	}
	if !strings.Contains(body, "foo.go") {
		t.Fatalf("expected markdown export to include file name, got %s", body)
	}
}
