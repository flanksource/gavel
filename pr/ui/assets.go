package ui

import _ "embed"

//go:embed dist/prui.js
var bundleJS string

//go:embed dist/prui.css
var bundleCSS string

//go:embed brand/gavel-icon.svg
var faviconSVG string

//go:embed brand/gavel-logo.svg
var logoSVG string

//go:embed brand/menubar.png
var menubarPNG []byte

//go:embed brand/menubar-unread.png
var menubarUnreadPNG []byte

//go:embed brand/manifest.webmanifest
var webManifest string

//go:embed brand/apple-touch-icon.png
var appleTouchIconPNG []byte

//go:embed brand/icon-192.png
var icon192PNG []byte

//go:embed brand/icon-512.png
var icon512PNG []byte

//go:embed assets/react-grab-plugin.js
var reactGrabPluginJS string

//go:embed assets/react-grab-install.html
var reactGrabInstallHTML string

// MenubarIconPNG returns the embedded macOS menubar icon bytes.
func MenubarIconPNG() []byte {
	return menubarPNG
}
