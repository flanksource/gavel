package ui

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleTodoSessionPlanRequiresNativeReference(t *testing.T) {
	s := &Server{}
	recorder := httptest.NewRecorder()
	s.handleTodoSessionPlan(recorder, httptest.NewRequest(http.MethodGet, "/api/todos/session/plan?sessionId=legacy", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleTodoSessionPlanSaveValidatesRevisionPayload(t *testing.T) {
	s := &Server{}
	tests := []struct {
		name string
		body string
	}{
		{name: "missing ref", body: `{"content":"# plan"}`},
		{name: "missing content", body: `{"ref":"issue-id"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			s.handleTodoSessionPlanSave(recorder, httptest.NewRequest(
				http.MethodPost, "/api/todos/session/plan", bytes.NewBufferString(test.body),
			))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
