package ui

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFaviconHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/favicon.svg", nil)
	rec := httptest.NewRecorder()
	handleFavicon(rec, req)

	if got := rec.Code; got != 200 {
		t.Errorf("status = %d, want 200", got)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Errorf("Content-Type = %q, want image/svg+xml", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<svg") {
		t.Errorf("body missing <svg tag")
	}
	if !strings.Contains(body, "#3578e5") {
		t.Errorf("favicon missing Flanksource Primary Blue (#3578e5)")
	}
	if !strings.Contains(body, "#0ea5e9") {
		t.Errorf("favicon missing Sky Blue (#0ea5e9)")
	}
}

func TestLogoHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/brand/gavel-logo.svg", nil)
	rec := httptest.NewRecorder()
	handleLogo(rec, req)

	if got := rec.Code; got != 200 {
		t.Errorf("status = %d, want 200", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "gavel") {
		t.Errorf("logo missing gavel wordmark text")
	}
	if !strings.Contains(body, "Fira Code") {
		t.Errorf("logo missing Fira Code font reference")
	}
}

func TestMenubarIconHandler(t *testing.T) {
	cases := []struct {
		name    string
		path    string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{"normal", "/brand/menubar.png", handleMenubarIcon},
		{"unread", "/brand/menubar-unread.png", handleMenubarUnreadIcon},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			rec := httptest.NewRecorder()
			tc.handler(rec, req)

			if got := rec.Code; got != 200 {
				t.Errorf("status = %d, want 200", got)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
				t.Errorf("Content-Type = %q, want image/png", ct)
			}
			img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
			if err != nil {
				t.Fatalf("response body is not a valid PNG: %v", err)
			}
			if h := img.Bounds().Dy(); h != 44 {
				t.Errorf("menubar icon height = %d, want 44 (2x retina for 22px menubar target)", h)
			}
		})
	}
}

// TestMenubarIconsDiffer guards against accidentally publishing the same PNG
// bytes for both variants — the whole point of the second file is that it
// has a visible difference (the unread dot).
func TestMenubarIconsDiffer(t *testing.T) {
	if bytes.Equal(menubarPNG, menubarUnreadPNG) {
		t.Error("normal and unread menubar PNGs are byte-identical; the unread variant is missing the indicator dot")
	}
}

func TestPageHTMLReferencesFavicon(t *testing.T) {
	html := pageHTML()
	if !strings.Contains(html, `href="/favicon.svg"`) {
		t.Errorf("pageHTML missing favicon link")
	}
	if !strings.Contains(html, `gavel`) {
		t.Errorf("pageHTML missing gavel in title")
	}
}
