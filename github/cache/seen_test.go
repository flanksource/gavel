package cache

import "testing"

// TestSeenDisabledNoOp verifies that all seen-tracking methods become silent
// no-ops when the cache is disabled. This is the path every CLI invocation
// takes when GAVEL_GITHUB_CACHE_DSN is unset, so it must not panic or error.
func TestSeenDisabledNoOp(t *testing.T) {
	s := &Store{disabled: true}

	if err := s.MarkSeen("owner/repo", 42); err != nil {
		t.Errorf("MarkSeen on disabled store: %v", err)
	}

	if err := s.MarkManySeen([]SeenKey{{Repo: "owner/repo", Number: 42}}); err != nil {
		t.Errorf("MarkManySeen on disabled store: %v", err)
	}

	m, err := s.LoadSeenMap([]SeenKey{{Repo: "owner/repo", Number: 42}})
	if err != nil {
		t.Errorf("LoadSeenMap on disabled store: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("LoadSeenMap on disabled store returned %d entries, want 0", len(m))
	}
}

// TestSeenDisabledEmpty guards the zero-argument paths. An empty key slice
// must not issue a database query, even on an enabled store, so handlers
// can safely call LoadSeenMap before checking len(prs).
func TestSeenDisabledEmpty(t *testing.T) {
	s := &Store{disabled: true}
	if _, err := s.LoadSeenMap(nil); err != nil {
		t.Errorf("LoadSeenMap(nil): %v", err)
	}
	if err := s.MarkManySeen(nil); err != nil {
		t.Errorf("MarkManySeen(nil): %v", err)
	}
}
