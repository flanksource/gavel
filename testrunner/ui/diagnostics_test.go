package testui_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestSnapshotIncludesDiagnosticsAvailability(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetDiagnosticsManager(testui.NewDiagnosticsManager(4242, newFakeDiagnosticsCollector()))

	var snap struct {
		DiagnosticsAvailable bool `json:"diagnostics_available"`
	}
	resp := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !snap.DiagnosticsAvailable {
		t.Fatalf("diagnostics_available = false, want true")
	}
}

func TestDiagnosticsEndpointReturnsProcessTree(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetDiagnosticsManager(testui.NewDiagnosticsManager(4242, newFakeDiagnosticsCollector()))

	var snap testui.DiagnosticsSnapshot
	resp := doRequest(t, handler, http.MethodGet, "/api/diagnostics", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.Code)
	}
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if snap.Root == nil || snap.Root.PID != 4242 {
		t.Fatalf("root = %#v, want pid 4242", snap.Root)
	}
	if len(snap.Root.Children) != 1 || snap.Root.Children[0].PID != 5151 {
		t.Fatalf("children = %#v, want child pid 5151", snap.Root.Children)
	}
}

func TestDiagnosticsCollectOverwritesLatestCapture(t *testing.T) {
	srv, handler := newTestServer(t)
	collector := newFakeDiagnosticsCollector()
	manager := testui.NewDiagnosticsManager(4242, collector)
	srv.SetDiagnosticsManager(manager)

	body, _ := json.Marshal(testui.StackCaptureRequest{PID: 4242})
	resp := doRequest(t, handler, http.MethodPost, "/api/diagnostics/collect", bytes.NewReader(body))
	if resp.Code != http.StatusOK {
		t.Fatalf("first collect status = %d, want 200", resp.Code)
	}

	var details testui.ProcessDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		t.Fatalf("decode first collect: %v", err)
	}
	if details.StackCapture == nil || !strings.Contains(details.StackCapture.Text, "goroutine 1 [running]:") {
		t.Fatalf("first stack capture = %#v", details.StackCapture)
	}

	collector.captureByPID[4242] = testui.StackCapture{
		Status:    "ready",
		Supported: true,
		Text:      "goroutine 2 [running]:\nmain.second()",
	}
	resp = doRequest(t, handler, http.MethodPost, "/api/diagnostics/collect", bytes.NewReader(body))
	if resp.Code != http.StatusOK {
		t.Fatalf("second collect status = %d, want 200", resp.Code)
	}

	var snap testui.DiagnosticsSnapshot
	resp = doRequest(t, handler, http.MethodGet, "/api/diagnostics", nil)
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode diagnostics: %v", err)
	}
	if snap.Root == nil || snap.Root.StackCapture == nil {
		t.Fatalf("root stack capture missing: %#v", snap.Root)
	}
	if strings.Contains(snap.Root.StackCapture.Text, "goroutine 1 [running]:") {
		t.Fatalf("expected latest-only capture, got stale stack %q", snap.Root.StackCapture.Text)
	}
	if !strings.Contains(snap.Root.StackCapture.Text, "main.second()") {
		t.Fatalf("latest stack capture = %q, want main.second()", snap.Root.StackCapture.Text)
	}
}

func TestDiagnosticsCollectChildReturnsUnsupported(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetDiagnosticsManager(testui.NewDiagnosticsManager(4242, newFakeDiagnosticsCollector()))

	body, _ := json.Marshal(testui.StackCaptureRequest{PID: 5151})
	resp := doRequest(t, handler, http.MethodPost, "/api/diagnostics/collect", bytes.NewReader(body))
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.Code)
	}

	var details testui.ProcessDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if details.StackCapture == nil || details.StackCapture.Status != "unsupported" {
		t.Fatalf("stack capture = %#v, want unsupported", details.StackCapture)
	}
}

func TestDiagnosticsRoutesServeHTMLAndRejectExport(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetDiagnosticsManager(testui.NewDiagnosticsManager(4242, newFakeDiagnosticsCollector()))

	resp := doHTMLRequest(t, handler, http.MethodGet, "/diagnostics/4242")
	if resp.Code != http.StatusOK {
		t.Fatalf("html status = %d, want 200", resp.Code)
	}
	if got := resp.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("content-type = %q, want html", got)
	}

	resp = doRequest(t, handler, http.MethodGet, "/diagnostics.json", nil)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("export status = %d, want 404", resp.Code)
	}
}

type fakeDiagnosticsCollector struct {
	snapshot     *testui.DiagnosticsSnapshot
	captureByPID map[int]testui.StackCapture
}

func newFakeDiagnosticsCollector() *fakeDiagnosticsCollector {
	openRoot := int32(14)
	openChild := int32(5)
	return &fakeDiagnosticsCollector{
		snapshot: &testui.DiagnosticsSnapshot{
			Root: &testui.ProcessNode{
				PID:        4242,
				Name:       "gavel",
				Command:    "gavel test --ui ./testrunner/ui",
				Status:     "running",
				CPUPercent: 11.2,
				RSS:        64 << 20,
				VMS:        320 << 20,
				OpenFiles:  &openRoot,
				IsRoot:     true,
				Children: []*testui.ProcessNode{{
					PID:        5151,
					PPID:       4242,
					Name:       "go",
					Command:    "go test ./testrunner/ui",
					Status:     "sleep",
					CPUPercent: 3.4,
					RSS:        22 << 20,
					VMS:        140 << 20,
					OpenFiles:  &openChild,
				}},
			},
		},
		captureByPID: map[int]testui.StackCapture{
			4242: {
				Status:    "ready",
				Supported: true,
				Text:      "goroutine 1 [running]:\nmain.first()",
			},
			5151: {
				Status:    "unsupported",
				Supported: false,
				Error:     "stack capture is only supported for the live gavel process",
			},
		},
	}
}

func (f *fakeDiagnosticsCollector) Snapshot(_ int) (*testui.DiagnosticsSnapshot, error) {
	data, err := json.Marshal(f.snapshot)
	if err != nil {
		return nil, err
	}
	var cloned testui.DiagnosticsSnapshot
	if err := json.Unmarshal(data, &cloned); err != nil {
		return nil, err
	}
	return &cloned, nil
}

func (f *fakeDiagnosticsCollector) CaptureStack(_ int, pid int) testui.StackCapture {
	return f.captureByPID[pid]
}
