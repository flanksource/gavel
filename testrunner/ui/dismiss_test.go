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

func TestLintDismissLastsAcrossResultsUntilRestored(t *testing.T) {
	const (
		kept      = "fp-kept"
		dismissed = "fp-dismissed"
	)
	srv, handler := newTestServer(t)
	review := func() []*linters.LinterResult {
		return []*linters.LinterResult{{Linter: "markdown-review", Violations: []models.Violation{{Fingerprint: kept}, {Fingerprint: dismissed}}}}
	}
	post := func(req testui.DismissRequest) {
		t.Helper()
		body, _ := json.Marshal(req)
		if resp := doRequest(t, handler, http.MethodPost, "/api/lint/dismiss", bytes.NewReader(body)); resp.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", resp.Code, resp.Body.String())
		}
	}
	listed := func() (open, set []string) {
		var snap testui.Snapshot
		if err := json.NewDecoder(doRequest(t, handler, http.MethodGet, "/api/tests", nil).Body).Decode(&snap); err != nil {
			t.Fatalf("decode: %v", err)
		}
		for _, v := range snap.Lint[0].Violations {
			open = append(open, v.Fingerprint)
		}
		for _, v := range snap.Lint[0].Dismissed {
			set = append(set, v.Fingerprint)
		}
		return open, set
	}

	srv.SetLintResults(review())
	post(testui.DismissRequest{Fingerprints: []string{dismissed}})
	srv.SetLintResults(review())
	openAfterRerun, setAfterRerun := listed()
	post(testui.DismissRequest{Fingerprints: []string{dismissed}, Restore: true})
	openAfterRestore, _ := listed()

	if len(openAfterRerun) != 1 || openAfterRerun[0] != kept || len(setAfterRerun) != 1 || setAfterRerun[0] != dismissed || len(openAfterRestore) != 2 {
		t.Fatalf("after rerun open=%v dismissed=%v; after restore open=%v", openAfterRerun, setAfterRerun, openAfterRestore)
	}
}

func TestLintDismissRequiresFingerprints(t *testing.T) {
	_, handler := newTestServer(t)
	resp := doRequest(t, handler, http.MethodPost, "/api/lint/dismiss", bytes.NewReader([]byte(`{}`)))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.Code)
	}
}
