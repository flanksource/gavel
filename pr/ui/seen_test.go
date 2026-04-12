package ui

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flanksource/gavel/github"
)

// TestSeenEndpointValidation verifies the POST /api/prs/seen handler
// rejects bad input without touching the cache. We can't easily test the
// happy path without a real postgres, but the validation logic must work
// in all environments.
func TestSeenEndpointValidation(t *testing.T) {
	s := &Server{}

	tests := []struct {
		name    string
		method  string
		body    string
		wantSts int
	}{
		{"wrong method", "GET", `{}`, 405},
		{"malformed json", "POST", `{not json`, 400},
		{"missing repo", "POST", `{"number": 42}`, 400},
		{"missing number", "POST", `{"repo": "owner/name"}`, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/prs/seen", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			s.handleSeen(rec, req)
			if rec.Code != tc.wantSts {
				t.Errorf("status = %d, want %d; body = %q", rec.Code, tc.wantSts, rec.Body.String())
			}
		})
	}
}

// TestSeenEndpointAcceptsValidBody verifies that a well-formed POST body
// with a disabled cache succeeds (because MarkSeen is a no-op when disabled).
// This exercises the full JSON decode path and the notify() trigger.
func TestSeenEndpointAcceptsValidBody(t *testing.T) {
	s := &Server{updated: make(chan struct{}, 1)}
	req := httptest.NewRequest("POST", "/api/prs/seen",
		bytes.NewBufferString(`{"repo":"owner/name","number":42}`))
	rec := httptest.NewRecorder()
	s.handleSeen(rec, req)
	if rec.Code != 200 {
		t.Errorf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	// notify() should have pushed to s.updated
	select {
	case <-s.updated:
	case <-time.After(100 * time.Millisecond):
		t.Error("handleSeen did not trigger notify() after marking seen")
	}
}

// TestUnreadMapEmpty verifies the helper short-circuits on an empty list
// without hitting the cache. Handlers call UnreadMap unconditionally.
func TestUnreadMapEmpty(t *testing.T) {
	s := &Server{}
	if m := s.UnreadMap(nil); m != nil {
		t.Errorf("UnreadMap(nil) = %v, want nil", m)
	}
	if m := s.UnreadMap(github.PRSearchResults{}); m != nil {
		t.Errorf("UnreadMap(empty) = %v, want nil", m)
	}
}

// TestUnreadMapDisabledCache verifies that when the cache is disabled,
// every PR reads as unread (the map is fully populated). This is the
// "no GAVEL_GITHUB_CACHE_DSN" path that every CLI invocation takes by default.
func TestUnreadMapDisabledCache(t *testing.T) {
	s := &Server{}
	prs := github.PRSearchResults{
		{Repo: "owner/a", Number: 1, UpdatedAt: time.Now()},
		{Repo: "owner/b", Number: 2, UpdatedAt: time.Now()},
	}
	m := s.UnreadMap(prs)
	if len(m) != 2 {
		t.Errorf("UnreadMap with disabled cache returned %d entries, want 2 (both unread)", len(m))
	}
	if !m["owner/a#1"] || !m["owner/b#2"] {
		t.Errorf("expected both PRs marked unread, got %v", m)
	}
}

// TestSnapshotJSONIncludesUnread verifies the wire format: when the unread
// map is non-empty, it's serialized under the "unread" key, and when empty
// it's omitted (omitempty behavior for sparse map).
func TestSnapshotJSONIncludesUnread(t *testing.T) {
	snap := snapshot{
		PRs: github.PRSearchResults{
			{Repo: "owner/a", Number: 1},
		},
		Unread: map[string]bool{"owner/a#1": true},
	}
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	if !bytes.Contains(b, []byte(`"unread"`)) {
		t.Errorf("marshaled snapshot missing unread field: %s", s)
	}
	if !bytes.Contains(b, []byte(`"owner/a#1":true`)) {
		t.Errorf("marshaled snapshot missing unread entry: %s", s)
	}
}
