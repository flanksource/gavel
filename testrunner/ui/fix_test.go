package testui_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestLintFixSendsOnlyOpenRequestedFindings(t *testing.T) {
	const (
		open      = "fp-open"
		dismissed = "fp-dismissed"
		model     = "agent:sonnet"
	)
	srv, handler := newTestServer(t)
	srv.SetLintResults([]*linters.LinterResult{{Linter: "vale", Violations: []models.Violation{{Fingerprint: open}, {Fingerprint: dismissed}}}})
	var received []testui.LintFixRequest
	srv.SetLintFixFunc(func(req testui.LintFixRequest, _ *testui.RerunOutputBuffer) error {
		received = append(received, req)
		srv.SetLintFix(&testui.LintFixRecord{Fingerprints: req.Fingerprints, Model: req.Model, Iterations: 1, StopReason: "condition-met"})
		return nil
	})
	post := func(path string, body any) *http.Response {
		t.Helper()
		data, _ := json.Marshal(body)
		return doRequest(t, handler, http.MethodPost, path, bytes.NewReader(data)).Result()
	}
	post("/api/lint/dismiss", testui.DismissRequest{Fingerprints: []string{dismissed}})

	stale := post("/api/lint/fix", testui.LintFixRequest{Fingerprints: []string{open, dismissed}, Model: model})
	fixed := post("/api/lint/fix", testui.LintFixRequest{Fingerprints: []string{open}, Model: model})
	var record testui.LintFixRecord
	if err := json.NewDecoder(fixed.Body).Decode(&record); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if stale.StatusCode != http.StatusConflict || fixed.StatusCode != http.StatusOK {
		t.Fatalf("statuses = %d (dismissed requested), %d (open only); want 409, 200", stale.StatusCode, fixed.StatusCode)
	}
	if len(received) != 1 || len(received[0].Fingerprints) != 1 || received[0].Fingerprints[0] != open || received[0].Model != model {
		t.Fatalf("fixer received %+v, want one request for %s with %s", received, open, model)
	}
	if record.StopReason != "condition-met" || srv.Snapshot().LintFix == nil || !srv.Snapshot().Status.LintFixSupported {
		t.Fatalf("record = %+v, snapshot fix = %+v", record, srv.Snapshot().LintFix)
	}
}

func TestLintFixIsUnsupportedWithoutAFixer(t *testing.T) {
	_, handler := newTestServer(t)
	resp := doRequest(t, handler, http.MethodPost, "/api/lint/fix", bytes.NewReader([]byte(`{"fingerprints":["x"]}`)))
	if resp.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501", resp.Code)
	}
}
