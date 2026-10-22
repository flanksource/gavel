package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/github"
)

// TestPRMergeEndpointValidation verifies the POST /api/prs/merge handler
// rejects bad input before contacting either hosted PR provider.
func TestPRMergeEndpointValidation(t *testing.T) {
	s := &Server{refreshCh: make(chan struct{}, 1)}

	tests := []struct {
		name    string
		method  string
		body    string
		wantSts int
	}{
		{"wrong method", "GET", `{}`, 405},
		{"malformed json", "POST", `{not json`, 400},
		{"missing URL", "POST", `{"method":"rebase"}`, 400},
		{"blank URL", "POST", `{"url":"  ","method":"rebase"}`, 400},
		{"invalid method", "POST", `{"url":"https://github.com/acme/service/pull/27","method":"invalid"}`, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/prs/merge", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			s.handlePRMerge(rec, req)
			if rec.Code != tc.wantSts {
				t.Errorf("status = %d, want %d; body = %q", rec.Code, tc.wantSts, rec.Body.String())
			}
		})
	}
}

// TestPRApproveEndpointValidation verifies the POST /api/prs/approve handler
// applies the same shared validation.
func TestPRApproveEndpointValidation(t *testing.T) {
	s := &Server{refreshCh: make(chan struct{}, 1)}

	tests := []struct {
		name    string
		method  string
		body    string
		wantSts int
	}{
		{"wrong method", "GET", `{}`, 405},
		{"malformed json", "POST", `{nope`, 400},
		{"missing nodeId", "POST", `{"repo":"owner/name","number":42}`, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/prs/approve", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			s.handlePRApprove(rec, req)
			if rec.Code != tc.wantSts {
				t.Errorf("status = %d, want %d; body = %q", rec.Code, tc.wantSts, rec.Body.String())
			}
		})
	}
}

// TestAfterPRActionTriggersRefresh verifies a completed action requests an
// immediate poll refresh so the merged/approved state shows up promptly.
func TestAfterPRActionTriggersRefresh(t *testing.T) {
	s := &Server{refreshCh: make(chan struct{}, 1), detailCache: NewDetailCache()}
	s.afterPRAction("owner/name", 42)
	select {
	case <-s.refreshCh:
	default:
		t.Error("afterPRAction did not request a refresh")
	}
}

// TestPRUpdateBranchConflictSaysWhyAIRebaseCannotRun verifies a rebase GitHub
// refuses over conflicts answers 409 with the conflict flagged, and — with no
// project holding a local checkout of the repo — the reason the AI rebase is
// unavailable instead of its variables.
func TestPRUpdateBranchConflictSaysWhyAIRebaseCannotRun(t *testing.T) {
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"Cannot update PR branch due to conflicts"}]}`))
	}))
	t.Cleanup(gh.Close)
	t.Setenv("GITHUB_API_URL", gh.URL)
	original := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	t.Cleanup(func() { projectsPath = original })

	s := &Server{refreshCh: make(chan struct{}, 1), ghOpts: github.Options{Token: "tok"}}
	req := httptest.NewRequest("POST", "/api/prs/update-branch", bytes.NewBufferString(`{"repo":"acme/widget","number":7,"nodeId":"PR_node_7","headSha":"abc"}`))
	rec := httptest.NewRecorder()
	s.handlePRUpdateBranch(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}
	var body prUpdateConflict
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	want := prUpdateConflict{
		Error:         "update branch (rebase): the branch does not rebase cleanly onto its base: GraphQL errors: Cannot update PR branch due to conflicts",
		Conflict:      true,
		RebaseBlocked: "no project has a local checkout of acme/widget; add one to rebase it with AI",
	}
	if body != want {
		t.Errorf("body = %+v, want %+v", body, want)
	}
	select {
	case <-s.refreshCh:
		t.Error("a refused update must not request a refresh")
	default:
	}
}
