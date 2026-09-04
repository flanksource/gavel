package ui

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

// TestPRMergeEndpointValidation verifies the POST /api/prs/merge handler
// rejects bad input before reaching GitHub. The happy path needs a real token
// + GitHub, so it's covered by the github package's mutation tests instead.
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
		{"missing nodeId", "POST", `{"repo":"owner/name","number":42}`, 400},
		{"blank nodeId", "POST", `{"repo":"owner/name","number":42,"nodeId":"  "}`, 400},
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
