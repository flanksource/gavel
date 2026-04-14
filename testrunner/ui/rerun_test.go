package testui_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func newTestServer(t *testing.T) (*testui.Server, *httptest.Server) {
	t.Helper()
	srv := testui.NewServer()
	httpSrv := httptest.NewServer(srv.Handler())
	t.Cleanup(httpSrv.Close)
	return srv, httpSrv
}

func TestSnapshotIncludesLint(t *testing.T) {
	srv, httpSrv := newTestServer(t)
	msg := "unused var"
	srv.SetLintResults([]*linters.LinterResult{{
		Linter:  "golangci-lint",
		Success: false,
		Violations: []models.Violation{{
			File: "foo.go", Line: 12, Severity: models.SeverityError, Message: &msg,
		}},
	}})

	resp, err := http.Get(httpSrv.URL + "/api/tests")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	var snap struct {
		Lint    []*linters.LinterResult `json:"lint"`
		LintRun bool                    `json:"lint_run"`
	}
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

func TestRerunRequiresPOST(t *testing.T) {
	_, httpSrv := newTestServer(t)
	resp, err := http.Get(httpSrv.URL + "/api/rerun")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestRerunWithoutHandlerReturns501(t *testing.T) {
	_, httpSrv := newTestServer(t)
	resp, err := http.Post(httpSrv.URL+"/api/rerun", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("status = %d, want 501", resp.StatusCode)
	}
}

func TestRerunBadJSON(t *testing.T) {
	srv, httpSrv := newTestServer(t)
	srv.SetRerunFunc(func(testui.RerunRequest) error { return nil })

	resp, err := http.Post(httpSrv.URL+"/api/rerun", "application/json", strings.NewReader(`not json`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestRerunSuccessAndPayload(t *testing.T) {
	srv, httpSrv := newTestServer(t)
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
	resp, err := http.Post(httpSrv.URL+"/api/rerun", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status = %d, want 202", resp.StatusCode)
	}
	if got.TestName != "TestX" || got.Framework != "ginkgo" || len(got.Suite) != 2 {
		t.Errorf("rerun callback received %+v", got)
	}
}

func TestRerunConcurrentReturns409(t *testing.T) {
	srv, httpSrv := newTestServer(t)
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
		resp, err := http.Post(httpSrv.URL+"/api/rerun", "application/json", strings.NewReader(`{}`))
		if err != nil {
			t.Errorf("first post: %v", err)
			return
		}
		resp.Body.Close()
	}()

	<-started
	resp, err := http.Post(httpSrv.URL+"/api/rerun", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("second post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("second status = %d, want 409", resp.StatusCode)
	}

	close(release)
	wg.Wait()
}
