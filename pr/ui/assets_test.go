package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
)

// The PWA home-screen assets (manifest + icons) must be served by the Go mux as
// concrete routes so they resolve in both prod and `--dev` mode (where only "/"
// is proxied to Vite).
func TestHandleHomeScreenAssets(t *testing.T) {
	s := NewServer(0, github.Options{}, SearchConfig{})
	cases := []struct {
		path     string
		wantType string
		contains string
	}{
		{"/manifest.webmanifest", "application/manifest+json", "/brand/icon-512.png"},
		{"/brand/apple-touch-icon.png", "image/png", ""},
		{"/brand/icon-192.png", "image/png", ""},
		{"/brand/icon-512.png", "image/png", ""},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status: got %d want 200", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, tc.wantType) {
				t.Fatalf("content-type: got %q want %q", ct, tc.wantType)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("body is empty")
			}
			if tc.contains != "" && !strings.Contains(rec.Body.String(), tc.contains) {
				t.Fatalf("body missing %q", tc.contains)
			}
		})
	}
}

// The served React Grab plugin must capture up to 2KB of the grabbed element's
// raw outerHTML into the todo body. Guards against the capture being dropped.
func TestReactGrabPluginCapturesHTML(t *testing.T) {
	s := NewServer(0, github.Options{}, SearchConfig{})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/react-grab-plugin.js", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"outerHTML", "2048", "```html"} {
		if !strings.Contains(body, want) {
			t.Errorf("plugin JS missing %q", want)
		}
	}
}

// The plugin must offer screenshot capture: a getDisplayMedia stream cropped to
// the grabbed element via Region Capture (CropTarget/cropTo), shipped to the todo
// form. Guards against the capture path being dropped.
func TestReactGrabPluginCapturesScreenshot(t *testing.T) {
	s := NewServer(0, github.Options{}, SearchConfig{})
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/react-grab-plugin.js", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"getDisplayMedia", "cropTo", "gavel-screenshot", "embed-ready"} {
		if !strings.Contains(body, want) {
			t.Errorf("plugin JS missing %q", want)
		}
	}
}

// The served SPA shell must advertise the home-screen metadata so iOS/Android
// offer "Add to Home Screen" with the gavel icon and standalone chrome.
func TestPageHTMLHomeScreenTags(t *testing.T) {
	html := pageHTML()
	for _, want := range []string{
		`rel="apple-touch-icon" href="/brand/apple-touch-icon.png"`,
		`rel="manifest" href="/manifest.webmanifest"`,
		`name="apple-mobile-web-app-title" content="gavel"`,
		`name="apple-mobile-web-app-capable" content="yes"`,
		`name="theme-color" content="#3578e5"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("pageHTML missing %q", want)
		}
	}
}
