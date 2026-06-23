package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleReactGrabPluginSubstitutesOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://gavel.example:9092/react-grab-plugin.js", nil)
	rec := httptest.NewRecorder()

	handleReactGrabPlugin(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/javascript") {
		t.Fatalf("Content-Type = %q, want application/javascript", ct)
	}

	body := rec.Body.String()
	if strings.Contains(body, "__GAVEL_ORIGIN__") {
		t.Error("origin placeholder __GAVEL_ORIGIN__ was not substituted")
	}
	for _, want := range []string{
		`"http://gavel.example:9092"`, // GAVEL_ORIGIN baked to the serving origin
		"/todos/new",                  // iframe target path the plugin builds
		"gavel-todo",                  // registered action id
	} {
		if !strings.Contains(body, want) {
			t.Errorf("plugin body missing %q", want)
		}
	}
}

func TestHandleReactGrabPluginHonorsForwardedProto(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://gavel.example/react-grab-plugin.js", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	handleReactGrabPlugin(rec, req)

	if got := rec.Body.String(); !strings.Contains(got, `"https://gavel.example"`) {
		t.Errorf("forwarded https origin not applied; body did not contain https origin")
	}
}

func TestHandleReactGrabInstallSubstitutesOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://gavel.example:9092/react-grab", nil)
	rec := httptest.NewRecorder()

	handleReactGrabInstall(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}

	body := rec.Body.String()
	if strings.Contains(body, "__GAVEL_ORIGIN__") {
		t.Error("origin placeholder __GAVEL_ORIGIN__ was not substituted")
	}
	if !strings.Contains(body, "http://gavel.example:9092/react-grab-plugin.js") {
		t.Error("install page bookmarklet does not point at this server's plugin script")
	}
}
