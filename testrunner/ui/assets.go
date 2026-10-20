package testui

import (
	_ "embed"
	"regexp"
)

//go:embed dist/testui.js
var rawBundleJS string

// bundleJS is the bundle in a form safe to inline in a <script> element.
var bundleJS = inlineScript(rawBundleJS)

// scriptBreakers are the sequences that change where an HTML parser ends an
// inline script: `</script` closes it, and `<!--` followed by `<script`
// enters the double-escaped state, where the closing tag no longer closes it.
// The bundle carries them in string and regex literals (an HTML sanitizer, a
// syntax grammar), where \x3C reads as the same `<`.
var scriptBreakers = regexp.MustCompile(`(?i)<(!--|/?script)`)

func inlineScript(js string) string {
	return scriptBreakers.ReplaceAllString(js, `\x3C$1`)
}

// BundleJS returns the compiled testrunner UI JavaScript bundle so it can
// be embedded in HTML pages served by other packages (e.g. the PR UI),
// already safe to inline in a <script> element.
func BundleJS() string { return bundleJS }
