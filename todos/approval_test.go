package todos

import (
	"context"
	"testing"
	"time"
)

func TestApprovalAwaitResolvedByAllow(t *testing.T) {
	r := &ApprovalRegistry{}
	req := ApprovalRequest{SessionID: "s1", Tool: "Bash", Input: map[string]any{"command": "ls"}}

	done := make(chan ApprovalDecision, 1)
	go func() {
		d, err := r.Await(context.Background(), req)
		if err != nil {
			t.Errorf("Await: %v", err)
		}
		done <- d
	}()

	// Wait for the request to register, then resolve it.
	if !waitFor(func() bool { _, ok := r.Pending("s1"); return ok }) {
		t.Fatal("approval never became pending")
	}
	if got, ok := r.Pending("s1"); !ok || got.Tool != "Bash" {
		t.Fatalf("Pending = %+v, %v", got, ok)
	}
	if err := r.Resolve("s1", ApprovalDecision{Allow: true, UpdatedInput: map[string]any{"command": "ls -l"}}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	select {
	case d := <-done:
		if !d.Allow || d.UpdatedInput["command"] != "ls -l" {
			t.Fatalf("decision = %+v", d)
		}
	case <-time.After(time.Second):
		t.Fatal("Await did not return after Resolve")
	}

	if _, ok := r.Pending("s1"); ok {
		t.Error("pending entry should be cleared after Await returns")
	}
}

func TestApprovalAwaitCancelled(t *testing.T) {
	r := &ApprovalRegistry{}
	ctx, cancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() {
		_, err := r.Await(ctx, ApprovalRequest{SessionID: "s2", Tool: "Bash"})
		errc <- err
	}()
	if !waitFor(func() bool { _, ok := r.Pending("s2"); return ok }) {
		t.Fatal("approval never became pending")
	}
	cancel()
	select {
	case err := <-errc:
		if err == nil {
			t.Fatal("expected a context error")
		}
	case <-time.After(time.Second):
		t.Fatal("Await did not return after cancel")
	}
}

func TestApprovalResolveWithoutPending(t *testing.T) {
	r := &ApprovalRegistry{}
	if err := r.Resolve("missing", ApprovalDecision{Allow: true}); err == nil {
		t.Fatal("Resolve should error when nothing is pending")
	}
}

func waitFor(cond func() bool) bool {
	for i := 0; i < 200; i++ {
		if cond() {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return false
}
