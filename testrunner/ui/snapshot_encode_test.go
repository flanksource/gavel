package testui

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func unencodableServer() *Server {
	srv := NewServer()
	srv.SetResults([]parsers.Test{{Name: "row", Context: map[string]any{"value": math.NaN()}}})
	return srv
}

func TestHandleJSONReportsSnapshotEncodeError(t *testing.T) {
	recorder := httptest.NewRecorder()
	unencodableServer().Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/tests", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (body %q)", recorder.Code, recorder.Body.String())
	}
	if body := recorder.Body.String(); !strings.Contains(body, "encode test snapshot") || !strings.Contains(body, "NaN") {
		t.Fatalf("body = %q, want the encoder's NaN error", body)
	}
}

func TestHandleSSEReportsSnapshotEncodeError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	recorder := httptest.NewRecorder()
	unencodableServer().Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/tests/stream", nil).WithContext(ctx))

	if body := recorder.Body.String(); !strings.Contains(body, "event: error") || !strings.Contains(body, "NaN") {
		t.Fatalf("body = %q, want an error event naming the NaN", body)
	}
}
