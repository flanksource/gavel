package testui

import _ "embed"

//go:embed dist/testui.js
var bundleJS string

// BundleJS returns the compiled testrunner UI JavaScript bundle so it can
// be embedded in HTML pages served by other packages (e.g. the PR UI).
func BundleJS() string { return bundleJS }
