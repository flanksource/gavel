package ui

import (
	"strings"
	"testing"
)

func TestBuildGlobalJS(t *testing.T) {
	prev := Build
	t.Cleanup(func() { Build = prev })

	Build = BuildInfo{Version: "v9.9.9", Commit: "deadbee", Date: "2026-06-16T00:00:00Z"}
	got := buildGlobalJS()
	want := `window.__GAVEL__={"version":"v9.9.9","commit":"deadbee","date":"2026-06-16T00:00:00Z"};`
	if got != want {
		t.Errorf("buildGlobalJS() = %q, want %q", got, want)
	}
}

func TestPageHTMLInjectsBuildGlobal(t *testing.T) {
	prev := Build
	t.Cleanup(func() { Build = prev })

	Build = BuildInfo{Version: "v1.2.3", Commit: "abc1234", Date: "2026-01-02T03:04:05Z"}
	html := pageHTML()

	want := `window.__GAVEL__={"version":"v1.2.3","commit":"abc1234","date":"2026-01-02T03:04:05Z"};`
	if !strings.Contains(html, want) {
		t.Errorf("pageHTML missing backend build global; want substring %s", want)
	}
	// The global must be defined before the app bundle runs, i.e. after the
	// #root mount point and somewhere in the body.
	if strings.Index(html, "window.__GAVEL__=") < strings.Index(html, `id="root"`) {
		t.Error("build global should appear after the #root element, before the app script")
	}
}
