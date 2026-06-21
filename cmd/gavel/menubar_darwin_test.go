//go:build darwin && cgo

package main

import (
	"testing"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestParseMenubarExternalURL(t *testing.T) {
	origin := &application.OriginInfo{
		Origin:      "http://127.0.0.1:3020/menubar",
		IsMainFrame: true,
	}

	tests := []struct {
		name    string
		message string
		origin  *application.OriginInfo
		want    string
		wantOK  bool
	}{
		{
			name:    "http URL from menubar origin",
			message: `{"type":"gavel:open-external","url":"http://localhost:3000"}`,
			origin:  origin,
			want:    "http://localhost:3000",
			wantOK:  true,
		},
		{
			name:    "https URL from menubar origin",
			message: `{"type":"gavel:open-external","url":"https://github.com/flanksource/gavel/pull/1"}`,
			origin:  origin,
			want:    "https://github.com/flanksource/gavel/pull/1",
			wantOK:  true,
		},
		{
			name:    "wrong message type",
			message: `{"type":"other","url":"https://github.com/flanksource/gavel"}`,
			origin:  origin,
		},
		{
			name:    "wrong origin",
			message: `{"type":"gavel:open-external","url":"https://github.com/flanksource/gavel"}`,
			origin: &application.OriginInfo{
				Origin:      "http://127.0.0.1:9999/menubar",
				IsMainFrame: true,
			},
		},
		{
			name:    "subframe message",
			message: `{"type":"gavel:open-external","url":"https://github.com/flanksource/gavel"}`,
			origin: &application.OriginInfo{
				Origin:      "http://127.0.0.1:3020/menubar",
				IsMainFrame: false,
			},
		},
		{
			name:    "non-web URL scheme",
			message: `{"type":"gavel:open-external","url":"file:///tmp/nope"}`,
			origin:  origin,
		},
		{
			name:    "relative URL is rejected",
			message: `{"type":"gavel:open-external","url":"/processes"}`,
			origin:  origin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseMenubarExternalURL("http://127.0.0.1:3020", tt.message, tt.origin)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("url = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseMenubarMessage(t *testing.T) {
	origin := &application.OriginInfo{
		Origin:      "http://127.0.0.1:3020/menubar",
		IsMainFrame: true,
	}

	tests := []struct {
		name     string
		message  string
		origin   *application.OriginInfo
		wantType string
		wantOK   bool
	}{
		{
			name:     "pointer enter",
			message:  `{"type":"gavel:pointer-enter"}`,
			origin:   origin,
			wantType: menubarPointerEnterMessage,
			wantOK:   true,
		},
		{
			name:     "pointer leave",
			message:  `{"type":"gavel:pointer-leave"}`,
			origin:   origin,
			wantType: menubarPointerLeaveMessage,
			wantOK:   true,
		},
		{
			name:    "wrong origin",
			message: `{"type":"gavel:pointer-leave"}`,
			origin: &application.OriginInfo{
				Origin:      "http://127.0.0.1:9999/menubar",
				IsMainFrame: true,
			},
		},
		{
			name:    "subframe",
			message: `{"type":"gavel:pointer-leave"}`,
			origin: &application.OriginInfo{
				Origin:      "http://127.0.0.1:3020/menubar",
				IsMainFrame: false,
			},
		},
		{
			name:    "unknown type",
			message: `{"type":"gavel:nope"}`,
			origin:  origin,
		},
		{
			name:    "malformed JSON",
			message: `{nope`,
			origin:  origin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseMenubarMessage("http://127.0.0.1:3020", tt.message, tt.origin)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got.Type != tt.wantType {
				t.Fatalf("type = %q, want %q", got.Type, tt.wantType)
			}
		})
	}
}

func TestMenubarHideControllerSchedulesHide(t *testing.T) {
	hidden := make(chan struct{}, 1)
	controller := newMenubarHideController(10*time.Millisecond, func() {
		hidden <- struct{}{}
	})

	controller.schedule()
	assertSignal(t, hidden, 500*time.Millisecond)
}

func TestMenubarHideControllerCancelStopsHide(t *testing.T) {
	hidden := make(chan struct{}, 1)
	controller := newMenubarHideController(10*time.Millisecond, func() {
		hidden <- struct{}{}
	})

	controller.schedule()
	controller.cancel()
	assertNoSignal(t, hidden, 50*time.Millisecond)
}

func TestMenubarHideControllerScheduleResetsTimer(t *testing.T) {
	hidden := make(chan struct{}, 1)
	controller := newMenubarHideController(40*time.Millisecond, func() {
		hidden <- struct{}{}
	})

	controller.schedule()
	time.Sleep(20 * time.Millisecond)
	controller.schedule()

	assertNoSignal(t, hidden, 25*time.Millisecond)
	assertSignal(t, hidden, 500*time.Millisecond)
}

func assertSignal(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("expected hide signal within %s", timeout)
	}
}

func assertNoSignal(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	t.Helper()
	select {
	case <-ch:
		t.Fatalf("unexpected hide signal within %s", timeout)
	case <-time.After(timeout):
	}
}
