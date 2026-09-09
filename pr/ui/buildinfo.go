package ui

import "encoding/json"

// BuildInfo holds the gavel binary's build metadata, surfaced to the PR
// dashboard UI so it can display the running backend version.
type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// Build is the gavel binary's build info, set once at startup from main's
// ldflags-injected vars (see cmd/gavel). Rendered into the served page as
// window.__GAVEL__ so the UI shows the backend version without a round-trip.
var Build BuildInfo

// buildGlobalJS renders Build as a `window.__GAVEL__ = {...}` script body.
// json.Marshal HTML-escapes <, >, and & so the result is safe inside <script>.
func buildGlobalJS() string {
	b, err := json.Marshal(Build)
	if err != nil {
		return "window.__GAVEL__={};"
	}
	return "window.__GAVEL__=" + string(b) + ";"
}
