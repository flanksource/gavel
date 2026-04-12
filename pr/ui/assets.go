package ui

import _ "embed"

//go:embed dist/prui.js
var bundleJS string

//go:embed brand/gavel-icon.svg
var faviconSVG string

//go:embed brand/gavel-logo.svg
var logoSVG string

//go:embed brand/menubar.png
var menubarPNG []byte

//go:embed brand/menubar-unread.png
var menubarUnreadPNG []byte
