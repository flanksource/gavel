package ui

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/github"
)

// runSSE drives handleSSE until it has written at least wantFrames frames (a
// frame being either a data: payload or a : ping comment), then cancels and
// waits for the handler to return so the recorder can be read without racing.
func runSSE(t *testing.T, s *Server, wantFrames int) string {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/prs/stream", nil).WithContext(ctx)

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.handleSSE(rec, req)
	}()

	// Nudge the handler once per expected frame; each notify wakes one loop
	// iteration immediately rather than waiting out the 2s liveness ticker.
	deadline := time.After(5 * time.Second)
	for i := 0; i < wantFrames; i++ {
		select {
		case s.updated <- struct{}{}:
		case <-deadline:
			t.Fatal("handleSSE stopped consuming updates")
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handleSSE did not return after context cancellation")
	}
	return rec.Body.String()
}

// An unchanged snapshot must not be re-pushed. handleSSE's ticker is a liveness
// cadence, not a change signal: before the dedupe it re-marshalled and re-sent
// the entire PR snapshot every 2s forever, and each frame minted a fresh object
// on the client that re-rendered the whole app — including on routes that show
// no PR data at all.
func TestPRStreamPingsWhenSnapshotUnchanged(t *testing.T) {
	s := &Server{
		updated:   make(chan struct{}),
		fetchedAt: time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC),
		prs:       github.PRSearchResults{{Number: 1, Title: "first", Repo: "flanksource/gavel"}},
	}

	body := runSSE(t, s, 3)

	if got := strings.Count(body, "data: "); got != 1 {
		t.Errorf("want exactly 1 data frame (the initial snapshot), got %d\nbody=%q", got, body)
	}
	if !strings.Contains(body, ": ping") {
		t.Errorf("want keep-alive ping frames for the unchanged snapshot, got body=%q", body)
	}
}

// A real change must still reach the client, or the dedupe would be a
// correctness bug rather than an optimisation.
func TestPRStreamPushesWhenSnapshotChanges(t *testing.T) {
	s := &Server{
		updated:   make(chan struct{}),
		fetchedAt: time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC),
		prs:       github.PRSearchResults{{Number: 1, Title: "first", Repo: "flanksource/gavel"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/prs/stream", nil).WithContext(ctx)

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.handleSSE(rec, req)
	}()

	// First wake: nothing moved, expect a ping. Then mutate the snapshot and
	// wake again, which must produce a second data frame.
	s.updated <- struct{}{}
	time.Sleep(10 * time.Millisecond)

	s.mu.Lock()
	s.prs = github.PRSearchResults{
		{Number: 1, Title: "first", Repo: "flanksource/gavel"},
		{Number: 2, Title: "second", Repo: "flanksource/gavel"},
	}
	s.mu.Unlock()

	s.updated <- struct{}{}
	time.Sleep(10 * time.Millisecond)

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handleSSE did not return after context cancellation")
	}

	body := rec.Body.String()
	if got := strings.Count(body, "data: "); got != 2 {
		t.Errorf("want 2 data frames (initial + the change), got %d\nbody=%q", got, body)
	}
	if !strings.Contains(body, "second") {
		t.Errorf("changed snapshot never reached the client, body=%q", body)
	}
}
