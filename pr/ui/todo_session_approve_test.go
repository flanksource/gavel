package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
)

func TestTodoSessionApproveFlow(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	sessionID := "approve-flow-test"

	// A driver awaits a tool-permission decision in the background.
	decided := make(chan todos.ApprovalDecision, 1)
	go func() {
		d, err := todos.GlobalApprovals().Await(context.Background(), todos.ApprovalRequest{
			SessionID: sessionID,
			Tool:      "Bash",
			Input:     map[string]any{"command": "ls"},
		})
		if err != nil {
			t.Errorf("Await: %v", err)
		}
		decided <- d
	}()

	if !waitForPending(sessionID) {
		t.Fatal("approval never became pending")
	}

	// The stats endpoint surfaces the pending approval and overrides the state.
	rec := httptest.NewRecorder()
	s.handleTodoSessionStats(rec, httptest.NewRequest(http.MethodGet,
		"/api/todos/session/stats?sessionId="+sessionID+"&dir="+workDir, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("stats status = %d; body = %q", rec.Code, rec.Body.String())
	}
	var stats struct {
		State    string `json:"state"`
		Approval *struct {
			Tool string `json:"tool"`
		} `json:"approval"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &stats); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if stats.State != "approval" {
		t.Errorf("state = %q, want approval", stats.State)
	}
	if stats.Approval == nil || stats.Approval.Tool != "Bash" {
		t.Fatalf("approval = %+v, want Bash", stats.Approval)
	}

	// Allow via the approve endpoint unblocks the driver.
	body, _ := json.Marshal(map[string]any{"sessionId": sessionID, "allow": true})
	rec2 := httptest.NewRecorder()
	s.handleTodoSessionApprove(rec2, httptest.NewRequest(http.MethodPost,
		"/api/todos/session/approve", bytes.NewReader(body)))
	if rec2.Code != http.StatusOK {
		t.Fatalf("approve status = %d; body = %q", rec2.Code, rec2.Body.String())
	}

	select {
	case d := <-decided:
		if !d.Allow {
			t.Fatalf("decision = %+v, want allow", d)
		}
	case <-time.After(time.Second):
		t.Fatal("driver Await did not return after approve")
	}
}

func TestTodoSessionApproveWithoutPending(t *testing.T) {
	s := &Server{ghOpts: github.Options{WorkDir: t.TempDir()}}
	body, _ := json.Marshal(map[string]any{"sessionId": "nobody-waiting", "allow": true})
	rec := httptest.NewRecorder()
	s.handleTodoSessionApprove(rec, httptest.NewRequest(http.MethodPost,
		"/api/todos/session/approve", bytes.NewReader(body)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("approve status = %d, want 409", rec.Code)
	}
}

func waitForPending(sessionID string) bool {
	for i := 0; i < 200; i++ {
		if _, ok := todos.GlobalApprovals().Pending(sessionID); ok {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return false
}
