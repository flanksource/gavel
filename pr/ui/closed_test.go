package ui

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// TestClosedEndpointValidation verifies POST /api/prs/closed rejects bad input
// without mutating server state.
func TestClosedEndpointValidation(t *testing.T) {
	s := &Server{refreshCh: make(chan struct{}, 1)}

	tests := []struct {
		name    string
		method  string
		body    string
		wantSts int
	}{
		{"wrong method", "GET", `{}`, 405},
		{"malformed json", "POST", `{not json`, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/prs/closed", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			s.handleClosed(rec, req)
			if rec.Code != tc.wantSts {
				t.Errorf("status = %d, want %d; body = %q", rec.Code, tc.wantSts, rec.Body.String())
			}
			if s.ShowClosed() {
				t.Error("invalid request must not enable showClosed")
			}
		})
	}
}

// TestClosedEndpointTogglesAndRefetches verifies a valid POST flips the flag and
// requests an immediate refetch so the widened fetch takes effect at once.
func TestClosedEndpointTogglesAndRefetches(t *testing.T) {
	s := &Server{refreshCh: make(chan struct{}, 1)}
	req := httptest.NewRequest("POST", "/api/prs/closed", bytes.NewBufferString(`{"show":true}`))
	rec := httptest.NewRecorder()
	s.handleClosed(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if !s.ShowClosed() {
		t.Error("ShowClosed() = false after POST {show:true}")
	}
	select {
	case <-s.refreshCh:
	default:
		t.Error("enabling showClosed did not request a refetch")
	}
}

// TestSetShowClosedRefetchesOnlyOnChange verifies a no-op set does not spend a
// refetch, while a real change does — mirroring SetIncludeBots.
func TestSetShowClosedRefetchesOnlyOnChange(t *testing.T) {
	s := &Server{refreshCh: make(chan struct{}, 1)}

	s.SetShowClosed(false) // default is already false → no refetch
	select {
	case <-s.refreshCh:
		t.Error("SetShowClosed(false) with no change requested a refetch")
	default:
	}

	s.SetShowClosed(true) // false → true is a change → refetch
	select {
	case <-s.refreshCh:
	default:
		t.Error("SetShowClosed(true) did not request a refetch")
	}
}

// TestSnapshotLockedIncludesShowClosed verifies the flag reaches the wire
// snapshot so the UI can converge its Closed/Merged chips with server state.
func TestSnapshotLockedIncludesShowClosed(t *testing.T) {
	s := &Server{}
	s.showClosed = true
	if !s.snapshotLocked().ShowClosed {
		t.Error("snapshotLocked did not carry showClosed")
	}
}

// TestSnapshotShowClosedOmitEmpty verifies the wire format: showClosed is present
// only when on (omitempty keeps the default-off snapshot lean).
func TestSnapshotShowClosedOmitEmpty(t *testing.T) {
	on, err := json.Marshal(snapshot{ShowClosed: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Contains(on, []byte(`"showClosed":true`)) {
		t.Errorf("snapshot with showClosed=true missing field: %s", on)
	}

	off, err := json.Marshal(snapshot{ShowClosed: false})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if bytes.Contains(off, []byte("showClosed")) {
		t.Errorf("snapshot with showClosed=false should omit field: %s", off)
	}
}
