package parsers

import "regexp"

var ansiSequenceRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes ANSI SGR/CSI escape sequences from s. Jest, Vitest, and
// Playwright all embed ANSI color codes in their JSON failure messages; we
// strip them before storing Message so downstream renderers don't double-color.
func StripANSI(s string) string {
	return ansiSequenceRe.ReplaceAllString(s, "")
}
